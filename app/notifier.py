"""消息通知推送模块。

支持通过 Telegram Bot API 与自定义 Webhook 发送任务执行结果通知。
"""

from typing import Tuple

import httpx

from app.models import HistoryItem, NotifySettings

# 全局共享 HTTP 客户端
_http_client = httpx.AsyncClient(timeout=10.0)


async def send_notifications(item: HistoryItem, notify: NotifySettings) -> None:
    """根据通知策略异步推送任务结果消息。"""
    if item.status == "success" and not notify.notify_on_success:
        return
    if item.status == "failed" and not notify.notify_on_failure:
        return

    message = (
        f"{item.status.upper()}\n"
        f"Task: {item.task_name}\n"
        f"By: {item.triggered_by}\n"
        f"Status code: {item.status_code}\n"
        f"Duration: {item.response_time_ms} ms\n"
        f"Message: {item.message}\n"
        f"Preview:\n{item.response_preview}"
    )

    # 推送 Telegram 通知
    if notify.telegram_enabled and notify.telegram_bot_token and notify.telegram_chat_id:
        tg_url = f"https://api.telegram.org/bot{notify.telegram_bot_token}/sendMessage"
        try:
            await _http_client.post(
                tg_url,
                data={"chat_id": notify.telegram_chat_id, "text": message},
            )
        except Exception:
            pass

    # 推送 Webhook 通知
    if notify.webhook_enabled and notify.webhook_url:
        try:
            await _http_client.post(
                notify.webhook_url,
                json={"title": item.status, "message": message, "item": item.model_dump()},
            )
        except Exception:
            pass


async def send_test_telegram_notification(notify: NotifySettings) -> Tuple[bool, str]:
    """尝试发送 Telegram 测试通知。返回 (成功标志, 错误消息)。"""
    if not (notify.telegram_enabled and notify.telegram_bot_token and notify.telegram_chat_id):
        return False, "Telegram 未配置或未启用，请先保存通知设置"

    tg_url = f"https://api.telegram.org/bot{notify.telegram_bot_token}/sendMessage"
    message = "This is a test notification from QianDao V2."

    try:
        resp = await _http_client.post(
            tg_url,
            data={"chat_id": notify.telegram_chat_id, "text": message},
        )
        if resp.status_code != 200:
            return False, f"Telegram API 返回异常: {resp.text}"
        return True, ""
    except Exception as exc:
        return False, f"发送请求失败: {str(exc)}"
