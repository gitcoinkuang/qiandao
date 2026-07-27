"""HTTP 任务执行引擎与评估逻辑模块。

处理任务的 HTTP 请求发起、响应结果关键词评估、并发控制、抢零点/Burst重试模式以及历史记录保存。
"""

import asyncio
import time
from typing import List

import httpx

from app.config import format_now, format_now_ms
from app.models import HistoryItem, Task
from app.notifier import send_notifications
from app.storage import StateManager

# 高性能连接池配置
_limits = httpx.Limits(max_keepalive_connections=128, max_connections=64)
_client = httpx.AsyncClient(limits=_limits, follow_redirects=True)

SCHEDULED_BURST_RETRIES: int = 3
SCHEDULED_BURST_INTERVAL: float = 0.250  # 250 毫秒


def evaluate_response(task: Task, status_code: int, body: str) -> str:
    """评估 HTTP 响应状态码及返回体关键词。

    根据 HTTP Status Code 和任务配置的成功/失败关键词判断请求判定结果。
    """
    if status_code < 200 or status_code >= 300:
        return f"http {status_code}"

    body_lower = body.lower()

    # 校验失败关键词
    if task.failure_keywords:
        fail_words = [w.strip().lower() for w in task.failure_keywords.split(",") if w.strip()]
        for word in fail_words:
            if word in body_lower:
                return "failure keyword matched"

    # 校验成功关键词
    if task.success_keywords:
        succ_words = [w.strip().lower() for w in task.success_keywords.split(",") if w.strip()]
        if succ_words:
            matched = any(word in body_lower for word in succ_words)
            if not matched:
                return "success keyword missing"

    return "success"


def should_use_burst_mode(task: Task, triggered_by: str) -> bool:
    """判断是否应当开启抢零点 Burst 重试模式。"""
    if not task.aggressive_mode:
        return False
    return triggered_by in ("schedule", "manual-schedule-check")


async def execute_task(task: Task, triggered_by: str, state_manager: StateManager) -> HistoryItem:
    """执行单个签到任务。

    包含最大重试、抢零点模式逻辑、结果评估与历史保存。
    """
    max_attempts: int = task.retry_count + 1
    burst_mode: bool = should_use_burst_mode(task, triggered_by)

    if burst_mode and max_attempts < SCHEDULED_BURST_RETRIES + 1:
        max_attempts = SCHEDULED_BURST_RETRIES + 1

    last_message: str = ""
    last_status_code: int = 0
    preview: str = ""
    duration_ms: int = 0
    last_request_started_at: str = ""

    for attempt in range(max_attempts):
        started_time: float = time.time()
        request_started_at: str = format_now_ms()
        last_request_started_at = request_started_at

        try:
            req_headers = dict(task.headers)
            body_bytes = task.body.encode("utf-8") if task.body else None
            response = await _client.request(
                method=task.method,
                url=task.url,
                headers=req_headers,
                content=body_bytes,
                timeout=float(task.timeout_seconds),
            )

            duration_ms = int((time.time() - started_time) * 1000)
            last_status_code = response.status_code

            # 截取前 4096 字节作为预览
            preview_bytes = response.content[:4096]
            preview = preview_bytes.decode("utf-8", errors="replace")

            last_message = evaluate_response(task, response.status_code, preview)

            if last_message == "success":
                result = HistoryItem(
                    task_id=task.id,
                    task_name=task.name,
                    status="success",
                    status_code=response.status_code,
                    message="request completed",
                    response_preview=preview,
                    response_time_ms=duration_ms,
                    request_started_at=request_started_at,
                    triggered_by=triggered_by,
                    created_at=format_now(),
                )
                await finish_run(task.id, result, state_manager)
                return result

        except Exception as exc:
            duration_ms = int((time.time() - started_time) * 1000)
            last_message = str(exc)

        # 失败重试等待
        if attempt < max_attempts - 1:
            if burst_mode and attempt < SCHEDULED_BURST_RETRIES:
                await asyncio.sleep(SCHEDULED_BURST_INTERVAL)
            else:
                await asyncio.sleep(attempt + 1)

    result = HistoryItem(
        task_id=task.id,
        task_name=task.name,
        status="failed",
        status_code=last_status_code,
        message=last_message,
        response_preview=preview,
        response_time_ms=duration_ms,
        request_started_at=last_request_started_at,
        triggered_by=triggered_by,
        created_at=format_now(),
    )
    await finish_run(task.id, result, state_manager)
    return result


async def finish_run(task_id: int, item: HistoryItem, state_manager: StateManager) -> None:
    """完成任务运行，记录历史日志并写回磁盘，最后异步发送通知。"""
    async with state_manager._lock:
        item.id = state_manager.state.next_history_id
        state_manager.state.next_history_id += 1
        state_manager.state.history.insert(0, item)

        # 限制历史记录最多 200 条
        if len(state_manager.state.history) > 200:
            state_manager.state.history = state_manager.state.history[:200]

        for t in state_manager.state.tasks:
            if t.id == task_id:
                t.last_status = item.status
                t.last_run_at = item.created_at
                t.last_duration_ms = item.response_time_ms
                t.updated_at = format_now()
                break

        await state_manager._save_unlocked()

    # 异步推送通知
    asyncio.create_task(send_notifications(item, state_manager.state.settings.notify))


async def run_all_enabled_tasks(
    triggered_by: str, state_manager: StateManager
) -> List[HistoryItem]:
    """并发批量运行所有启用的任务。"""
    tasks = [t for t in state_manager.state.tasks if t.enabled]
    if not tasks:
        return []

    max_workers = state_manager.state.settings.schedule.max_workers
    if max_workers < 1:
        max_workers = 4
    elif max_workers > 8:
        max_workers = 8

    sem = asyncio.Semaphore(max_workers)

    async def worker(task: Task) -> HistoryItem:
        async with sem:
            return await execute_task(task, triggered_by, state_manager)

    results = await asyncio.gather(*(worker(t) for t in tasks))
    # 按 task_id 倒序排列
    sorted_results = sorted(list(results), key=lambda x: x.task_id, reverse=True)
    return sorted_results
