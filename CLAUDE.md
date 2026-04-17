# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Run the server (listens on :8080)
go run main.go

# Run with custom routes config
go run main.go path/to/routes.json

# Build binary
go build -o http-server.exe
```

## Architecture

High-performance HTTP mock server with route activation control.

### Route Configuration (routes.json)

```json
{
  "routes": [
    {
      "active": true,
      "method": "POST",
      "path": "/v1/chat-messages",
      "file": "data/dify/chat-messages.json"
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

## Dependencies

- Go 1.23+
- gin-gonic/gin v1.10.0