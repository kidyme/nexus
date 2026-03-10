package httpport

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kidyme/nexus/common/httpx"
	"github.com/kidyme/nexus/common/registry"
	nodeapp "github.com/kidyme/nexus/control/internal/control/application/node"
	nodedomain "github.com/kidyme/nexus/control/internal/control/domain/node"
	nodegen "github.com/kidyme/nexus/control/internal/control/port/http/gen/node"
)

// NodeHandler implements the generated OpenAPI server interface.
type NodeHandler struct {
	service *nodeapp.Service
}

// registerNodeRoutes registers node-related HTTP routes.
func registerNodeRoutes(router gin.IRouter, handler *NodeHandler) {
	nodegen.RegisterHandlersWithOptions(router, handler, nodegen.GinServerOptions{
		ErrorHandler: openAPIErrorHandler,
	})
}

// NewNodeHandler creates a node handler.
func NewNodeHandler(nodeService *nodeapp.Service) *NodeHandler {
	return &NodeHandler{service: nodeService}
}

// ListNodes handles GET /api/meta/nodes.
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

// GetNode handles GET /api/meta/nodes/{node_id}.
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
	case errors.Is(err, registry.ErrNodeNotFound):
		status = 404
		errno = httpx.ErrnoNotFound
	}
	httpx.Fail(c, status, errno, err.Error())
}
