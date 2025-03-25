package retry

import (
	"errors"
	"net"
	"net/http"
	"syscall"
)

// IsNetworkError 判断是否为网络错误
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}

	// 检查常见的网络错误
	if errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNABORTED) ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}

	return false
}

// IsHTTPRetryable 判断HTTP错误是否可重试
func IsHTTPRetryable(statusCode int) bool {
	// 5xx 服务器错误和部分 4xx 客户端错误可重试
	return statusCode >= 500 || statusCode == http.StatusTooManyRequests || statusCode == http.StatusRequestTimeout
}

// IsRetryableHTTPError 判断HTTP错误是否可重试
func IsRetryableHTTPError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否为网络错误
	if IsNetworkError(err) {
		return true
	}

	// 检查是否为 HTTP 错误
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return IsHTTPRetryable(httpErr.StatusCode)
	}

	return false
}

// HTTPError 表示 HTTP 错误
type HTTPError struct {
	StatusCode int
	Message    string
}

// Error 实现 error 接口
func (e *HTTPError) Error() string {
	return e.Message
}

// NewHTTPError 创建新的 HTTP 错误
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}
