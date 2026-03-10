// Package httpx provides shared HTTP response helpers.
package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ErrnoOK           = 0
	ErrnoBadRequest   = 400
	ErrnoNotFound     = 404
	ErrnoInternal     = 500
	traceIDContextKey = "trace_id"
	traceIDHeader     = "X-Trace-Id"
)

// Response is the shared HTTP response envelope.
type Response[T any] struct {
	Errno   int    `json:"errno"`
	Message string `json:"msg"`
	Data    T      `json:"data"`
	TraceID string `json:"trace_id"`
}

// OK writes a success response.
func OK[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, Response[T]{
		Errno:   ErrnoOK,
		Message: "ok",
		Data:    data,
		TraceID: TraceID(c),
	})
}

// Fail writes an error response.
func Fail(c *gin.Context, httpStatus, errno int, msg string) {
	c.JSON(httpStatus, Response[any]{
		Errno:   errno,
		Message: msg,
		TraceID: TraceID(c),
	})
}

// BadRequest writes a 400 error response.
func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, ErrnoBadRequest, msg)
}

// NotFound writes a 404 error response.
func NotFound(c *gin.Context, msg string) {
	Fail(c, http.StatusNotFound, ErrnoNotFound, msg)
}

// InternalError writes a 500 error response.
func InternalError(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, ErrnoInternal, msg)
}

// TraceID returns the current trace id if present.
func TraceID(c *gin.Context) string {
	if traceID := c.GetString(traceIDContextKey); traceID != "" {
		return traceID
	}
	return c.GetHeader(traceIDHeader)
}
