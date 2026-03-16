// Package httpport 提供 control 服务的 HTTP 端口适配层。
package httpport

import (
	"github.com/gin-gonic/gin"
	"github.com/kidyme/nexus/common/httpx"
)

// Handlers 汇总路由所需的 HTTP handler。
type Handlers struct {
	Common   *CommonHandler
	Node     *NodeHandler
	User     *UserHandler
	Item     *ItemHandler
	Feedback *FeedbackHandler
}

// NewRouter 创建 control 的 HTTP 路由。
func NewRouter(handlers Handlers) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	if handlers.Common != nil {
		registerCommonRoutes(router, handlers.Common)
	}
	if handlers.Node != nil {
		registerNodeRoutes(router, handlers.Node)
	}
	if handlers.User != nil {
		registerUserRoutes(router, handlers.User)
	}
	if handlers.Item != nil {
		registerItemRoutes(router, handlers.Item)
	}
	if handlers.Feedback != nil {
		registerFeedbackRoutes(router, handlers.Feedback)
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
