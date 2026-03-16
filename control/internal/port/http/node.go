package httpport

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kidyme/nexus/common/httpx"
	feedbackapp "github.com/kidyme/nexus/control/internal/application/feedback"
	itemapp "github.com/kidyme/nexus/control/internal/application/item"
	"github.com/kidyme/nexus/common/registry"
	nodeapp "github.com/kidyme/nexus/control/internal/application/node"
	userapp "github.com/kidyme/nexus/control/internal/application/user"
	feedbackdomain "github.com/kidyme/nexus/control/internal/domain/feedback"
	itemdomain "github.com/kidyme/nexus/control/internal/domain/item"
	nodedomain "github.com/kidyme/nexus/control/internal/domain/node"
	userdomain "github.com/kidyme/nexus/control/internal/domain/user"
	nodegen "github.com/kidyme/nexus/control/internal/port/http/gen/node"
)

// NodeHandler 实现生成的 OpenAPI 服务端接口。
type NodeHandler struct {
	service *nodeapp.Service
}

// registerNodeRoutes 注册 node 相关 HTTP 路由。
func registerNodeRoutes(router gin.IRouter, handler *NodeHandler) {
	nodegen.RegisterHandlersWithOptions(router, handler, nodegen.GinServerOptions{
		ErrorHandler: openAPIErrorHandler,
	})
}

// NewNodeHandler 创建 node handler。
func NewNodeHandler(nodeService *nodeapp.Service) *NodeHandler {
	return &NodeHandler{service: nodeService}
}

// ListNodes 处理 GET /api/meta/nodes。
func (h *NodeHandler) ListNodes(c *gin.Context, params nodegen.ListNodesParams) {
	var (
		nodes []nodedomain.Node
		err   error
	)
	if params.Service != nil && strings.TrimSpace(*params.Service) != "" {
		nodes, err = h.service.ListByService(c.Request.Context(), *params.Service)
	} else {
		nodes, err = h.service.List(c.Request.Context())
	}
	if err != nil {
		writeError(c, err)
		return
	}

	result := make([]nodegen.Node, 0, len(nodes))
	for _, item := range nodes {
		result = append(result, toNodeResponse(item))
	}
	httpx.OK(c, result)
}

// GetNode 处理 GET /api/meta/nodes/{node_id}。
func (h *NodeHandler) GetNode(c *gin.Context, nodeID string) {
	item, err := h.service.Find(c.Request.Context(), nodeID)
	if err != nil {
		writeError(c, err)
		return
	}
	httpx.OK(c, toNodeResponse(*item))
}

func toNodeResponse(src nodedomain.Node) nodegen.Node {
	return nodegen.Node{
		NodeId:      src.NodeID,
		ServiceName: src.ServiceName,
		Endpoint:    src.Endpoint,
		Status:      src.Status,
		Version:     src.Version,
		HeartbeatAt: src.HeartbeatAt,
	}
}

func writeError(c *gin.Context, err error) {
	status := 500
	errno := httpx.ErrnoInternal
	switch {
	case errors.Is(err, nodeapp.ErrServiceNameRequired), errors.Is(err, nodeapp.ErrNodeIDRequired):
		status = 400
		errno = httpx.ErrnoBadRequest
	case errors.Is(err, userapp.ErrInvalidPage),
		errors.Is(err, userapp.ErrInvalidSize),
		errors.Is(err, itemapp.ErrInvalidPage),
		errors.Is(err, itemapp.ErrInvalidSize),
		errors.Is(err, feedbackapp.ErrInvalidPage),
		errors.Is(err, feedbackapp.ErrInvalidSize),
		errors.Is(err, errPageMustBePositive),
		errors.Is(err, errSizeMustBePositive):
		status = 400
		errno = httpx.ErrnoBadRequest
	case errors.Is(err, userdomain.ErrUserIDRequired),
		errors.Is(err, itemdomain.ErrItemIDRequired),
		errors.Is(err, feedbackdomain.ErrFeedbackTypeRequired),
		errors.Is(err, feedbackdomain.ErrUserIDRequired),
		errors.Is(err, feedbackdomain.ErrItemIDRequired):
		status = 400
		errno = httpx.ErrnoBadRequest
	case errors.Is(err, userdomain.ErrUserNotFound),
		errors.Is(err, itemdomain.ErrItemNotFound),
		errors.Is(err, feedbackdomain.ErrFeedbackNotFound):
		status = 404
		errno = httpx.ErrnoNotFound
	case errors.Is(err, registry.ErrNodeNotFound):
		status = 404
		errno = httpx.ErrnoNotFound
	}
	httpx.Fail(c, status, errno, err.Error())
}
