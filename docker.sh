#!/bin/bash

echo "=== 推送 Docker 镜像到 DockerHub ==="

# 1. 检查是否安装 Docker
if ! command -v docker &> /dev/null; then
    echo "错误: 未找到 Docker 命令。请先安装 Docker。"
    exit 1
fi

# 2. 获取 DockerHub 用户名
DOCKER_USER=$1
if [ -z "$DOCKER_USER" ]; then
    read -p "请输入您的 DockerHub 用户名 (例如 iwanlebron): " DOCKER_USER
fi

if [ -z "$DOCKER_USER" ]; then
    echo "错误: 用户名不能为空。"
    exit 1
fi

IMAGE_NAME="stock-analysis"
FULL_IMAGE_NAME="$DOCKER_USER/$IMAGE_NAME:latest"

# 3. 登录 DockerHub
echo "请登录 DockerHub..."
docker login

# 4. 初始化 Buildx 构建器
echo "正在初始化 Buildx 构建器..."
docker buildx create --use --name multiarch-builder || docker buildx use multiarch-builder
docker buildx inspect --bootstrap

# 5. 使用 Buildx 构建多架构镜像并推送
echo "正在构建多架构镜像 (amd64, arm64)..."
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag $FULL_IMAGE_NAME \
  --push \
  .

echo "=== 完成！ ==="
echo "您现在可以使用以下命令拉取并运行镜像："
echo "docker run -p 8000:8000 $FULL_IMAGE_NAME"
