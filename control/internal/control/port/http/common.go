package httpport

import (
	"github.com/gin-gonic/gin"
	"github.com/kidyme/nexus/common/httpx"
	commongen "github.com/kidyme/nexus/control/internal/control/port/http/gen/common"
)

// CommonHandler implements common HTTP endpoints for control.
type CommonHandler struct{}

// NewCommonHandler creates a common handler.
func NewCommonHandler() *CommonHandler {
	return &CommonHandler{}
}

// registerCommonRoutes registers shared HTTP routes.
func registerCommonRoutes(router gin.IRouter, handler *CommonHandler) {
	commongen.RegisterHandlersWithOptions(router, handler, commongen.GinServerOptions{
		ErrorHandler: openAPIErrorHandler,
	})
}

// Healthz handles GET /healthz.
func (h *CommonHandler) Healthz(c *gin.Context) {
	httpx.OK(c, gin.H{"status": "ok"})
}
