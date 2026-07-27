FROM python:3.12-slim

WORKDIR /app

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    APP_ADDR=0.0.0.0:8080 \
    APP_TIMEZONE=Asia/Shanghai \
    TZ=Asia/Shanghai

# 安装 tzdata 时区组件
RUN apt-get update && apt-get install -y --no-install-recommends tzdata && rm -rf /var/lib/apt/lists/*

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY app ./app
COPY templates ./templates
COPY static ./static

RUN mkdir -p /app/data

EXPOSE 8080

CMD ["python", "-m", "app.main"]
