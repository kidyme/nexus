package httpport

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kidyme/nexus/common/httpx"
	feedbackapp "github.com/kidyme/nexus/control/internal/application/feedback"
	feedbackdomain "github.com/kidyme/nexus/control/internal/domain/feedback"
	feedbackgen "github.com/kidyme/nexus/control/internal/port/http/gen/feedback"
)

// FeedbackHandler 实现反馈相关 HTTP 接口。
type FeedbackHandler struct {
	service *feedbackapp.Service
}

// NewFeedbackHandler 创建反馈 handler。
func NewFeedbackHandler(service *feedbackapp.Service) *FeedbackHandler {
	return &FeedbackHandler{service: service}
}

// registerFeedbackRoutes 注册反馈路由。
func registerFeedbackRoutes(router gin.IRouter, handler *FeedbackHandler) {
	feedbackgen.RegisterHandlersWithOptions(router, handler, feedbackgen.GinServerOptions{
		ErrorHandler: openAPIErrorHandler,
	})
}

// ListFeedback 处理 GET /api/feedback。
func (h *FeedbackHandler) ListFeedback(c *gin.Context, params feedbackgen.ListFeedbackParams) {
	page, size, err := normalizePagination(params.Page, params.Size)
	if err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	filter := feedbackdomain.Filter{}
	if params.FeedbackType != nil {
		filter.FeedbackType = *params.FeedbackType
	}
	if params.UserId != nil {
		filter.UserID = *params.UserId
	}
	if params.ItemId != nil {
		filter.ItemID = *params.ItemId
	}

	feedbacks, total, err := h.service.ListPage(c.Request.Context(), filter, page, size)
	if err != nil {
		writeError(c, err)
		return
	}

	result := make([]feedbackgen.Feedback, 0, len(feedbacks))
	for _, item := range feedbacks {
		result = append(result, toFeedbackResponse(item))
	}
	httpx.OK(c, feedbackgen.FeedbackPageData{
		List:  result,
		Page:  page,
		Size:  size,
		Total: total,
	})
}

// GetFeedback 处理 GET /api/feedback/{feedback_type}/{user_id}/{item_id}。
func (h *FeedbackHandler) GetFeedback(c *gin.Context, feedbackType feedbackgen.FeedbackType, userID feedbackgen.UserID, itemID feedbackgen.ItemID) {
	entity, err := h.service.Find(c.Request.Context(), feedbackType, userID, itemID)
	if err != nil {
		writeError(c, err)
		return
	}
	httpx.OK(c, toFeedbackResponse(*entity))
}

// CreateFeedbackCollection 处理 POST /api/feedback。
func (h *FeedbackHandler) CreateFeedbackCollection(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	single, many, err := decodeSingleOrSlice[feedbackgen.CreateFeedbackRequest](payload)
	if err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	if single != nil {
		entity := toCreateFeedback(*single)
		entity = normalizeFeedbackKey(entity)
		if err := h.service.Create(c.Request.Context(), entity); err != nil {
			writeError(c, err)
			return
		}

		created, err := h.service.Find(c.Request.Context(), entity.FeedbackType, entity.UserID, entity.ItemID)
		if err != nil {
			writeError(c, err)
			return
		}
		c.Header("Location", feedbackLocation(*created))
		httpx.Created(c, toFeedbackResponse(*created))
		return
	}

	feedbacks := make([]feedbackdomain.Feedback, 0, len(many))
	for _, item := range many {
		feedbacks = append(feedbacks, toCreateFeedback(item))
	}
	if err := h.service.CreateBatch(c.Request.Context(), feedbacks); err != nil {
		writeError(c, err)
		return
	}

	resp := make([]feedbackgen.Feedback, 0, len(feedbacks))
	for _, item := range feedbacks {
		resp = append(resp, toFeedbackResponse(item))
	}
	httpx.Created(c, resp)
}

// PatchFeedback 处理 PATCH /api/feedback/{feedback_type}/{user_id}/{item_id}。
func (h *FeedbackHandler) PatchFeedback(c *gin.Context, feedbackType feedbackgen.FeedbackType, userID feedbackgen.UserID, itemID feedbackgen.ItemID) {
	var req feedbackgen.PatchFeedbackJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	current, err := h.service.Find(c.Request.Context(), feedbackType, userID, itemID)
	if err != nil {
		writeError(c, err)
		return
	}
	if req.Value != nil {
		current.Value = *req.Value
	}
	if req.Timestamp != nil {
		current.Timestamp = *req.Timestamp
	}
	if err := h.service.Update(c.Request.Context(), *current); err != nil {
		writeError(c, err)
		return
	}

	updated, err := h.service.Find(c.Request.Context(), feedbackType, userID, itemID)
	if err != nil {
		writeError(c, err)
		return
	}
	httpx.OK(c, toFeedbackResponse(*updated))
}

// PatchFeedbackCollection 处理 PATCH /api/feedback。
func (h *FeedbackHandler) PatchFeedbackCollection(c *gin.Context) {
	var req feedbackgen.PatchFeedbackCollectionJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	result := make([]feedbackdomain.Feedback, 0, len(req))
	for _, item := range req {
		current, err := h.service.Find(c.Request.Context(), item.FeedbackType, item.UserId, item.ItemId)
		if err != nil {
			writeError(c, err)
			return
		}
		if item.Value != nil {
			current.Value = *item.Value
		}
		if item.Timestamp != nil {
			current.Timestamp = *item.Timestamp
		}
		result = append(result, *current)
	}

	if err := h.service.UpdateBatch(c.Request.Context(), result); err != nil {
		writeError(c, err)
		return
	}

	resp := make([]feedbackgen.Feedback, 0, len(result))
	for _, item := range result {
		resp = append(resp, toFeedbackResponse(item))
	}
	httpx.OK(c, resp)
}

// DeleteFeedback 处理 DELETE /api/feedback/{feedback_type}/{user_id}/{item_id}。
func (h *FeedbackHandler) DeleteFeedback(c *gin.Context, feedbackType feedbackgen.FeedbackType, userID feedbackgen.UserID, itemID feedbackgen.ItemID) {
	if err := h.service.Delete(c.Request.Context(), feedbackType, userID, itemID); err != nil {
		writeError(c, err)
		return
	}
	httpx.NoContent(c)
}

// DeleteFeedbackCollection 处理 DELETE /api/feedback。
func (h *FeedbackHandler) DeleteFeedbackCollection(c *gin.Context) {
	var req feedbackgen.DeleteFeedbackCollectionJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	keys := make([]feedbackdomain.Key, 0, len(req))
	for _, item := range req {
		keys = append(keys, feedbackdomain.Key{
			FeedbackType: item.FeedbackType,
			UserID:       item.UserId,
			ItemID:       item.ItemId,
		})
	}
	if err := h.service.DeleteBatch(c.Request.Context(), keys); err != nil {
		writeError(c, err)
		return
	}
	httpx.NoContent(c)
}

func toFeedbackResponse(src feedbackdomain.Feedback) feedbackgen.Feedback {
	return feedbackgen.Feedback{
		FeedbackType: src.FeedbackType,
		UserId:       src.UserID,
		ItemId:       src.ItemID,
		Value:        src.Value,
		Timestamp:    src.Timestamp,
	}
}

func toCreateFeedback(req feedbackgen.CreateFeedbackRequest) feedbackdomain.Feedback {
	entity := feedbackdomain.Feedback{
		FeedbackType: req.FeedbackType,
		UserID:       req.UserId,
		ItemID:       req.ItemId,
	}
	if req.Value != nil {
		entity.Value = *req.Value
	}
	if req.Timestamp != nil {
		entity.Timestamp = *req.Timestamp
	}
	return entity
}

func normalizeFeedbackKey(src feedbackdomain.Feedback) feedbackdomain.Feedback {
	src.FeedbackType = strings.TrimSpace(src.FeedbackType)
	src.UserID = strings.TrimSpace(src.UserID)
	src.ItemID = strings.TrimSpace(src.ItemID)
	return src
}

func feedbackLocation(src feedbackdomain.Feedback) string {
	return "/api/feedback/" + url.PathEscape(src.FeedbackType) + "/" + url.PathEscape(src.UserID) + "/" + url.PathEscape(src.ItemID)
}
