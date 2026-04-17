package server

import (
	"context"
	"encoding/json"
	"fmt"
	"http-server/internal/config"
	"http-server/internal/template"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// FileCache 文件缓存
type FileCache struct {
	JSONData interface{}
	ModTime  time.Time
}

// HTTPServer HTTP服务器
type HTTPServer struct {
	Routes     *config.RoutesConfig
	RoutesMu   sync.RWMutex
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

	if err := server.LoadRoutes(); err != nil {
		return nil, err
	}

	return server, nil
}

// LoadRoutes 加载路由配置和激活的文件
func (s *HTTPServer) LoadRoutes() error {
	data, err := os.ReadFile(s.routesPath)
	if err != nil {
		return fmt.Errorf("read routes: %w", err)
	}

	var cfg config.RoutesConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse routes: %w", err)
	}

	// 解析延迟时间
	config.ParseAllDelays(cfg.Routes)

	s.RoutesMu.Lock()
	s.Routes = &cfg
	s.RoutesMu.Unlock()

	// 只加载激活的路由文件
	for _, route := range cfg.Routes {
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

// GetFile 获取文件内容（带缓存检查）
func (s *HTTPServer) GetFile(path string) (*FileCache, error) {
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

// GetRoutesModTime 获取路由配置文件的修改时间
func (s *HTTPServer) GetRoutesModTime() time.Time {
	if info, err := os.Stat(s.routesPath); err == nil {
		return info.ModTime()
	}
	return time.Time{}
}

// GetCachedFilePaths 获取所有缓存文件路径
func (s *HTTPServer) GetCachedFilePaths() []string {
	s.cacheMu.RLock()
	paths := make([]string, 0, len(s.cache))
	for p := range s.cache {
		paths = append(paths, p)
	}
	s.cacheMu.RUnlock()
	return paths
}

// Watch 监听配置和文件变化
func (s *HTTPServer) Watch(ctx context.Context) {
	routesTicker := time.NewTicker(1 * time.Second)
	filesTicker := time.NewTicker(500 * time.Millisecond)
	defer routesTicker.Stop()
	defer filesTicker.Stop()

	routesModTime := s.GetRoutesModTime()

	for {
		select {
		case <-ctx.Done():
			return
		case <-routesTicker.C:
			s.checkRoutesChange(&routesModTime)
		case <-filesTicker.C:
			s.checkFilesChange()
		}
	}
}

// checkRoutesChange 检查路由配置文件变化
func (s *HTTPServer) checkRoutesChange(modTime *time.Time) {
	info, err := os.Stat(s.routesPath)
	if err != nil {
		return
	}

	if !info.ModTime().After(*modTime) {
		return
	}

	*modTime = info.ModTime()
	fmt.Println("[HTTP] Routes changed, reloading...")
	if err := s.LoadRoutes(); err != nil {
		fmt.Printf("[HTTP] Reload error: %v\n", err)
	} else {
		fmt.Println("[HTTP] Routes reloaded successfully")
	}
}

// checkFilesChange 检查已缓存文件变化
func (s *HTTPServer) checkFilesChange() {
	paths := s.GetCachedFilePaths()
	for _, path := range paths {
		s.reloadFileIfNeeded(path)
	}
}

// reloadFileIfNeeded 如果文件已更新则重新加载
func (s *HTTPServer) reloadFileIfNeeded(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	s.cacheMu.RLock()
	cached := s.cache[path]
	s.cacheMu.RUnlock()

	if cached == nil || !info.ModTime().After(cached.ModTime) {
		return
	}

	if err := s.loadFile(path); err == nil {
		fmt.Printf("[HTTP] Reloaded: %s\n", path)
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

	s.RoutesMu.RLock()
	routes := s.Routes.Routes
	s.RoutesMu.RUnlock()

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
		// 如果配置了延迟，则等待
		if route.DelayParsed > 0 {
			time.Sleep(route.DelayParsed)
		}

		cache, err := s.GetFile(route.File)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "file not loaded",
				"file":  route.File,
			})
			return
		}

		response := template.Render(cache.JSONData, params)
		c.JSON(http.StatusOK, response)
		return
	}

	// 未找到激活的匹配路由
	c.JSON(http.StatusNotFound, gin.H{
		"error":  "route not found or inactive",
		"method": method,
		"path":   path,
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
