"""秒级定时调度模块。

在后台提供基于微秒对齐的高精度的秒级循环调度，支持去重以及触发抢零点模式。
"""

import asyncio
import time
from datetime import datetime
from typing import List, Set

from app.config import get_app_now
from app.models import HistoryItem, Task
from app.runner import execute_task
from app.storage import StateManager

# 已运行的定时任务去重 Key 集合
_schedule_ran: Set[str] = set()


def should_trigger_at(task: Task, now: datetime, hour: int, minute: int, second: int) -> bool:
    """判断当前时间点是否符合触发要求 (3 秒触发窗口)。"""
    if now.hour != hour or now.minute != minute:
        return False

    current_second: int = now.second
    return current_second >= second and current_second < (second + 3)


def scheduled_run_key(task_id: int, now: datetime, hour: int, minute: int, second: int) -> str:
    """生成唯一定时任务运行标记 Key。"""
    return f"{task_id}:{now.year:04d}-{now.month:02d}-{now.day:02d} {hour:02d}:{minute:02d}:{second:02d}"


async def run_scheduled_tasks(
    now: datetime, triggered_by: str, state_manager: StateManager
) -> List[HistoryItem]:
    """检查并触发所有到达定时时间的任务。"""
    tasks: List[Task] = list(state_manager.state.tasks)
    schedule = state_manager.state.settings.schedule
    due_tasks: List[Task] = []

    for task in tasks:
        if not task.enabled:
            continue

        hour, minute, second = schedule.hour, schedule.minute, schedule.second
        should_run: bool = schedule.enabled

        if task.schedule_enabled:
            hour, minute, second = task.schedule_hour, task.schedule_minute, task.schedule_second
            should_run = True

        if should_run and should_trigger_at(task, now, hour, minute, second):
            key = scheduled_run_key(task.id, now, hour, minute, second)
            if key not in _schedule_ran:
                _schedule_ran.add(key)
                due_tasks.append(task)

    if not due_tasks:
        return []

    results: List[HistoryItem] = []
    for task in due_tasks:
        res = await execute_task(task, triggered_by, state_manager)
        results.append(res)

    return results


async def scheduler_loop(state_manager: StateManager) -> None:
    """后台运行的秒级循环调度器。

    精确对齐到系统秒级边界，确保准确在设定的秒数触发。
    """
    while True:
        try:
            now: datetime = get_app_now()
            await run_scheduled_tasks(now, "schedule", state_manager)

            # 对齐下一次整秒时间点
            sleep_duration: float = 1.0 - (time.time() % 1.0)
            await asyncio.sleep(sleep_duration)
        except asyncio.CancelledError:
            break
        except Exception:
            await asyncio.sleep(1)
