// BizError 业务错误接口（用于单体应用）
type BizError interface {
	error
	GetCode() string
	GetStatus() int
}

// ErrorHandler 定义错误处理器类型
type ErrorHandler func(c *gin.Context, err error)

// HTTPServerOption 定义服务器选项
type HTTPServerOption func(*{{.Name}}HTTPServer)

// {{.Name}}HTTPServer HTTP服务器结构
type {{.Name}}HTTPServer struct {
	server {{.ServiceName}}
	router gin.IRouter
	errorHandler ErrorHandler
}

// DefaultErrorHandler 默认错误处理器（支持单体和分布式）
func DefaultErrorHandler(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// 1. 优先检查是否是业务错误（单体应用场景）
	if bizErr, ok := err.(BizError); ok {
		c.JSON(bizErr.GetStatus(), gin.H{
			"code": bizErr.GetCode(),
			"message": bizErr.Error(),
			"success": false,
		})
		return
	}

	// 2. 检查是否是 gRPC 错误（分布式场景）
	if s, ok := status.FromError(err); ok {
		httpStatus := grpcCodeToHTTP(s.Code())
		c.JSON(httpStatus, gin.H{
			"code": s.Code().String(),
			"message": s.Message(),
			"success": false,
		})
		return
	}

	// 3. 参数绑定错误
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": "INVALID_JSON",
			"message": "请求格式错误",
			"success": false,
		})
		return
	}

	// 4. 默认错误处理
	c.JSON(http.StatusInternalServerError, gin.H{
		"code": "INTERNAL_ERROR",
		"message": err.Error(),
		"success": false,
	})
}

// grpcCodeToHTTP 将 gRPC 状态码转换为 HTTP 状态码
func grpcCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Canceled:
		return 499 // Client Closed Request
	case codes.Unknown, codes.Internal, codes.DataLoss:
		return http.StatusInternalServerError
	case codes.Aborted:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// WithErrorHandler 设置自定义错误处理器
func WithErrorHandler(h ErrorHandler) HTTPServerOption {
	return func(s *{{.Name}}HTTPServer) {
		s.errorHandler = h
	}
}

// Register{{.ServiceName}}HTTPServer 注册HTTP服务器
func Register{{.ServiceName}}HTTPServer(srv {{.ServiceName}}, r gin.IRouter, opts ...HTTPServerOption) {
	s := {{.Name}}HTTPServer{
		server: srv,
		router: r,
		errorHandler: DefaultErrorHandler,
	}

	// 应用选项
	for _, opt := range opts {
		opt(&s)
	}

	s.RegisterService()
}
{{range .Methods}}
func (s *{{$.Name}}HTTPServer) {{.HandlerName}}(c *gin.Context) {
	var in {{.Request}}
{{if eq .Method "GET" "DELETE"}}
	if err := c.ShouldBindQuery(&in); err != nil {
		s.errorHandler(c, err)
		return
	}
{{else if eq .Method "POST" "PUT" "PATCH"}}
	if err := c.ShouldBindJSON(&in); err != nil {
		s.errorHandler(c, err)
		return
	}
{{else}}
	if err := c.ShouldBind(&in); err != nil {
		s.errorHandler(c, err)
		return
	}
{{end}}
{{if .HasPathParams}}{{range $param := .PathParams}}
	in.{{ProtoFieldName $param}} = c.Param("{{$param}}")
{{end}}{{end}}
	ctx := c.Request.Context()
	out, err := s.server.{{.Name}}(ctx, &in)
	if err != nil {
		s.errorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": "SUCCESS",
		"message": "ok",
		"data": out,
		"success": true,
	})
}
{{end}}
func (s *{{.Name}}HTTPServer) RegisterService() {
{{range .Methods}}	s.router.Handle("{{.Method}}", "{{.Path}}", s.{{.HandlerName}})
{{end}}
}