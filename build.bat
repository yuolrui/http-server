@echo off
REM HTTP Mock Server 交叉编译脚本 (Windows)

set OUTPUT_DIR=release
set VERSION=1.0.0

echo 清理构建目录...
if exist %OUTPUT_DIR% rmdir /s /q %OUTPUT_DIR%
mkdir %OUTPUT_DIR%

echo 开始交叉编译...

echo [1/6] Windows amd64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o %OUTPUT_DIR%/http-server-%VERSION%-windows-amd64.exe ./cmd/http-server

echo [2/6] Windows arm64...
set GOOS=windows
set GOARCH=arm64
go build -ldflags "-s -w" -o %OUTPUT_DIR%/http-server-%VERSION%-windows-arm64.exe ./cmd/http-server

echo [3/6] Linux amd64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o %OUTPUT_DIR%/http-server-%VERSION%-linux-amd64 ./cmd/http-server

echo [4/6] Linux arm64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags "-s -w" -o %OUTPUT_DIR%/http-server-%VERSION%-linux-arm64 ./cmd/http-server

echo [5/6] macOS amd64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-s -w" -o %OUTPUT_DIR%/http-server-%VERSION%-darwin-amd64 ./cmd/http-server

echo [6/6] macOS arm64...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags "-s -w" -o %OUTPUT_DIR%/http-server-%VERSION%-darwin-arm64 ./cmd/http-server

echo.
echo 编译完成！输出目录: %OUTPUT_DIR%
dir %OUTPUT_DIR%