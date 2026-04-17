#!/bin/bash

# HTTP Mock Server 交叉编译脚本
# 支持 Windows、Linux、macOS 多平台多架构

set -e

VERSION=${VERSION:-"1.0.0"}
OUTPUT_DIR="release"
MAIN_PATH="./cmd/http-server"

# 清理构建目录
echo "清理构建目录..."
rm -rf ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}

# 编译函数
build() {
    GOOS=$1
    GOARCH=$2
    OUTPUT_NAME=$3

    echo "编译 ${GOOS}/${GOARCH}..."
    GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "-s -w" -o ${OUTPUT_DIR}/${OUTPUT_NAME} ${MAIN_PATH}

    if [ $? -eq 0 ]; then
        echo "✓ 成功: ${OUTPUT_DIR}/${OUTPUT_NAME}"
    else
        echo "✗ 失败: ${GOOS}/${GOARCH}"
    fi
}

# 交叉编译
echo "开始交叉编译..."
echo ""

# Windows
build "windows" "amd64" "http-server-${VERSION}-windows-amd64.exe"
build "windows" "arm64" "http-server-${VERSION}-windows-arm64.exe"

# Linux
build "linux" "amd64" "http-server-${VERSION}-linux-amd64"
build "linux" "arm64" "http-server-${VERSION}-linux-arm64"

# macOS
build "darwin" "amd64" "http-server-${VERSION}-darwin-amd64"
build "darwin" "arm64" "http-server-${VERSION}-darwin-arm64"

echo ""
echo "编译完成！输出目录: ${OUTPUT_DIR}"
echo ""
ls -lh ${OUTPUT_DIR}