---
name: add-feature
description: 为HTTP模拟服务器添加新功能。用于实现新特性、添加配置选项或扩展现有能力。
---

为 HTTP 模拟服务器项目添加新功能。

## 当前架构

服务器是一个高性能 HTTP 模拟服务器，具有：
- 通过 `routes.json` 控制路由激活状态
- 配置和数据文件热加载
- 文件缓存，支持并发安全访问
- 路径参数支持（`:id` 语法）
- 可选的响应延迟模拟

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
├── routes.json               # 路由配置
└── CLAUDE.md
```

## 关键文件职责

| 包/文件 | 职责 | 主要内容 |
|---------|------|----------|
| `internal/config` | 配置定义和解析 | RoutesConfig, RouteConfig, ParseDelay |
| `internal/server` | 服务器核心 | HTTPServer, NewHTTPServer, LoadRoutes, Watch, SetupRoutes, handleRequest |
| `internal/template` | 响应模板渲染 | Render |
| `cmd/http-server` | 程序入口 | main 函数 |

## 实现步骤

1. **理解需求** - 明确需要什么功能
2. **确定位置** - 根据职责选择正确的包
   - 新配置字段 → `internal/config/config.go`
   - 服务器行为、请求处理、监听 → `internal/server/server.go`
   - 响应模板处理 → `internal/template/template.go`
3. **更新 RouteConfig** - 如需新配置字段，在 config.go 添加
4. **实现逻辑** - 在对应包中添加功能
5. **更新文档** - 在 CLAUDE.md 和 README.md 中记录新功能
6. **测试验证** - 构建并运行服务器测试

## RouteConfig 扩展模式

在 `internal/config/config.go` 中添加新字段：

```go
type RouteConfig struct {
    Active      bool          `json:"active"`
    Method      string        `json:"method"`
    Path        string        `json:"path"`
    File        string        `json:"file"`
    Delay       string        `json:"delay"`
    DelayParsed time.Duration
    // 新字段示例：
    // Headers     map[string]string `json:"headers"`  // 自定义响应头
}
```

需要解析的字段在 config.go 的 `ParseDelay` 附近添加解析逻辑。

## 配置使用模式

在 `internal/server/server.go` 的 `handleRequest` 中使用新配置：

```go
// 在路由匹配成功后使用
if route.NewConfigField != "" {
    // 应用配置
}
```

## 测试验证

```bash
# 构建验证
go build -o http-server.exe ./cmd/http-server

# 交叉编译所有平台（输出到 release/ 目录）
make build-all

# 运行服务器
go run ./cmd/http-server

# 测试请求
curl http://localhost:8080/users
curl -X POST http://localhost:8080/v1/chat-messages -d "{}"
```

> **注意**：项目采用标准 Go 项目结构，入口在 `cmd/http-server/`，运行命令必须指定路径。

## 开发规范

- 保持热加载能力（变更在 1 秒内生效）
- 使用适当的互斥锁保证并发安全
- 按职责选择正确的包，不要混放
- 导出的字段/函数首字母大写（如 `RoutesMu`、`LoadRoutes`）
- 遵循现有代码模式和风格
- 在 CLAUDE.md 和 README.md 中更新新配置字段说明
- **运行命令**：`go run ./cmd/http-server`，不是 `go run main.go`