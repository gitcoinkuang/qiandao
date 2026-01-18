# 自动签到管理工具

一个基于 Flask 开发的自动签到管理工具，支持多网站签到、定时任务、Telegram 通知等功能。

## 功能特性

- ✅ 支持通过 Curl 命令添加签到配置
- ✅ 支持多网站同时管理
- ✅ 支持手动运行单个或全部签到
- ✅ 支持定时自动签到
- ✅ 支持 Telegram 通知
- ✅ 支持密码保护
- ✅ 支持签到历史记录
- ✅ 支持响应内容查看
- ✅ 支持 Docker 部署
- ✅ 支持日志记录

## 技术栈

- **后端**：Flask 2.3.2
- **HTTP 客户端**：requests 2.31.0
- **安全保护**：Flask-WTF 1.2.1（CSRF 保护）
- **速率限制**：flask-limiter 3.5.0
- **部署方式**：Docker

## 安装部署

### 直接运行

1. 克隆项目
2. 安装依赖：`pip install -r requirements.txt`
3. 运行应用：`python app.py`
4. 访问：http://localhost:5000

### Docker 部署

1. 构建镜像：`docker build -t auto-signin .`
2. 运行容器：`docker run -d -p 5000:5000 auto-signin`
3. 访问：http://localhost:5000

## 使用说明

### 1. 添加签到配置

1. 切换到「添加签到」标签页
2. 输入 Curl 命令
3. 输入网站名称
4. 选择请求方法（GET/POST）
5. 可选：设置定时签到
6. 点击「解析 Curl」
7. 点击「保存签到」

### 2. 管理签到配置

1. 切换到「管理签到」标签页
2. 可以查看所有签到配置
3. 支持手动运行单个签到
4. 支持编辑和删除签到配置
5. 支持运行全部签到

### 3. 查看签到结果

- 手动运行签到后，会显示详细的返回内容
- 可以查看状态码和返回内容预览
- 成功信息会在 5 秒后自动隐藏
- 失败信息会一直显示，直到手动刷新

## 配置说明

### Telegram 通知设置

1. 切换到「通知设置」标签页
2. 输入 Bot Token（从 @BotFather 获取）
3. 输入 Chat ID（从 @userinfobot 获取）
4. 点击「保存配置」
5. 点击「测试通知」验证配置

### 定时任务设置

1. 切换到「定时设置」标签页
2. 启用定时签到
3. 设置签到时间（小时和分钟）
4. 点击「保存配置」

### 密码保护设置

1. 切换到「密码设置」标签页
2. 启用密码保护
3. 输入新密码（至少 6 个字符）
4. 点击「保存配置」
5. 保存后会自动跳转到登录页面

## 日志记录

- 日志文件：`app.log`
- 日志级别：INFO
- 包含签到开始、成功、失败等信息
- 包含返回内容预览
- 包含 HTTP 状态码和错误信息

## 历史记录

- 历史记录文件：`signin_history.json`
- 保留最近 100 条记录
- 包含签到时间、网站名称、URL、方法、结果、状态码等信息

## 开发说明

### 项目结构

```
.
├── app.py              # 主应用文件
├── requirements.txt     # 依赖列表
├── Dockerfile          # Docker 构建文件
├── .dockerignore       # Docker 忽略文件
├── signin_history.json # 签到历史记录
├── signin_configs.json # 签到配置
├── notify_config.json  # 通知配置
├── schedule_config.json # 定时任务配置
├── password_config.json # 密码配置
└── templates/          # 模板目录
    └── index.html      # 主页面
```

### 主要功能模块

1. **parse_curl**：解析 Curl 命令
2. **run_signin**：执行签到
3. **send_telegram_notification**：发送 Telegram 通知
4. **check_schedule**：检查并执行定时任务
5. **login_required**：登录验证装饰器

## 许可证

MIT License
