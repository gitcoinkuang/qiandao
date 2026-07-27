"""身份认证与安全模块。

处理密码 SHA256 哈希算法、Cookie Session 管理以及 FastAPI API 认证依赖。
"""

import hashlib
import secrets
from datetime import datetime, timedelta
from typing import Dict, Optional

from fastapi import HTTPException, Request, status

from app.config import get_app_now, get_default_password

# 内存 Session 存储：token -> 过期时间 datetime
_sessions: Dict[str, datetime] = {}


def hash_password(password: str) -> str:
    """计算密码的 SHA256 哈希值。"""
    return hashlib.sha256(password.encode("utf-8")).hexdigest()


def generate_token(length: int = 32) -> str:
    """生成指定长度的安全随机十六进制 Token。"""
    return secrets.token_hex(length)


def create_session(token: str) -> None:
    """创建新的 Session 记录，有效期 7 天。"""
    _sessions[token] = get_app_now() + timedelta(days=7)


def delete_session(token: str) -> None:
    """删除指定的 Session。"""
    _sessions.pop(token, None)


def is_authenticated(request: Request, security_enabled: bool) -> bool:
    """判断请求是否通过认证。

    若安全校验未启用，始终返回 True；
    若启用了安全校验，则验证 Cookie 中的 qiandao_session。
    """
    if not security_enabled:
        return True

    cookie_val: Optional[str] = request.cookies.get("qiandao_session")
    if not cookie_val:
        return False

    expires_at: Optional[datetime] = _sessions.get(cookie_val)
    if not expires_at:
        return False

    if get_app_now() > expires_at:
        _sessions.pop(cookie_val, None)
        return False

    return True


def check_password(password: str, stored_hash: str, security_enabled: bool) -> bool:
    """校验明文密码是否匹配。"""
    if not security_enabled:
        return True
    return stored_hash == hash_password(password)
