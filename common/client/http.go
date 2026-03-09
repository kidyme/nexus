package client

import (
	"net/http"
	"time"
)

// DefaultHTTPClient 封装默认 HTTP 客户端，便于各服务间调用。
var DefaultHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}
