package httpport

import (
	"github.com/gin-gonic/gin"
	"github.com/kidyme/nexus/common/httpx"
	commongen "github.com/kidyme/nexus/control/internal/port/http/gen/common"
)

// CommonHandler 实现 control 的通用 HTTP 接口。
type CommonHandler struct{}

// NewCommonHandler 创建通用 handler。
func NewCommonHandler() *CommonHandler {
	return &CommonHandler{}
}

// registerCommonRoutes 注册共享 HTTP 路由。
func registerCommonRoutes(router gin.IRouter, handler *CommonHandler) {
	commongen.RegisterHandlersWithOptions(router, handler, commongen.GinServerOptions{
		ErrorHandler: openAPIErrorHandler,
	})
}

// Healthz 处理 GET /healthz。
func (h *CommonHandler) Healthz(c *gin.Context) {
	httpx.OK(c, gin.H{"status": "ok"})
}
