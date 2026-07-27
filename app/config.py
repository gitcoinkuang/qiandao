"""配置与时间处理模块。

定义应用运行所需的环境变量解析与带时区的时间处理函数。
"""

import os
from datetime import datetime
from zoneinfo import ZoneInfo, ZoneInfoNotFoundError

# 默认环境变量值
DEFAULT_TIMEZONE: str = "Asia/Shanghai"
DEFAULT_ADMIN_PASSWORD: str = "admin123456"
DEFAULT_ADDR: str = "0.0.0.0:8080"


def get_app_timezone():
    """获取应用配置的时区。

    若 APP_TIMEZONE 无效，则尝试回退到 Asia/Shanghai；若均不可用，回退到本地系统时区。
    """
    tz_name: str = os.getenv("APP_TIMEZONE", DEFAULT_TIMEZONE).strip()
    try:
        return ZoneInfo(tz_name)
    except Exception:
        try:
            return ZoneInfo(DEFAULT_TIMEZONE)
        except Exception:
            # 兜底：使用本地系统当前时区
            local_tz = datetime.now().astimezone().tzinfo
            return local_tz if local_tz is not None else ZoneInfo("UTC")


def get_app_now() -> datetime:
    """获取当前时区的时间对象。"""
    return datetime.now(get_app_timezone())


def format_now() -> str:
    """格式化当前时间为标准日期时间字符串 (YYYY-MM-DD HH:MM:SS)。"""
    return get_app_now().strftime("%Y-%m-%d %H:%M:%S")


def format_now_ms() -> str:
    """格式化当前时间为带毫秒的日期时间字符串 (YYYY-MM-DD HH:MM:SS.mmm)。"""
    now: datetime = get_app_now()
    return now.strftime("%Y-%m-%d %H:%M:%S.") + f"{now.microsecond // 1000:03d}"


def get_default_password() -> str:
    """获取默认管理员密码。"""
    val: str = os.getenv("QIANGDAO_DEFAULT_PASSWORD", "").strip()
    return val if val else DEFAULT_ADMIN_PASSWORD
