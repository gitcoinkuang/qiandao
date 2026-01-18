FROM python:3.9-slim

# 安装系统依赖（只需要gcc用于可能的依赖编译）
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /app

# 复制项目文件
COPY . /app

# 安装Python依赖
RUN pip install --no-cache-dir -r requirements.txt

# 暴露端口
EXPOSE 5000

# 设置Flask为生产模式
ENV FLASK_ENV=production
ENV FLASK_APP=app.py

# 运行应用
CMD ["python", "app.py"]
