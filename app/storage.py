"""状态持久化存储模块。

负责从 data/state.json 读取和保存应用状态，保证线程/协程安全。
"""

import asyncio
import json
import os
from typing import Any, Dict

from app.auth import get_default_password, hash_password
from app.models import AppState, NotifySettings, ScheduleSettings, SecuritySettings, Settings


class StateManager:
    """系统状态管理器，提供线程/协程安全的 JSON 文件读写功能。"""

    def __init__(self, state_path: str = os.path.join("data", "state.json")) -> None:
        self.state_path: str = state_path
        self._lock: asyncio.Lock = asyncio.Lock()
        self.state: AppState = self._default_state()

    def _default_state(self) -> AppState:
        """生成系统默认配置状态。"""
        return AppState(
            next_task_id=1,
            next_history_id=1,
            tasks=[],
            history=[],
            settings=Settings(
                notify=NotifySettings(notify_on_failure=True),
                schedule=ScheduleSettings(hour=8, minute=0, second=0, max_workers=4),
                security=SecuritySettings(
                    enabled=True,
                    password_hash=hash_password(get_default_password()),
                ),
            ),
        )

    async def load(self) -> None:
        """异步加载 data/state.json 配置文件。如果文件不存在或损坏则使用默认状态。"""
        async with self._lock:
            if not os.path.exists(self.state_path):
                os.makedirs(os.path.dirname(self.state_path), exist_ok=True)
                self.state = self._default_state()
                await self._save_unlocked()
                return

            try:
                with open(self.state_path, "r", encoding="utf-8") as f:
                    raw_data: Dict[str, Any] = json.load(f)
                state_obj = AppState.model_validate(raw_data)
            except Exception:
                state_obj = self._default_state()

            # 兼容性容错修复
            if state_obj.next_task_id < 1:
                state_obj.next_task_id = 1
            if state_obj.next_history_id < 1:
                state_obj.next_history_id = 1
            if not state_obj.settings.security.password_hash:
                state_obj.settings.security.password_hash = hash_password(get_default_password())
                state_obj.settings.security.enabled = True

            self.state = state_obj

    async def save(self) -> None:
        """保存状态到磁盘（加锁）。"""
        async with self._lock:
            await self._save_unlocked()

    async def _save_unlocked(self) -> None:
        """写入临时文件后重命名以实现原子写入。"""
        os.makedirs(os.path.dirname(self.state_path), exist_ok=True)
        tmp_path: str = f"{self.state_path}.tmp"
        json_data: str = self.state.model_dump_json(indent=2)
        with open(tmp_path, "w", encoding="utf-8") as f:
            f.write(json_data)
        os.replace(tmp_path, self.state_path)
