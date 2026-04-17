package config

import (
	"fmt"
	"time"
)

// RoutesConfig 路由配置
type RoutesConfig struct {
	Routes []RouteConfig `json:"routes"`
}

// RouteConfig 单个路由配置
type RouteConfig struct {
	Active      bool          `json:"active"`
	Method      string        `json:"method"`
	Path        string        `json:"path"`
	File        string        `json:"file"`
	Delay       string        `json:"delay"` // 延迟时间字符串，如 "500ms", "2s"
	DelayParsed time.Duration // 解析后的延迟时间（导出供其他包访问）
}

// ParseDelay 解析延迟时间配置
func (r *RouteConfig) ParseDelay() {
	if r.Delay == "" {
		return
	}

	d, err := time.ParseDuration(r.Delay)
	if err != nil {
		fmt.Printf("[Warn] Invalid delay '%s' for %s: %v\n", r.Delay, r.Path, err)
		return
	}

	r.DelayParsed = d
}

// ParseAllDelays 解析所有路由的延迟配置
func ParseAllDelays(routes []RouteConfig) {
	for i := range routes {
		routes[i].ParseDelay()
	}
}
