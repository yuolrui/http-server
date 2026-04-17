# HTTP Mock Server

基于 Go 和 Gin 构建的高性能 HTTP Mock 服务器，支持路由激活控制。

## 特性

- **路由激活控制**：通过 `active` 字段启用/禁用路由，无需重启服务
- **热重载**：routes.json 变更 1 秒内生效，数据文件变更 500ms 内生效
- **路径参数**：支持 `:id` 风格的路径参数及模板替换
- **内存高效**：仅加载激活的路由文件到内存
- **并发安全**：使用 RWMutex 保护并发访问
- **优雅关闭**：支持 5 秒超时的优雅关闭

## 快速开始

```bash
# 使用默认 routes.json 运行
go run main.go

# 使用自定义路由配置运行
go run main.go path/to/routes.json

# 编译
go build -o http-server.exe
```

服务器默认监听 `:8080` 端口。

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

# POST 聊天消息
curl -X POST http://localhost:8080/v1/chat-messages \
  -H "Content-Type: application/json" \
  -d '{"message": "hello"}'

# 未激活的路由返回 404
curl http://localhost:8080/products
# {"error": "route not found or inactive", "method": "GET", "path": "/products"}
```

## 架构

```
┌─────────────────────────────────────────────┐
│                  main.go                     │
├─────────────────────────────────────────────┤
│  HTTPServer                                  │
│  ├── routes (RoutesConfig)                   │
│  │   └── routes.json (热重载 1s)            │
│  ├── cache (map[string]*FileCache)          │
│  │   └── *.json 文件 (热重载 500ms)         │
│  └── Watch() - 后台文件监听器               │
├─────────────────────────────────────────────┤
│  Gin Engine                                  │
│  └── NoRoute handler → 匹配激活的路由       │
└─────────────────────────────────────────────┘
```

## 环境要求

- Go 1.23+
- gin-gonic/gin v1.10.0

## 许可证

MIT