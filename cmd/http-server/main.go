package main

import (
	"context"
	"fmt"
	"http-server/internal/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	routesPath := "routes.json"
	if len(os.Args) > 1 {
		routesPath = os.Args[1]
	}

	httpServer, err := server.NewHTTPServer(routesPath)
	if err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		os.Exit(1)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	httpServer.SetupRoutes(r)

	ctx, cancel := context.WithCancel(context.Background())
	go httpServer.Watch(ctx)

	// 统计激活路由
	httpServer.RoutesMu.RLock()
	activeCount := 0
	for _, route := range httpServer.Routes.Routes {
		if route.Active {
			activeCount++
		}
	}
	httpServer.RoutesMu.RUnlock()

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
