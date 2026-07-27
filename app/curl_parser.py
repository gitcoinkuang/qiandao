"""Curl 命令解析与构造模块。

支持解析粘贴的 curl 命令行参数提取 URL、HTTP 方法、请求头和请求体，以及将 Task 重新合成为 curl 命令。
"""

import re
from typing import Dict, Tuple

from app.models import Task


def first_non_empty(*values: str) -> str:
    """返回参数列表中第一个非空字符串。"""
    for val in values:
        if val:
            return val
    return ""


def parse_curl(command: str) -> Tuple[str, str, Dict[str, str], str]:
    """解析 Curl 命令行字符串。

    返回 (url, method, headers, body)。
    """
    command = command.strip()
    if not command:
        raise ValueError("curl 命令不能为空")

    url: str = ""
    method: str = "GET"
    headers: Dict[str, str] = {}
    body: str = ""

    # 提取 URL
    url_pattern = re.compile(r"curl\s+(?:'([^']+)'|\"([^\"]+)\"|([^\s]+))")
    match_url = url_pattern.search(command)
    if match_url:
        url = first_non_empty(match_url.group(1) or "", match_url.group(2) or "", match_url.group(3) or "")

    # 提取 请求方法 (-X 或者 --request)
    method_pattern = re.compile(r"(?:-X|--request)\s+([A-Za-z]+)")
    match_method = method_pattern.search(command)
    if match_method:
        method = match_method.group(1).upper()

    # 提取 请求头 (-H 或者 --header)
    header_pattern = re.compile(r"(?:-H|--header)\s+(?:'([^']+)'|\"([^\"]+)\")")
    for match in header_pattern.finditer(command):
        header_str = first_non_empty(match.group(1) or "", match.group(2) or "")
        if ":" in header_str:
            parts = header_str.split(":", 1)
            headers[parts[0].strip()] = parts[1].strip()

    # 提取 Cookie (-b 或者 --cookie)
    cookie_pattern = re.compile(r"(?:-b|--cookie)\s+(?:'([^']+)'|\"([^\"]+)\")")
    match_cookie = cookie_pattern.search(command)
    if match_cookie:
        headers["Cookie"] = first_non_empty(match_cookie.group(1) or "", match_cookie.group(2) or "")

    # 提取 请求体 (-d / --data / --data-raw / --data-binary)
    body_pattern = re.compile(r"(?:-d|--data|--data-raw|--data-binary)\s+(?:'([^']*)'|\"([^\"]*)\")")
    match_body = body_pattern.search(command)
    if match_body:
        body = first_non_empty(match_body.group(1) or "", match_body.group(2) or "")
        if method == "GET":
            method = "POST"

    if not url:
        raise ValueError("无法从 curl 命令中提取有效的 URL")

    return url, method, headers, body


def build_curl(task: Task) -> str:
    """将 Task 配置构建成标准 Curl 命令字符串。"""
    parts = [f"curl '{task.url}'"]
    if task.method != "GET":
        parts.append(f"-X {task.method}")
    for key, value in task.headers.items():
        parts.append(f"-H '{key}: {value}'")
    if task.body:
        escaped_body = task.body.replace("'", "'\"'\"'")
        parts.append(f"-d '{escaped_body}'")
    return " \\\n  ".join(parts)


def normalize_task(input_task: Task) -> Task:
    """规范化并校验任务配置参数。"""
    task = input_task.model_copy()

    if task.curl_command and task.curl_command.strip():
        parsed_url, parsed_method, parsed_headers, parsed_body = parse_curl(task.curl_command)
        if not task.url:
            task.url = parsed_url
        if not task.headers:
            task.headers = parsed_headers
        if not task.body:
            task.body = parsed_body
        if not task.method:
            task.method = parsed_method

    task.name = task.name.strip()
    task.url = task.url.strip()
    task.method = task.method.strip().upper()

    if not task.name:
        raise ValueError("任务名称为必填项")
    if not (task.url.startswith("http://") or task.url.startswith("https://")):
        raise ValueError("任务 URL 必须以 http:// 或 https:// 开头")
    if not task.method:
        task.method = "GET"
    if task.method not in ["GET", "POST", "PUT", "PATCH", "DELETE"]:
        raise ValueError("不支持的 HTTP 请求方法")

    if task.timeout_seconds < 1:
        task.timeout_seconds = 30
    elif task.timeout_seconds > 120:
        task.timeout_seconds = 120

    if task.retry_count < 0:
        task.retry_count = 0
    elif task.retry_count > 5:
        task.retry_count = 5

    if task.schedule_hour < 0 or task.schedule_hour > 23:
        task.schedule_hour = 8
    if task.schedule_minute < 0 or task.schedule_minute > 59:
        task.schedule_minute = 0
    if task.schedule_second < 0 or task.schedule_second > 59:
        task.schedule_second = 0

    if not task.curl_command:
        task.curl_command = build_curl(task)

    return task
