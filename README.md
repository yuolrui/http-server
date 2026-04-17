# HTTP Mock Server

基于 Go 和 Gin 构建的高性能 HTTP Mock 服务器，支持路由激活控制。

## 特性

- **路由激活控制**：通过 `active` 字段启用/禁用路由，无需重启服务
- **热重载**：routes.json 变更 1 秒内生效，数据文件变更 500ms 内生效
- **路径参数**：支持 `:id` 风格的路径参数及模板替换
- **响应延迟**：支持配置延迟时间模拟慢接口
- **内存高效**：仅加载激活的路由文件到内存
- **并发安全**：使用 RWMutex 保护并发访问
- **优雅关闭**：支持 5 秒超时的优雅关闭

## 快速开始

```bash
# 使用默认 routes.json 运行
go run ./cmd/http-server

# 使用自定义路由配置运行
go run ./cmd/http-server path/to/routes.json

# 编译当前平台
go build -o http-server.exe ./cmd/http-server

# 交叉编译所有平台
make build-all              # 使用 Makefile
./build.sh                  # Linux/macOS/Git Bash
build.bat                   # Windows CMD
```

> **注意**：项目采用标准 Go 项目结构，入口在 `cmd/http-server/`，必须指定路径运行。

服务器默认监听 `:8080` 端口。

## 交叉编译

支持多平台多架构编译：

| 平台 | 架构 | 输出文件 |
|------|------|---------|
| Windows | amd64 | `http-server-1.0.0-windows-amd64.exe` |
| Windows | arm64 | `http-server-1.0.0-windows-arm64.exe` |
| Linux | amd64 | `http-server-1.0.0-linux-amd64` |
| Linux | arm64 | `http-server-1.0.0-linux-arm64` |
| macOS | amd64 | `http-server-1.0.0-darwin-amd64` |
| macOS | arm64 | `http-server-1.0.0-darwin-arm64` |

编译产物输出到 `release/` 目录。

## 项目结构

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
│   └ template/
│       └ template.go         # JSON模板渲染
├── data/                     # JSON响应数据
│   ├── common/
│   └── dify/
├── routes.json               # 路由配置
├── README.md
└── CLAUDE.md
```

| 目录/文件 | 职责 |
|----------|------|
| `cmd/http-server/` | 主入口、服务器启动、优雅关闭 |
| `internal/config/` | 配置结构体、延迟解析 |
| `internal/server/` | HTTPServer结构体、路由加载、缓存管理、请求处理、监听 |
| `internal/template/` | JSON模板渲染 |
| `data/` | JSON响应数据文件 |
| `build.sh` | Linux/macOS/Git Bash 交叉编译脚本 |
| `build.bat` | Windows CMD 交叉编译脚本 |
| `Makefile` | Make 构建脚本 |

## Make 常用命令

```bash
make build          # 编译当前平台
make build-all      # 交叉编译所有平台
make run            # 运行服务器
make clean          # 清理构建产物
make fmt            # 格式化代码
make vet            # 静态检查
make help           # 显示所有可用命令
```

## 路由配置

路由在 `routes.json` 中定义：

```json
{
  "routes": [
    {
      "active": true,
      "method": "GET",
      "path": "/users",
      "file": "data/common/users.json"
    },
    {
      "active": false,
      "method": "GET",
      "path": "/products",
      "file": "data/common/products.json"
    },
    {
      "active": true,
      "method": "GET",
      "path": "/users/:id",
      "file": "data/common/user.json"
    },
    {
      "active": true,
      "method": "POST",
      "path": "/v1/chat-messages",
      "file": "data/dify/chat-messages.json",
      "delay": "2s"
    }
  ]
}
```

### 字段说明

| 字段 | 类型 | 描述 |
|------|------|------|
| `active` | boolean | `true` = 加载并响应，`false` = 忽略（返回 404） |
| `method` | string | HTTP 方法（GET、POST、PUT、DELETE 等） |
| `path` | string | URL 路径，支持 `:id` 参数语法 |
| `file` | string | JSON 数据文件路径（相对于工作目录） |
| `delay` | string | 可选，响应延迟时间（如 `"500ms"`、`"2s"`） |

## 数据文件

数据文件存放在 `data/<product>/` 目录下：

```
data/
├── common/          # 通用 API
│   ├── hello.json
│   ├── users.json
│   ├── user.json
│   ├── products.json
│   ├── product.json
│   └── orders.json
└── dify/            # Dify 产品 API
    └── chat-messages.json
```

### 模板变量

使用 `{{.PathParams.<name>}}` 语法进行路径参数替换：

**data/common/user.json**:
```json
{
  "id": "{{.PathParams.id}}",
  "name": "张三",
  "email": "zhangsan@example.com"
}
```

请求 `GET /users/123` 返回：
```json
{
  "id": "123",
  "name": "张三",
  "email": "zhangsan@example.com"
}
```

## 添加新路由

1. 创建 JSON 文件：`data/product/api.json`
2. 在 `routes.json` 中添加路由：
   ```json
   {"active": true, "method": "GET", "path": "/product/api", "file": "data/product/api.json"}
   ```
3. 服务器在 1 秒内自动加载

## 切换路由状态

直接编辑 `routes.json`，将 `active` 设为 `true` 或 `false`：

- 变更在 1 秒内生效
- 未激活的路由返回 404：`{"error": "route not found or inactive"}`
- 适合管理 100+ 路由，按需激活

## API 示例

```bash
# GET 根路径
curl http://localhost:8080/

# GET 用户列表
curl http://localhost:8080/users

# GET 指定用户（路径参数）
curl http://localhost:8080/users/123

# POST 聊天消息（配置了2秒延迟）
curl -X POST http://localhost:8080/v1/chat-messages \
  -H "Content-Type: application/json" \
  -d '{"message": "hello"}'

# 未激活的路由返回 404
curl http://localhost:8080/products
# {"error": "route not found or inactive", "method": "GET", "path": "/products"}
```

## 环境要求

- Go 1.23+
- gin-gonic/gin v1.10.0

## 许可证

MIT