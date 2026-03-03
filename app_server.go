package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// AppServer 应用服务器
type AppServer struct {
	service    *HeyboxService
	mcpServer  *mcp.Server
	httpServer *http.Server
}

// NewAppServer 创建应用服务器
func NewAppServer(service *HeyboxService) *AppServer {
	mcpServer := InitMCPServer(service)

	return &AppServer{
		service:   service,
		mcpServer: mcpServer,
	}
}

// Start 启动服务器
func (s *AppServer) Start(port string) error {
	// 创建 HTTP 路由
	mux := http.NewServeMux()

	// 注册路由
	s.registerRoutes(mux)

	// 创建 HTTP 服务器
	s.httpServer = &http.Server{
		Addr:         port,
		Handler:      s.loggingMiddleware(s.recoveryMiddleware(mux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logrus.Infof("🚀 heybox-mcp 服务启动，监听端口 %s", port)
	logrus.Infof("📋 MCP 端点: http://localhost%s/mcp", port)
	logrus.Infof("❤️  健康检查: http://localhost%s/health", port)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("服务器启动失败: %w", err)
	}

	return nil
}

// Shutdown 关闭服务器
func (s *AppServer) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// registerRoutes 注册路由
func (s *AppServer) registerRoutes(mux *http.ServeMux) {
	// 健康检查
	mux.HandleFunc("/health", s.handleHealth)

	// MCP 端点 - 使用 StreamableHTTP 传输
	mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)
	mux.Handle("/mcp", mcpHandler)
}

// handleHealth 健康检查
func (s *AppServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"heybox-mcp","version":"1.0.0"}`))
}
