# 自动签到管理工具

一个基于Flask的自动化签到管理系统，支持解析Curl命令、定时签到、Telegram通知等功能，采用黑客帝国风格UI。

## 功能特点

### 🚀 核心功能
- **Curl命令解析**：自动解析curl命令，提取URL、 headers、data等信息
- **多网站管理**：支持添加、编辑、删除多个签到配置
- **定时签到**：支持全局定时和每个任务单独定时设置
- **Telegram通知**：签到结果实时推送到Telegram
- **请求方法选择**：支持GET和POST请求方法
- **成功检测**：基于HTTP状态码和响应内容分析签到结果

### 🔒 安全特性
- **密码保护**：支持启用密码访问，防止未授权访问
- **会话管理**：安全的用户会话管理

### 🎨 界面设计
- **黑客帝国风格**：绿色文字+黑色背景，带有数字雨效果
- **响应式布局**：适配不同屏幕尺寸
- **现代化交互**：平滑过渡动画，直观的操作界面

## 技术栈

- **后端**：Python 3.9+, Flask
- **前端**：HTML5, CSS3, JavaScript
- **网络**：Requests库
- **通知**：Telegram Bot API
- **容器化**：Docker, Docker Compose

## 快速开始

### 1. 环境要求

- Python 3.9 或更高版本
- pip 包管理器
- Docker (可选，用于容器化部署)

### 2. 安装依赖

```bash
# 克隆项目
git clone https://github.com/gitcoinkuang/qiandao.git
cd qiandao

# 安装依赖
pip3 install -r requirements.txt
```

### 3. 启动服务

```bash
# 开发模式
python3 app.py

# 生产模式 (使用Docker)
docker-compose up -d
```

服务将运行在 `http://IP:5000`

### 4. 首次访问

1. 打开浏览器访问 `http://IP:5000`
2. 进入「密码设置」页面，设置访问密码
3. 进入「通知设置」页面，配置Telegram Bot
4. 进入「添加签到」页面，粘贴Curl命令并保存

## 使用指南

### 添加签到

1. 在「添加签到」页面粘贴完整的curl命令
2. 输入网站名称
3. 选择请求方法 (GET/POST)
4. 设置定时选项（可选）
5. 点击「解析Curl」按钮
6. 确认解析结果后点击「保存签到」

### 管理签到

1. 在「管理签到」页面查看所有签到配置
2. 点击「运行签到」立即执行签到
3. 点击「编辑」修改现有签到配置
4. 点击「删除」移除不需要的签到

### 配置通知

1. 在Telegram中创建Bot并获取Token
2. 获取Chat ID (使用 @userinfobot)
3. 在「通知设置」页面填写配置
4. 点击「测试通知」验证配置是否正确

### 设置定时

1. 在「定时设置」页面启用全局定时
2. 设置每天执行签到的时间
3. 或在添加/编辑签到时设置单独的定时

## 项目结构

```
signin-manager/
├── app.py                 # 主应用文件
├── templates/             # HTML模板
│   ├── index.html         # 主页面
│   └── login.html         # 登录页面
├── Dockerfile             # Docker构建文件
├── docker-compose.yml     # Docker Compose配置
├── requirements.txt       # 依赖包列表
└── README.md              # 项目说明
```

## Docker部署

### 使用Docker Compose

```bash
docker-compose up -d
```

## 注意事项

1. **密码管理**：请妥善保管密码，忘记密码后需要删除 `password_config.json` 文件重置
2. **Curl命令格式**：请确保粘贴完整的curl命令，包括所有参数
3. **定时任务**：确保服务器在指定时间处于运行状态
4. **通知配置**：确保Telegram Bot已添加到您的聊天中并具有发送消息的权限
5. **安全性**：建议在生产环境中使用HTTPS和强密码

## 故障排查

### 常见问题

1. **签到失败**：检查curl命令是否正确，特别是headers和data部分
2. **通知不发送**：检查Telegram Bot Token和Chat ID是否正确
3. **定时任务不执行**：检查服务器时间和定时设置是否正确
4. **密码无法登录**：删除 `password_config.json` 文件重置密码

### 查看日志

```bash
# 直接运行时
python3 app.py

# Docker运行时
docker-compose logs qiandao-app
```

## 贡献

欢迎提交Issue和Pull Request来改进这个项目！

## 许可证

MIT License
=======
