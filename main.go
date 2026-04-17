package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// RoutesConfig 路由配置
type RoutesConfig struct {
	Routes []RouteConfig `json:"routes"`
}

// RouteConfig 单个路由配置
type RouteConfig struct {
	Active bool   `json:"active"`
	Method string `json:"method"`
	Path   string `json:"path"`
	File   string `json:"file"`
}

// FileCache 文件缓存
type FileCache struct {
	JSONData interface{}
	ModTime  time.Time
}

// HTTPServer HTTP服务器
type HTTPServer struct {
	routes     *RoutesConfig
	routesMu   sync.RWMutex
	cache      map[string]*FileCache
	cacheMu    sync.RWMutex
	routesPath string
}

// NewHTTPServer 创建HTTP服务器
func NewHTTPServer(routesPath string) (*HTTPServer, error) {
	server := &HTTPServer{
		cache:      make(map[string]*FileCache),
		routesPath: routesPath,
	}

	if err := server.loadRoutes(); err != nil {
		return nil, err
	}

	return server, nil
}

// loadRoutes 加载路由配置和激活的文件
func (s *HTTPServer) loadRoutes() error {
	data, err := os.ReadFile(s.routesPath)
	if err != nil {
		return fmt.Errorf("read routes: %w", err)
	}

	var config RoutesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parse routes: %w", err)
	}

	s.routesMu.Lock()
	s.routes = &config
	s.routesMu.Unlock()

	// 只加载激活的路由文件
	for _, route := range config.Routes {
		if !route.Active {
			// 未激活的从缓存移除
			s.cacheMu.Lock()
			delete(s.cache, route.File)
			s.cacheMu.Unlock()
			continue
		}

		if err := s.loadFile(route.File); err != nil {
			fmt.Printf("[Warn] Failed to load %s: %v\n", route.File, err)
		} else {
			fmt.Printf("[HTTP] Loaded: %s (active)\n", route.Path)
		}
	}

	return nil
}

// loadFile 加载文件到缓存
func (s *HTTPServer) loadFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return err
	}

	s.cacheMu.Lock()
	s.cache[path] = &FileCache{
		JSONData: jsonData,
		ModTime:  info.ModTime(),
	}
	s.cacheMu.Unlock()

	return nil
}

// getFile 获取文件内容（带缓存检查）
func (s *HTTPServer) getFile(path string) (*FileCache, error) {
	s.cacheMu.RLock()
	cached := s.cache[path]
	s.cacheMu.RUnlock()

	if cached == nil {
		return nil, fmt.Errorf("not cached")
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// 文件已更新，重新加载
	if info.ModTime().After(cached.ModTime) {
		if err := s.loadFile(path); err != nil {
			return cached, nil // 返回旧缓存
		}
		s.cacheMu.RLock()
		cached = s.cache[path]
		s.cacheMu.RUnlock()
	}

	return cached, nil
}

// Watch 监听配置和文件变化
func (s *HTTPServer) Watch(ctx context.Context) {
	routesTicker := time.NewTicker(1 * time.Second)
	filesTicker := time.NewTicker(500 * time.Millisecond)
	defer routesTicker.Stop()
	defer filesTicker.Stop()

	var routesModTime time.Time
	if info, err := os.Stat(s.routesPath); err == nil {
		routesModTime = info.ModTime()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-routesTicker.C:
			// 检查 routes.json 变化
			if info, err := os.Stat(s.routesPath); err == nil {
				if info.ModTime().After(routesModTime) {
					routesModTime = info.ModTime()
					fmt.Println("[HTTP] Routes changed, reloading...")
					if err := s.loadRoutes(); err != nil {
						fmt.Printf("[HTTP] Reload error: %v\n", err)
					} else {
						fmt.Println("[HTTP] Routes reloaded successfully")
					}
				}
			}
		case <-filesTicker.C:
			// 检查已缓存文件变化
			s.cacheMu.RLock()
			paths := make([]string, 0, len(s.cache))
			for p := range s.cache {
				paths = append(paths, p)
			}
			s.cacheMu.RUnlock()

			for _, path := range paths {
				info, err := os.Stat(path)
				if err != nil {
					continue
				}
				s.cacheMu.RLock()
				cached := s.cache[path]
				s.cacheMu.RUnlock()

				if cached != nil && info.ModTime().After(cached.ModTime) {
					if err := s.loadFile(path); err == nil {
						fmt.Printf("[HTTP] Reloaded: %s\n", path)
					}
				}
			}
		}
	}
}

// SetupRoutes 设置路由
func (s *HTTPServer) SetupRoutes(r *gin.Engine) {
	r.NoRoute(s.handleRequest)
}

// handleRequest 处理请求
func (s *HTTPServer) handleRequest(c *gin.Context) {
	method := c.Request.Method
	path := c.Request.URL.Path

	s.routesMu.RLock()
	routes := s.routes.Routes
	s.routesMu.RUnlock()

	// 只匹配激活的路由
	for _, route := range routes {
		if !route.Active {
			continue
		}
		if route.Method != method {
			continue
		}

		params, ok := matchPath(route.Path, path)
		if !ok {
			continue
		}

		// 找到激活的匹配路由
		cache, err := s.getFile(route.File)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "file not loaded",
				"file": route.File,
			})
			return
		}

		response := renderTemplate(cache.JSONData, params)
		c.JSON(http.StatusOK, response)
		return
	}

	// 未找到激活的匹配路由
	c.JSON(http.StatusNotFound, gin.H{
		"error": "route not found or inactive",
		"method": method,
		"path": path,
	})
}

// matchPath 路径匹配
func matchPath(pattern, path string) (map[string]string, bool) {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i := 0; i < len(patternParts); i++ {
		pp := patternParts[i]
		ep := pathParts[i]

		if strings.HasPrefix(pp, ":") {
			params[strings.TrimPrefix(pp, ":")] = ep
		} else if pp != ep {
			return nil, false
		}
	}

	return params, true
}

// renderTemplate 渲染模板
func renderTemplate(data interface{}, params map[string]string) interface{} {
	switch v := data.(type) {
	case string:
		result := v
		for k, val := range params {
			result = strings.ReplaceAll(result, "{{.PathParams."+k+"}}", val)
		}
		return result

	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = renderTemplate(val, params)
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = renderTemplate(val, params)
		}
		return result

	default:
		return data
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	routesPath := "routes.json"
	if len(os.Args) > 1 {
		routesPath = os.Args[1]
	}

	server, err := NewHTTPServer(routesPath)
	if err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		os.Exit(1)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	server.SetupRoutes(r)

	ctx, cancel := context.WithCancel(context.Background())
	go server.Watch(ctx)

	// 统计激活路由
	server.routesMu.RLock()
	activeCount := 0
	for _, route := range server.routes.Routes {
		if route.Active {
			activeCount++
		}
	}
	server.routesMu.RUnlock()

	fmt.Printf("[HTTP] Server running on :8080\n")
	fmt.Printf("[HTTP] Routes: %s\n", routesPath)
	fmt.Printf("[HTTP] Active routes: %d\n", activeCount)

	// 使用 http.Server 实现优雅关闭
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 启动服务器（非阻塞）
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\n[HTTP] Shutting down...")

	// 优雅关闭，最多等待5秒
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	cancel() // 停止 watch

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	}
	fmt.Println("[HTTP] Server stopped")
}