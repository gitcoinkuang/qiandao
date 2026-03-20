# QianDao V2

一个面向 Docker 部署场景重写的自托管签到管理工具。

## 技术栈

- Go 1.23
- 标准库 HTTP 服务
- JSON 文件持久化
- 服务端模板 + 静态 CSS/JS

## 功能

- 任务增删改查
- 支持导入 `curl` 命令
- 支持并发执行全部任务
- 支持全局定时和单任务定时
- 支持 Telegram 和 Webhook 通知
- 默认启用登录保护
- 适合直接容器化部署

## 默认登录信息

- 默认账号体系只有密码，没有用户名
- 默认密码：`admin123456`
- 也可以通过环境变量 `QIANGDAO_DEFAULT_PASSWORD` 覆盖默认密码

如果你之前已经运行过项目，并且 `data/state.json` 里保存了旧配置，默认密码不会自动覆盖你已有的密码设置。

## 本地运行

```bash
go run ./cmd/server
```

打开 [http://localhost:8080](http://localhost:8080)。

## Docker 运行

```bash
docker compose up --build
```

打开 [http://localhost:8080](http://localhost:8080)。
