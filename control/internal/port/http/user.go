package httpport

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kidyme/nexus/common/httpx"
	userapp "github.com/kidyme/nexus/control/internal/application/user"
	userdomain "github.com/kidyme/nexus/control/internal/domain/user"
	usergen "github.com/kidyme/nexus/control/internal/port/http/gen/user"
)

// UserHandler 实现用户相关 HTTP 接口。
type UserHandler struct {
	service *userapp.Service
}

// NewUserHandler 创建用户 handler。
func NewUserHandler(service *userapp.Service) *UserHandler {
	return &UserHandler{service: service}
}

// registerUserRoutes 注册用户路由。
func registerUserRoutes(router gin.IRouter, handler *UserHandler) {
	usergen.RegisterHandlersWithOptions(router, handler, usergen.GinServerOptions{
		ErrorHandler: openAPIErrorHandler,
	})
}

// GetUser 处理 GET /api/users/{user_id}。
func (h *UserHandler) GetUser(c *gin.Context, userID string) {
	entity, err := h.service.Find(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	resp, err := toUserResponse(*entity)
	if err != nil {
		writeError(c, err)
		return
	}
	httpx.OK(c, resp)
}

// CreateUsers 处理 POST /api/users。
func (h *UserHandler) CreateUsers(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	single, many, err := decodeSingleOrSlice[usergen.CreateUserRequest](payload)
	if err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	if single != nil {
		entity, err := toCreateUser(*single)
		if err != nil {
			writeError(c, err)
			return
		}
		entity.UserID = strings.TrimSpace(entity.UserID)
		if err := h.service.Create(c.Request.Context(), entity); err != nil {
			writeError(c, err)
			return
		}

		created, err := h.service.Find(c.Request.Context(), entity.UserID)
		if err != nil {
			writeError(c, err)
			return
		}
		resp, err := toUserResponse(*created)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", "/api/users/"+url.PathEscape(created.UserID))
		httpx.Created(c, resp)
		return
	}

	users := make([]userdomain.User, 0, len(many))
	for _, item := range many {
		entity, err := toCreateUser(item)
		if err != nil {
			writeError(c, err)
			return
		}
		users = append(users, entity)
	}
	if err := h.service.CreateBatch(c.Request.Context(), users); err != nil {
		writeError(c, err)
		return
	}

	resp := make([]usergen.User, 0, len(users))
	for _, item := range users {
		out, err := toUserResponse(item)
		if err != nil {
			writeError(c, err)
			return
		}
		resp = append(resp, out)
	}
	httpx.Created(c, resp)
}

// ListUsers 处理 GET /api/users。
func (h *UserHandler) ListUsers(c *gin.Context, params usergen.ListUsersParams) {
	page, size, err := normalizePagination(params.Page, params.Size)
	if err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}
	users, total, err := h.service.ListPage(c.Request.Context(), page, size)
	if err != nil {
		writeError(c, err)
		return
	}

	result := make([]usergen.User, 0, len(users))
	for _, item := range users {
		out, err := toUserResponse(item)
		if err != nil {
			writeError(c, err)
			return
		}
		result = append(result, out)
	}
	httpx.OK(c, usergen.UserPageData{
		List:  result,
		Page:  page,
		Size:  size,
		Total: total,
	})
}

// PatchUser 处理 PATCH /api/users/{user_id}。
func (h *UserHandler) PatchUser(c *gin.Context, userID string) {
	var req usergen.PatchUserJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	current, err := h.service.Find(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
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

	updated, err := h.service.Find(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	resp, err := toUserResponse(*updated)
	if err != nil {
		writeError(c, err)
		return
	}
	httpx.OK(c, resp)
}

// PatchUsers 处理 PATCH /api/users。
func (h *UserHandler) PatchUsers(c *gin.Context) {
	var req usergen.PatchUsersJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	result := make([]userdomain.User, 0, len(req))
	for _, item := range req {
		current, err := h.service.Find(c.Request.Context(), item.UserId)
		if err != nil {
			writeError(c, err)
			return
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

	resp := make([]usergen.User, 0, len(result))
	for _, item := range result {
		out, err := toUserResponse(item)
		if err != nil {
			writeError(c, err)
			return
		}
		resp = append(resp, out)
	}
	httpx.OK(c, resp)
}

// DeleteUsers 处理 DELETE /api/users。
func (h *UserHandler) DeleteUsers(c *gin.Context) {
	var req usergen.DeleteUsersJSONRequestBody
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

// DeleteUser 处理 DELETE /api/users/{user_id}。
func (h *UserHandler) DeleteUser(c *gin.Context, userID string) {
	if err := h.service.Delete(c.Request.Context(), userID); err != nil {
		writeError(c, err)
		return
	}
	httpx.NoContent(c)
}

func toCreateUser(req usergen.CreateUserRequest) (userdomain.User, error) {
	labels, err := marshalLabels(req.Labels)
	if err != nil {
		return userdomain.User{}, err
	}

	entity := userdomain.User{
		UserID: req.UserId,
		Labels: labels,
	}
	if req.Comment != nil {
		entity.Comment = *req.Comment
	}
	return entity, nil
}

func toUserResponse(src userdomain.User) (usergen.User, error) {
	labels, err := unmarshalLabels(src.Labels)
	if err != nil {
		return usergen.User{}, err
	}
	return usergen.User{
		UserId:  src.UserID,
		Labels:  labels,
		Comment: stringPtr(src.Comment),
	}, nil
}
