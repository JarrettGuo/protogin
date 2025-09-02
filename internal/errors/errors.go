package errors

import "fmt"

// BizError 业务错误接口
type BizError interface {
	error
	GetCode() string
	GetStatus() int
}

// bizError 业务错误实现
type bizError struct {
	Code    string
	Message string
	Status  int
}

func (e *bizError) Error() string {
	return e.Message
}

func (e *bizError) GetCode() string {
	return e.Code
}

func (e *bizError) GetStatus() int {
	return e.Status
}

// New 创建业务错误
func New(code string, message string, status int) BizError {
	return &bizError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Newf 创建格式化的业务错误
func Newf(code string, status int, format string, args ...interface{}) BizError {
	return &bizError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Status:  status,
	}
}

// 预定义的常用错误
var (
	ErrNotFound     = New("NOT_FOUND", "资源不存在", 404)
	ErrInvalidParam = New("INVALID_PARAM", "参数错误", 400)
	ErrUnauthorized = New("UNAUTHORIZED", "未授权", 401)
	ErrForbidden    = New("FORBIDDEN", "禁止访问", 403)
	ErrConflict     = New("CONFLICT", "资源冲突", 409)
	ErrInternal     = New("INTERNAL", "内部错误", 500)
)

// Wrap 包装错误
func Wrap(err error, code string, status int) BizError {
	if err == nil {
		return nil
	}
	return &bizError{
		Code:    code,
		Message: err.Error(),
		Status:  status,
	}
}

// Is 判断是否是指定的错误
func Is(err error, target BizError) bool {
	if err == nil || target == nil {
		return false
	}

	bizErr, ok := err.(BizError)
	if !ok {
		return false
	}

	return bizErr.GetCode() == target.GetCode()
}
