// Package httpx 提供共享 HTTP 响应辅助能力。
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

// Response 是共享 HTTP 响应包裹结构。
type Response[T any] struct {
	Errno   int    `json:"errno"`
	Message string `json:"msg"`
	Data    T      `json:"data"`
	TraceID string `json:"trace_id"`
}

// OK 写入成功响应。
func OK[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, Response[T]{
		Errno:   ErrnoOK,
		Message: "ok",
		Data:    data,
		TraceID: TraceID(c),
	})
}

// Created 写入 201 成功响应。
func Created[T any](c *gin.Context, data T) {
	c.JSON(http.StatusCreated, Response[T]{
		Errno:   ErrnoOK,
		Message: "ok",
		Data:    data,
		TraceID: TraceID(c),
	})
}

// NoContent 写入 204 成功响应。
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Fail 写入失败响应。
func Fail(c *gin.Context, httpStatus, errno int, msg string) {
	c.JSON(httpStatus, Response[any]{
		Errno:   errno,
		Message: msg,
		TraceID: TraceID(c),
	})
}

// BadRequest 写入 400 错误响应。
func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, ErrnoBadRequest, msg)
}

// NotFound 写入 404 错误响应。
func NotFound(c *gin.Context, msg string) {
	Fail(c, http.StatusNotFound, ErrnoNotFound, msg)
}

// InternalError 写入 500 错误响应。
func InternalError(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, ErrnoInternal, msg)
}

// TraceID 返回当前 trace id；若不存在则返回空字符串。
func TraceID(c *gin.Context) string {
	if traceID := c.GetString(traceIDContextKey); traceID != "" {
		return traceID
	}
	return c.GetHeader(traceIDHeader)
}
