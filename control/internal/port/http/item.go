package httpport

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kidyme/nexus/common/httpx"
	itemapp "github.com/kidyme/nexus/control/internal/application/item"
	itemdomain "github.com/kidyme/nexus/control/internal/domain/item"
	itemgen "github.com/kidyme/nexus/control/internal/port/http/gen/item"
)

// ItemHandler 实现物品相关 HTTP 接口。
type ItemHandler struct {
	service *itemapp.Service
}

// NewItemHandler 创建物品 handler。
func NewItemHandler(service *itemapp.Service) *ItemHandler {
	return &ItemHandler{service: service}
}

// registerItemRoutes 注册物品路由。
func registerItemRoutes(router gin.IRouter, handler *ItemHandler) {
	itemgen.RegisterHandlersWithOptions(router, handler, itemgen.GinServerOptions{
		ErrorHandler: openAPIErrorHandler,
	})
}

// GetItem 处理 GET /api/items/{item_id}。
func (h *ItemHandler) GetItem(c *gin.Context, itemID string) {
	entity, err := h.service.Find(c.Request.Context(), itemID)
	if err != nil {
		writeError(c, err)
		return
	}
	resp, err := toItemResponse(*entity)
	if err != nil {
		writeError(c, err)
		return
	}
	httpx.OK(c, resp)
}

// CreateItems 处理 POST /api/items。
func (h *ItemHandler) CreateItems(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	single, many, err := decodeSingleOrSlice[itemgen.CreateItemRequest](payload)
	if err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	if single != nil {
		entity, err := toCreateItem(*single)
		if err != nil {
			writeError(c, err)
			return
		}
		entity.ItemID = strings.TrimSpace(entity.ItemID)
		if err := h.service.Create(c.Request.Context(), entity); err != nil {
			writeError(c, err)
			return
		}

		created, err := h.service.Find(c.Request.Context(), entity.ItemID)
		if err != nil {
			writeError(c, err)
			return
		}
		resp, err := toItemResponse(*created)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", "/api/items/"+url.PathEscape(created.ItemID))
		httpx.Created(c, resp)
		return
	}

	items := make([]itemdomain.Item, 0, len(many))
	for _, item := range many {
		entity, err := toCreateItem(item)
		if err != nil {
			writeError(c, err)
			return
		}
		items = append(items, entity)
	}
	if err := h.service.CreateBatch(c.Request.Context(), items); err != nil {
		writeError(c, err)
		return
	}

	resp := make([]itemgen.Item, 0, len(items))
	for _, item := range items {
		out, err := toItemResponse(item)
		if err != nil {
			writeError(c, err)
			return
		}
		resp = append(resp, out)
	}
	httpx.Created(c, resp)
}

// ListItems 处理 GET /api/items。
func (h *ItemHandler) ListItems(c *gin.Context, params itemgen.ListItemsParams) {
	page, size, err := normalizePagination(params.Page, params.Size)
	if err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}
	items, total, err := h.service.ListPage(c.Request.Context(), page, size)
	if err != nil {
		writeError(c, err)
		return
	}

	result := make([]itemgen.Item, 0, len(items))
	for _, item := range items {
		out, err := toItemResponse(item)
		if err != nil {
			writeError(c, err)
			return
		}
		result = append(result, out)
	}
	httpx.OK(c, itemgen.ItemPageData{
		List:  result,
		Page:  page,
		Size:  size,
		Total: total,
	})
}

// PatchItem 处理 PATCH /api/items/{item_id}。
func (h *ItemHandler) PatchItem(c *gin.Context, itemID string) {
	var req itemgen.PatchItemJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	current, err := h.service.Find(c.Request.Context(), itemID)
	if err != nil {
		writeError(c, err)
		return
	}
	if req.IsHidden != nil {
		current.IsHidden = *req.IsHidden
	}
	if req.Categories != nil {
		current.Categories = cloneStrings(req.Categories)
	}
	if req.Timestamp != nil {
		current.Timestamp = *req.Timestamp
	}
	if req.Comment != nil {
		current.Comment = *req.Comment
	}
	if req.Labels != nil {
		current.Labels, err = marshalLabels(req.Labels)
		if err != nil {
			writeError(c, err)
			return
		}
	}
	if err := h.service.Update(c.Request.Context(), *current); err != nil {
		writeError(c, err)
		return
	}

	updated, err := h.service.Find(c.Request.Context(), itemID)
	if err != nil {
		writeError(c, err)
		return
	}
	resp, err := toItemResponse(*updated)
	if err != nil {
		writeError(c, err)
		return
	}
	httpx.OK(c, resp)
}

// PatchItems 处理 PATCH /api/items。
func (h *ItemHandler) PatchItems(c *gin.Context) {
	var req itemgen.PatchItemsJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	result := make([]itemdomain.Item, 0, len(req))
	for _, item := range req {
		current, err := h.service.Find(c.Request.Context(), item.ItemId)
		if err != nil {
			writeError(c, err)
			return
		}
		if item.IsHidden != nil {
			current.IsHidden = *item.IsHidden
		}
		if item.Categories != nil {
			current.Categories = cloneStrings(item.Categories)
		}
		if item.Timestamp != nil {
			current.Timestamp = *item.Timestamp
		}
		if item.Comment != nil {
			current.Comment = *item.Comment
		}
		if item.Labels != nil {
			current.Labels, err = marshalLabels(item.Labels)
			if err != nil {
				writeError(c, err)
				return
			}
		}
		result = append(result, *current)
	}

	if err := h.service.UpdateBatch(c.Request.Context(), result); err != nil {
		writeError(c, err)
		return
	}

	resp := make([]itemgen.Item, 0, len(result))
	for _, item := range result {
		out, err := toItemResponse(item)
		if err != nil {
			writeError(c, err)
			return
		}
		resp = append(resp, out)
	}
	httpx.OK(c, resp)
}

// DeleteItems 处理 DELETE /api/items。
func (h *ItemHandler) DeleteItems(c *gin.Context) {
	var req itemgen.DeleteItemsJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}
	if err := h.service.DeleteBatch(c.Request.Context(), []string(req)); err != nil {
		writeError(c, err)
		return
	}
	httpx.NoContent(c)
}

// DeleteItem 处理 DELETE /api/items/{item_id}。
func (h *ItemHandler) DeleteItem(c *gin.Context, itemID string) {
	if err := h.service.Delete(c.Request.Context(), itemID); err != nil {
		writeError(c, err)
		return
	}
	httpx.NoContent(c)
}

func toItemResponse(src itemdomain.Item) (itemgen.Item, error) {
	labels, err := unmarshalLabels(src.Labels)
	if err != nil {
		return itemgen.Item{}, err
	}
	return itemgen.Item{
		ItemId:     src.ItemID,
		IsHidden:   src.IsHidden,
		Categories: src.Categories,
		Timestamp:  src.Timestamp,
		Labels:     labels,
		Comment:    stringPtr(src.Comment),
	}, nil
}

func cloneStrings(src *[]string) []string {
	if src == nil {
		return nil
	}
	out := make([]string, len(*src))
	copy(out, *src)
	return out
}

func toCreateItem(req itemgen.CreateItemRequest) (itemdomain.Item, error) {
	labels, err := marshalLabels(req.Labels)
	if err != nil {
		return itemdomain.Item{}, err
	}

	entity := itemdomain.Item{
		ItemID:     req.ItemId,
		IsHidden:   req.IsHidden != nil && *req.IsHidden,
		Categories: cloneStrings(req.Categories),
		Labels:     labels,
	}
	if req.Timestamp != nil {
		entity.Timestamp = *req.Timestamp
	}
	if req.Comment != nil {
		entity.Comment = *req.Comment
	}
	return entity, nil
}
