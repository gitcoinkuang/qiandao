# QianDao V2

A rebuilt self-hosted sign-in control panel designed for Docker-first usage.

## Stack

- Go 1.23
- Standard library HTTP server
- JSON file persistence
- Server-rendered shell with static CSS/JS

## Features

- Task CRUD with curl import
- Concurrent run-all execution
- Global schedule and per-task schedule
- Telegram and webhook notifications
- Optional login protection
- Docker-ready deployment

## Run locally

```bash
go run ./cmd/server
```

Open [http://localhost:8080](http://localhost:8080).

## Run with Docker

```bash
docker compose up --build
```

Open [http://localhost:8080](http://localhost:8080).
