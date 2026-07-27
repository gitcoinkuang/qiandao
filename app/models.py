"""数据模型定义模块。

使用 Pydantic 定义 API 请求/响应以及持久化数据的类型规范，与旧 Go 版数据保持 100% 结构兼容。
"""

from typing import Any, Dict, List, Optional
from pydantic import BaseModel, Field


class Task(BaseModel):
    """签到任务数据模型。"""

    id: int = 0
    name: str
    url: str
    method: str = "GET"
    headers: Dict[str, str] = Field(default_factory=dict)
    body: str = ""
    curl_command: str = ""
    enabled: bool = False
    schedule_enabled: bool = False
    schedule_hour: int = Field(default=8, ge=0, le=23)
    schedule_minute: int = Field(default=0, ge=0, le=59)
    schedule_second: int = Field(default=0, ge=0, le=59)
    timeout_seconds: int = Field(default=30, ge=1, le=120)
    retry_count: int = Field(default=0, ge=0, le=5)
    aggressive_mode: bool = False
    success_keywords: str = ""
    failure_keywords: str = ""
    last_status: str = "idle"
    last_run_at: str = ""
    last_duration_ms: int = 0
    created_at: str = ""
    updated_at: str = ""


class HistoryItem(BaseModel):
    """历史执行记录数据模型。"""

    id: int = 0
    task_id: int
    task_name: str
    status: str
    status_code: int = 0
    message: str = ""
    response_preview: str = ""
    response_time_ms: int = 0
    request_started_at: str = ""
    triggered_by: str = ""
    created_at: str = ""


class NotifySettings(BaseModel):
    """通知推送设置数据模型。"""

    telegram_enabled: bool = False
    telegram_bot_token: str = ""
    telegram_chat_id: str = ""
    webhook_enabled: bool = False
    webhook_url: str = ""
    notify_on_success: bool = False
    notify_on_failure: bool = True


class ScheduleSettings(BaseModel):
    """全局定时调度设置数据模型。"""

    enabled: bool = False
    hour: int = Field(default=8, ge=0, le=23)
    minute: int = Field(default=0, ge=0, le=59)
    second: int = Field(default=0, ge=0, le=59)
    max_workers: int = Field(default=4, ge=1, le=8)


class SecuritySettings(BaseModel):
    """访问安全/登录密码设置数据模型。"""

    enabled: bool = True
    password_hash: str = ""


class Settings(BaseModel):
    """系统全局配置数据模型。"""

    notify: NotifySettings = Field(default_factory=NotifySettings)
    schedule: ScheduleSettings = Field(default_factory=ScheduleSettings)
    security: SecuritySettings = Field(default_factory=SecuritySettings)


class AppState(BaseModel):
    """系统状态持久化存储模型（即 state.json）。"""

    next_task_id: int = 1
    next_history_id: int = 1
    tasks: List[Task] = Field(default_factory=list)
    history: List[HistoryItem] = Field(default_factory=list)
    settings: Settings = Field(default_factory=Settings)


class APIResponse(BaseModel):
    """统一 API 响应格式模型。"""

    success: bool
    error: Optional[str] = None
    data: Optional[Any] = None
    task: Optional[Any] = None
    tasks: Optional[Any] = None
    result: Optional[Any] = None
    results: Optional[Any] = None
    config: Optional[Any] = None
    history: Optional[Any] = None
