package errors

import (
	"fmt"
)

// ErrorCode 错误码类型
type ErrorCode string

const (
	// ErrCodeLoginRequired 需要登录
	ErrCodeLoginRequired ErrorCode = "LOGIN_REQUIRED"

	// ErrCodeNetworkError 网络错误
	ErrCodeNetworkError ErrorCode = "NETWORK_ERROR"

	// ErrCodeBrowserError 浏览器错误
	ErrCodeBrowserError ErrorCode = "BROWSER_ERROR"

	// ErrCodeElementNotFound 元素未找到
	ErrCodeElementNotFound ErrorCode = "ELEMENT_NOT_FOUND"

	// ErrCodeTimeout 超时
	ErrCodeTimeout ErrorCode = "TIMEOUT"

	// ErrCodeInvalidParameter 无效参数
	ErrCodeInvalidParameter ErrorCode = "INVALID_PARAMETER"

	// ErrCodeUploadFailed 上传失败
	ErrCodeUploadFailed ErrorCode = "UPLOAD_FAILED"

	// ErrCodePublishFailed 发布失败
	ErrCodePublishFailed ErrorCode = "PUBLISH_FAILED"
)

// AppError 应用错误
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError 创建新的应用错误
func NewAppError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 预定义错误构造函数

// ErrLoginRequired 创建登录必需错误
func ErrLoginRequired(err error) *AppError {
	return NewAppError(ErrCodeLoginRequired, "需要登录后才能操作", err)
}

// ErrNetwork 创建网络错误
func ErrNetwork(message string, err error) *AppError {
	return NewAppError(ErrCodeNetworkError, message, err)
}

// ErrBrowser 创建浏览器错误
func ErrBrowser(message string, err error) *AppError {
	return NewAppError(ErrCodeBrowserError, message, err)
}

// ErrElementNotFound 创建元素未找到错误
func ErrElementNotFound(selector string, err error) *AppError {
	return NewAppError(ErrCodeElementNotFound, fmt.Sprintf("未找到元素: %s", selector), err)
}

// ErrTimeout 创建超时错误
func ErrTimeout(message string, err error) *AppError {
	return NewAppError(ErrCodeTimeout, message, err)
}

// ErrInvalidParameter 创建无效参数错误
func ErrInvalidParameter(param string) *AppError {
	return NewAppError(ErrCodeInvalidParameter, fmt.Sprintf("无效参数: %s", param), nil)
}

// ErrUploadFailed 创建上传失败错误
func ErrUploadFailed(message string, err error) *AppError {
	return NewAppError(ErrCodeUploadFailed, message, err)
}

// ErrPublishFailed 创建发布失败错误
func ErrPublishFailed(message string, err error) *AppError {
	return NewAppError(ErrCodePublishFailed, message, err)
}
