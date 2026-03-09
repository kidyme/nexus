package server

import (
	"context"
	"net/http"
)

// HTTPServer 封装 HTTP 服务，便于多服务复用。
type HTTPServer struct {
	Addr   string
	Server *http.Server
}

// NewHTTPServer 创建 HTTP 服务。
func NewHTTPServer(addr string, handler http.Handler) *HTTPServer {
	return &HTTPServer{
		Addr: addr,
		Server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

// Start 启动监听。
func (s *HTTPServer) Start() error {
	return s.Server.ListenAndServe()
}

// Shutdown 优雅关闭。
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
