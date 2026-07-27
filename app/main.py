"""FastAPI 主程序与路由入口。

包含路由注册、静态文件挂载、模板渲染以及生命周期 (lifespan) 管理。
"""

import asyncio
from contextlib import asynccontextmanager
from typing import Any, AsyncGenerator, Dict, List, Optional

import uvicorn
from fastapi import Depends, FastAPI, Form, HTTPException, Request, Response, status
from fastapi.responses import HTMLResponse, JSONResponse, RedirectResponse
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates

from app.auth import (
    check_password,
    create_session,
    delete_session,
    generate_token,
    hash_password,
    is_authenticated,
)
from app.config import format_now, get_app_now
from app.curl_parser import normalize_task, parse_curl
from app.models import (
    APIResponse,
    HistoryItem,
    NotifySettings,
    ScheduleSettings,
    Task,
)
from app.notifier import send_test_telegram_notification
from app.runner import execute_task, run_all_enabled_tasks
from app.scheduler import run_scheduled_tasks, scheduler_loop
from app.storage import StateManager

state_manager = StateManager()
templates = Jinja2Templates(directory="templates")


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """FastAPI 生命周期管理器：启动时加载存储状态并启动定时器任务。"""
    await state_manager.load()
    scheduler_task = asyncio.create_task(scheduler_loop(state_manager))
    yield
    scheduler_task.cancel()
    try:
        await scheduler_task
    except asyncio.CancelledError:
        pass


app = FastAPI(title="QianDao V2", lifespan=lifespan)

# 挂载静态文件
app.mount("/static", StaticFiles(directory="static"), name="static")


# API 认证依赖项
async def require_api_auth(request: Request) -> None:
    """API 请求认证依赖。未登录时抛出 401 HTTPException。"""
    security_enabled = state_manager.state.settings.security.enabled
    if security_enabled and not is_authenticated(request, security_enabled):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="unauthorized",
        )


# 自定义 HTTPException 处理器，返回符合要求的 JSON 数据格式
@app.exception_handler(HTTPException)
async def custom_http_exception_handler(
    request: Request, exc: HTTPException
) -> JSONResponse:
    return JSONResponse(
        status_code=exc.status_code,
        content={"success": False, "error": exc.detail},
    )


# ---------------- Web 视图路由 ----------------


@app.get("/", response_class=HTMLResponse)
async def handle_index(request: Request) -> Any:
    """首页总览控制台页面。"""
    security_enabled = state_manager.state.settings.security.enabled
    if security_enabled and not is_authenticated(request, security_enabled):
        return RedirectResponse(url="/login", status_code=status.HTTP_303_SEE_OTHER)
    return templates.TemplateResponse(
        request=request, name="index.html", context={"Title": "QianDao V2"}
    )


@app.get("/login", response_class=HTMLResponse)
async def handle_login_page(request: Request) -> Any:
    """登录页面。"""
    security_enabled = state_manager.state.settings.security.enabled
    if not security_enabled:
        return RedirectResponse(url="/", status_code=status.HTTP_303_SEE_OTHER)
    return templates.TemplateResponse(
        request=request, name="login.html", context={"Error": ""}
    )


@app.post("/login", response_class=HTMLResponse)
async def handle_login_submit(request: Request, password: str = Form("")) -> Any:
    """提交登录表单。"""
    security_enabled = state_manager.state.settings.security.enabled
    if not security_enabled:
        return RedirectResponse(url="/", status_code=status.HTTP_303_SEE_OTHER)

    stored_hash = state_manager.state.settings.security.password_hash
    if not check_password(password.strip(), stored_hash, security_enabled):
        return templates.TemplateResponse(
            request=request, name="login.html", context={"Error": "密码错误"}
        )

    token = generate_token(32)
    create_session(token)

    response = RedirectResponse(url="/", status_code=status.HTTP_303_SEE_OTHER)
    response.set_cookie(
        key="qiandao_session",
        value=token,
        path="/",
        httponly=True,
        samesite="lax",
    )
    return response


@app.get("/logout")
async def handle_logout(request: Request) -> Any:
    """退出登录并清除 Cookie。"""
    cookie_val = request.cookies.get("qiandao_session")
    if cookie_val:
        delete_session(cookie_val)

    response = RedirectResponse(url="/login", status_code=status.HTTP_303_SEE_OTHER)
    response.delete_cookie(key="qiandao_session", path="/", httponly=True)
    return response


# ---------------- REST API 路由 ----------------


@app.get("/api/bootstrap", dependencies=[Depends(require_api_auth)])
async def handle_bootstrap() -> Dict[str, Any]:
    """系统初始化数据引导接口。"""
    async with state_manager._lock:
        history: List[HistoryItem] = list(state_manager.state.history)
        tasks: List[Task] = list(state_manager.state.tasks)

        success_count = sum(1 for item in history if item.status == "success")
        failed_count = sum(1 for item in history if item.status == "failed")
        duration_items = [
            item.response_time_ms for item in history if item.response_time_ms > 0
        ]
        avg_duration = (
            sum(duration_items) // len(duration_items) if duration_items else 0
        )
        enabled_count = sum(1 for t in tasks if t.enabled)

        data = {
            "stats": {
                "total_tasks": len(tasks),
                "enabled_tasks": enabled_count,
                "recent_success": success_count,
                "recent_failed": failed_count,
                "avg_duration_ms": avg_duration,
            },
            "tasks": [t.model_dump() for t in tasks],
            "history": [h.model_dump() for h in history],
            "notify_config": state_manager.state.settings.notify.model_dump(),
            "schedule_config": state_manager.state.settings.schedule.model_dump(),
            "security_config": {
                "enabled": state_manager.state.settings.security.enabled
            },
        }

    return {"success": True, "data": data}


@app.get("/api/tasks", dependencies=[Depends(require_api_auth)])
async def handle_get_tasks() -> Dict[str, Any]:
    """获取所有任务列表。"""
    async with state_manager._lock:
        tasks = [t.model_dump() for t in state_manager.state.tasks]
    return {"success": True, "tasks": tasks}


@app.post("/api/tasks", dependencies=[Depends(require_api_auth)])
async def handle_create_task(payload: Task) -> JSONResponse:
    """新建任务。"""
    try:
        task = normalize_task(payload)
    except ValueError as exc:
        return JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            content={"success": False, "error": str(exc)},
        )

    async with state_manager._lock:
        task.id = state_manager.state.next_task_id
        state_manager.state.next_task_id += 1
        now_str = format_now()
        task.created_at = now_str
        task.updated_at = now_str
        task.last_status = "idle"
        state_manager.state.tasks.insert(0, task)
        await state_manager._save_unlocked()

    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={"success": True, "task": task.model_dump()},
    )


@app.post("/api/tasks/parse", dependencies=[Depends(require_api_auth)])
async def handle_parse_task(payload: Task) -> JSONResponse:
    """解析 Curl 命令提取任务参数。"""
    try:
        url, method, headers, body = parse_curl(payload.curl_command)
        parsed = Task(
            name=payload.name if payload.name else "",
            url=url,
            method=payload.method.upper() if payload.method else method,
            headers=headers,
            body=body,
            curl_command=payload.curl_command,
        )
        return JSONResponse(
            status_code=status.HTTP_200_OK,
            content={"success": True, "config": parsed.model_dump()},
        )
    except Exception as exc:
        return JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            content={"success": False, "error": str(exc)},
        )


@app.post("/api/tasks/run-all", dependencies=[Depends(require_api_auth)])
async def handle_run_all_tasks() -> Dict[str, Any]:
    """批量触发运行所有已启用的任务。"""
    results = await run_all_enabled_tasks("manual-batch", state_manager)
    return {"success": True, "results": [r.model_dump() for r in results]}


@app.post("/api/tasks/{id}/run", dependencies=[Depends(require_api_auth)])
async def handle_run_task_by_id(id: int) -> JSONResponse:
    """手动触发运行单个任务。"""
    target_task: Optional[Task] = None
    async with state_manager._lock:
        for t in state_manager.state.tasks:
            if t.id == id:
                target_task = t
                break

    if not target_task:
        return JSONResponse(
            status_code=status.HTTP_404_NOT_FOUND,
            content={"success": False, "error": "task not found"},
        )

    result = await execute_task(target_task, "manual", state_manager)
    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={"success": True, "result": result.model_dump()},
    )


@app.put("/api/tasks/{id}", dependencies=[Depends(require_api_auth)])
async def handle_update_task(id: int, payload: Task) -> JSONResponse:
    """更新任务配置。"""
    try:
        updated = normalize_task(payload)
    except ValueError as exc:
        return JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            content={"success": False, "error": str(exc)},
        )

    found = False
    async with state_manager._lock:
        for i, t in enumerate(state_manager.state.tasks):
            if t.id == id:
                updated.id = id
                updated.created_at = t.created_at
                updated.last_status = t.last_status
                updated.last_run_at = t.last_run_at
                updated.last_duration_ms = t.last_duration_ms
                updated.updated_at = format_now()
                state_manager.state.tasks[i] = updated
                await state_manager._save_unlocked()
                found = True
                break

    if not found:
        return JSONResponse(
            status_code=status.HTTP_404_NOT_FOUND,
            content={"success": False, "error": "task not found"},
        )

    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={"success": True, "task": updated.model_dump()},
    )


@app.delete("/api/tasks/{id}", dependencies=[Depends(require_api_auth)])
async def handle_delete_task(id: int) -> JSONResponse:
    """删除指定任务及其历史记录。"""
    found = False
    async with state_manager._lock:
        for i, t in enumerate(state_manager.state.tasks):
            if t.id == id:
                state_manager.state.tasks.pop(i)
                state_manager.state.history = [
                    h for h in state_manager.state.history if h.task_id != id
                ]
                await state_manager._save_unlocked()
                found = True
                break

    if not found:
        return JSONResponse(
            status_code=status.HTTP_404_NOT_FOUND,
            content={"success": False, "error": "task not found"},
        )

    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={"success": True},
    )


@app.get("/api/history", dependencies=[Depends(require_api_auth)])
async def handle_get_history() -> Dict[str, Any]:
    """获取所有运行历史记录。"""
    async with state_manager._lock:
        history = [h.model_dump() for h in state_manager.state.history]
    return {"success": True, "history": history}


@app.delete("/api/history", dependencies=[Depends(require_api_auth)])
async def handle_clear_history() -> Dict[str, Any]:
    """清空所有历史记录。"""
    async with state_manager._lock:
        state_manager.state.history = []
        await state_manager._save_unlocked()
    return {"success": True}


@app.put("/api/settings/notify", dependencies=[Depends(require_api_auth)])
async def handle_update_notify_settings(config: NotifySettings) -> Dict[str, Any]:
    """更新通知配置。"""
    async with state_manager._lock:
        state_manager.state.settings.notify = config
        await state_manager._save_unlocked()
    return {"success": True, "config": config.model_dump()}


@app.post("/api/settings/notify/test", dependencies=[Depends(require_api_auth)])
async def handle_test_notify() -> Dict[str, Any]:
    """发送测试通知。"""
    notify_config = state_manager.state.settings.notify
    success, err_msg = await send_test_telegram_notification(notify_config)
    if not success:
        return {"success": False, "error": err_msg}
    return {"success": True}


@app.put("/api/settings/schedule", dependencies=[Depends(require_api_auth)])
async def handle_update_schedule_settings(
    config: ScheduleSettings,
) -> JSONResponse:
    """更新全局定时配置。"""
    if (
        config.hour < 0
        or config.hour > 23
        or config.minute < 0
        or config.minute > 59
        or config.second < 0
        or config.second > 59
    ):
        return JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            content={"success": False, "error": "invalid schedule time"},
        )

    if config.max_workers < 1:
        config.max_workers = 1
    elif config.max_workers > 8:
        config.max_workers = 8

    async with state_manager._lock:
        state_manager.state.settings.schedule = config
        await state_manager._save_unlocked()

    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={"success": True, "config": config.model_dump()},
    )


@app.post("/api/settings/schedule/check", dependencies=[Depends(require_api_auth)])
async def handle_schedule_check() -> Dict[str, Any]:
    """手动触发一次定时检查。"""
    now = get_app_now()
    results = await run_scheduled_tasks(now, "manual-schedule-check", state_manager)
    return {"success": True, "results": [r.model_dump() for r in results]}


@app.put("/api/settings/security")
async def handle_update_security_settings(
    request: Request, payload: Dict[str, Any]
) -> JSONResponse:
    """更新安全/登录密码配置。"""
    security_enabled = state_manager.state.settings.security.enabled
    if security_enabled and not is_authenticated(request, security_enabled):
        return JSONResponse(
            status_code=status.HTTP_401_UNAUTHORIZED,
            content={"success": False, "error": "unauthorized"},
        )

    enabled: bool = bool(payload.get("enabled", True))
    password: str = str(payload.get("password", "")).strip()

    if password and len(password) < 6:
        return JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            content={"success": False, "error": "密码长度至少为 6 位"},
        )

    async with state_manager._lock:
        if password:
            state_manager.state.settings.security.password_hash = hash_password(
                password
            )
        state_manager.state.settings.security.enabled = enabled
        await state_manager._save_unlocked()

    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={"success": True, "config": {"enabled": enabled}},
    )


def main() -> None:
    """脚本运行主入口。"""
    uvicorn.run("app.main:app", host="0.0.0.0", port=8080, reload=False)


if __name__ == "__main__":
    main()
