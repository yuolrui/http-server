# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Run the server (listens on :8080)
go run ./cmd/http-server

# Run with custom routes config
go run ./cmd/http-server path/to/routes.json

# Build binary (current platform)
go build -o http-server.exe ./cmd/http-server

# Cross-compile all platforms
make build-all              # 使用 Makefile

# Or use scripts:
./build.sh                  # Linux/macOS/Git Bash
build.bat                   # Windows CMD
```

## Cross-Compilation Output

| Platform | Architecture | Output |
|----------|-------------|--------|
| Windows | amd64 | `http-server-1.0.0-windows-amd64.exe` |
| Windows | arm64 | `http-server-1.0.0-windows-arm64.exe` |
| Linux | amd64 | `http-server-1.0.0-linux-amd64` |
| Linux | arm64 | `http-server-1.0.0-linux-arm64` |
| macOS | amd64 | `http-server-1.0.0-darwin-amd64` |
| macOS | arm64 | `http-server-1.0.0-darwin-arm64` |

## Architecture

High-performance HTTP mock server with route activation control.

### Project Structure

```
http-server/
├── cmd/
│   └── http-server/
│       └── main.go           # 主入口
├── internal/
│   ├── config/
│   │   └── config.go         # 配置结构体和解析
│   ├── server/
│   │   └── server.go         # HTTPServer核心、路由加载、请求处理、监听
│   └── template/
│       └── template.go       # JSON模板渲染
├── data/                     # JSON响应数据
├── routes.json               # 路由配置
└── CLAUDE.md
```

### Route Configuration (routes.json)

```json
{
  "routes": [
    {
      "active": true,
      "method": "POST",
      "path": "/v1/chat-messages",
      "file": "data/dify/chat-messages.json",
      "delay": "2s"
    },
    {
      "active": false,
      "method": "GET",
      "path": "/products",
      "file": "data/common/products.json"
    }
  ]
}
```

**Fields:**
- `active`: `true` = loaded & responding, `false` = ignored
- `method`: HTTP method (GET, POST, PUT, DELETE, etc.)
- `path`: URL path, supports `:id` parameter syntax
- `file`: JSON data file path
- `delay`: Optional delay duration (e.g., `"500ms"`, `"2s"`, `"1.5s"`) to simulate slow API responses

### Activation Control

- Only `active=true` routes are loaded into memory
- Inactive routes return 404 (not found or inactive)
- Toggle routes by editing routes.json, changes apply within 1 second
- Perfect for managing 100+ routes with selective activation

### Performance Features

- Pre-loads only active route files (saves memory)
- Routes checked once on startup, cached in memory
- Hot reload: routes.json changes within 1s, data files within 500ms
- RWMutex for concurrent-safe access
- No routing overhead for inactive routes

### Data Files

Located in `data/<product>/` directories:
- `data/common/` - Common APIs
- `data/dify/` - Dify product APIs
- Create any subdirectory for new products

Template syntax: `{{.PathParams.id}}` for path parameters.

### Adding New Route

1. Create JSON file: `data/product/api.json`
2. Add to routes.json:

   ```json
   {"active": true, "method": "GET", "path": "/product/api", "file": "data/product/api.json"}
   ```

3. Server auto-loads within 1 second

### Adding New Feature

1. **配置相关** - 修改 `internal/config/config.go`
2. **服务器核心** - 修改 `internal/server/server.go`
3. **模板渲染** - 修改 `internal/template/template.go`
4. **更新文档** - 更新 CLAUDE.md 和 README.md
5. **测试验证** - `go build -o http-server.exe ./cmd/http-server`

## Dependencies

- Go 1.23+
- gin-gonic/gin v1.10.0