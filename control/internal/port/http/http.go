// Package httpport provides HTTP ports for the control service.
package httpport

import (
	"github.com/gin-gonic/gin"
	"github.com/kidyme/nexus/common/httpx"
)

// Handlers groups HTTP handlers required by the router.
type Handlers struct {
	Common *CommonHandler
	Node   *NodeHandler
}

// NewRouter creates the HTTP router for control.
func NewRouter(handlers Handlers) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	if handlers.Common != nil {
		registerCommonRoutes(router, handlers.Common)
	}
	if handlers.Node != nil {
		registerNodeRoutes(router, handlers.Node)
	}
	return router
}

func openAPIErrorHandler(c *gin.Context, err error, statusCode int) {
	errno := httpx.ErrnoInternal
	switch statusCode {
	case 400:
		errno = httpx.ErrnoBadRequest
	case 404:
		errno = httpx.ErrnoNotFound
	}
	httpx.Fail(c, statusCode, errno, err.Error())
}
