# QianDao V2

一个面向 Docker 部署场景重写的自托管签到管理工具，支持保存任务、定时执行、通知推送，以及面向整点签到的抢零点优化。

## 技术栈

- Go 1.23
- 标准库 HTTP 服务
- JSON 文件持久化
- 服务端模板 + 静态 CSS/JS

## 功能特性

- 任务增删改查
- 支持导入 `curl` 命令
- 支持手动执行单任务和批量执行
- 支持全局定时和单任务定时
- 支持 Telegram 和 Webhook 通知
- 默认启用登录保护
- 支持抢零点模式
- 历史记录展示真实发起时间

## 抢零点模式说明

针对需要抢 `00:00` 首签的任务，可以在任务编辑页勾选“抢零点模式”。

开启后会有这些优化：

- 调度按秒级检查，而不是 15 秒轮询
- 整点触发后会在短窗口内进行更快的补发尝试
- 请求使用共享连接池，减少重复建连开销
- 历史记录中会显示真实请求发起时间，方便排查是调度慢还是站点响应慢

注意：

- 这会明显提升整点签到成功概率，但不能绝对保证第一
- 实际结果仍然会受 VPS 网络、目标站点响应速度、站点风控策略影响
- 建议只给需要抢零点的任务开启，普通任务没必要启用

## 时区说明

项目默认使用 `Asia/Shanghai` 时区。

这是为了避免容器默认使用 `UTC` 时，出现“设置 00:00，实际 08:00 执行”的问题。

默认会同时通过下面两个环境变量保证时区一致：

- `TZ=Asia/Shanghai`
- `APP_TIMEZONE=Asia/Shanghai`

如果你部署在其他时区，可以自行修改 `docker-compose.yml`。

## 默认登录信息

- 默认没有用户名，只有密码
- 默认密码：`admin123456`
- 可通过环境变量 `QIANGDAO_DEFAULT_PASSWORD` 覆盖默认密码

如果你之前已经运行过项目，并且 `data/state.json` 里保存过旧配置，那么默认密码不会覆盖你已有的安全设置。

## 本地运行

```bash
go run ./cmd/server
```

打开 [http://localhost:8080](http://localhost:8080)

## Docker 运行

```bash
docker compose up --build -d
```

打开 [http://localhost:8080](http://localhost:8080)

## 更新部署

如果你已经部署在 VPS 上，更新代码后建议这样升级：

```bash
git pull
docker compose down
docker compose up --build -d
```

说明：

- `git pull` 只会更新 Git 跟踪的文件
- 如果你手动删除了受版本控制的文件，比如 `docker-compose.yml`，需要先用 `git restore docker-compose.yml` 恢复
- 这次更新如果涉及 Go 代码或前端静态文件，记得使用 `--build`

## 数据目录

项目数据默认保存在：

- `data/state.json`

这个文件会保存：

- 任务配置
- 定时配置
- 通知配置
- 安全配置
- 历史记录
