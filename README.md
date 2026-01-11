# 自动签到管理工具

一个基于Flask的自动签到管理工具，支持多网站自动签到、定时任务、Telegram通知等功能。

## 功能特性

- **Curl命令解析**：支持解析Curl命令并转换为可执行的签到任务
- **多网站管理**：支持添加、编辑、删除多个网站的签到任务
- **定时签到**：支持全局定时和每个任务单独设置定时
- **Telegram通知**：签到结果通过Telegram机器人发送通知
- **密码保护**：支持设置密码保护，防止未授权访问
- **请求方法选择**：支持GET和POST请求方法
- **详细错误信息**：失败时显示详细的错误信息，成功时不显示冗余信息
- **黑客帝国风格UI**：具有数字雨效果的Matrix风格界面
- **Docker部署**：支持使用Docker容器化部署

## 技术栈

- **后端**：Flask 2.3.2
- **前端**：HTML/CSS/JavaScript
- **HTTP客户端**：requests 2.31.0
- **容器化**：Docker
- **通知**：Telegram Bot API

## 安装部署

### 方法一：直接运行

1. **克隆项目**
   ```bash
   git clone https://github.com/gitcoinkuang/qiandao.git
   cd qiandao
   ```

2. **安装依赖**
   ```bash
   pip install -r requirements.txt
   ```

3. **运行项目**
   ```bash
   python app.py
   ```

4. **访问应用**
   打开浏览器访问 `http://localhost:5000`

### 方法二：Docker部署

1. **克隆项目**
   ```bash
   git clone https://github.com/gitcoinkuang/qiandao.git
   cd qiandao
   ```

2. **构建并运行容器**
   ```bash
   docker-compose up --build -d
   ```

3. **访问应用**
   打开浏览器访问 `http://localhost:5000`

## 使用方法

### 1. 添加签到任务

1. 在"添加签到"标签页中，粘贴Curl命令
2. 输入网站名称
3. 选择请求方法（GET/POST）
4. 可选：设置定时任务
5. 点击"解析Curl"按钮
6. 点击"保存签到"按钮

### 2. 管理签到任务

1. 在"管理签到"标签页中，可以查看所有已添加的签到任务
2. 点击"运行签到"按钮执行单个签到任务
3. 点击"编辑"按钮修改现有签到任务
4. 点击"删除"按钮删除不需要的签到任务
5. 点击"运行全部签到"按钮执行所有签到任务

### 3. 配置通知

1. 在"通知设置"标签页中，输入Telegram Bot Token和Chat ID
2. 点击"保存配置"按钮
3. 点击"测试通知"按钮验证配置是否正确

### 4. 配置定时任务

1. 在"定时设置"标签页中，设置全局定时时间
2. 点击"保存配置"按钮
3. 或者在添加/编辑签到任务时，设置单个任务的定时时间

### 5. 设置密码保护

1. 在"密码设置"标签页中，启用密码保护
2. 输入密码
3. 点击"保存配置"按钮
4. 下次访问时需要输入密码

## 配置文件

项目使用以下配置文件：

- `signin_configs.json`：签到任务配置
- `notify_config.json`：Telegram通知配置
- `schedule_config.json`：定时任务配置
- `password_config.json`：密码保护配置

## 常见问题

### 1. Curl命令解析失败

- 确保Curl命令格式正确
- 确保包含完整的URL和必要的请求头

### 2. 签到失败显示HTTP 403错误

- 检查Curl命令中的Cookie信息是否正确
- 检查请求头是否包含必要的信息（如User-Agent）

### 3. Telegram通知不工作

- 确保Bot Token正确
- 确保Chat ID正确
- 确保机器人已添加到聊天中并获得发送消息的权限

### 4. 定时任务不执行

- 确保定时时间设置正确
- 确保应用程序持续运行

## 项目结构

```
qiandao/
├── app.py                 # 主应用文件
├── templates/
│   ├── index.html         # 主页面
│   └── login.html         # 登录页面
├── requirements.txt       # 依赖项
├── Dockerfile             # Docker构建文件
├── docker-compose.yml     # Docker Compose配置
└── README.md              # 项目说明
```

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License

## 项目地址

[https://github.com/gitcoinkuang/qiandao](https://github.com/gitcoinkuang/qiandao)