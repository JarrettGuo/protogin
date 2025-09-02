package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apiv1 "github.com/JarrettGuo/protogin/gen/api/v1"
	"github.com/JarrettGuo/protogin/internal/errors"
)

// 实现服务接口
type demoServer struct {
	apiv1.UnimplementedDemoServiceServer
	mode string // "monolithic" 或 "distributed"
}

func (s *demoServer) GetUser(ctx context.Context, req *apiv1.GetUserRequest) (*apiv1.GetUserResponse, error) {
	log.Printf("GetUser called with ID: %s", req.UserId)

	// 演示不同模式下的错误处理
	if req.UserId == "404" {
		if s.mode == "monolithic" {
			// 单体模式：返回业务错误
			return nil, errors.New("USER_NOT_FOUND", "用户不存在", 404)
		} else {
			// 分布式模式：返回 gRPC 错误
			return nil, status.Error(codes.NotFound, "user not found")
		}
	}

	if req.UserId == "403" {
		if s.mode == "monolithic" {
			return nil, errors.New("FORBIDDEN", "没有权限访问该用户", 403)
		} else {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
	}

	if req.UserId == "500" {
		if s.mode == "monolithic" {
			return nil, errors.New("INTERNAL_ERROR", "服务器内部错误", 500)
		} else {
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &apiv1.GetUserResponse{
		UserId: req.UserId,
		Name:   "Test User",
		Email:  "test@example.com",
		Age:    25,
	}, nil
}

func (s *demoServer) CreateUser(ctx context.Context, req *apiv1.CreateUserRequest) (*apiv1.CreateUserResponse, error) {
	log.Printf("CreateUser called: %s", req.Name)

	// 参数验证
	if req.Name == "" {
		if s.mode == "monolithic" {
			return nil, errors.New("INVALID_NAME", "用户名不能为空", 400)
		} else {
			return nil, status.Error(codes.InvalidArgument, "name is required")
		}
	}

	if req.Age < 0 || req.Age > 150 {
		if s.mode == "monolithic" {
			return nil, errors.New("INVALID_AGE", "年龄必须在0-150之间", 400)
		} else {
			return nil, status.Error(codes.InvalidArgument, "age must be between 0 and 150")
		}
	}

	// 模拟用户已存在的情况
	if req.Email == "exists@example.com" {
		if s.mode == "monolithic" {
			return nil, errors.New("USER_EXISTS", "该邮箱已被注册", 409)
		} else {
			return nil, status.Error(codes.AlreadyExists, "email already registered")
		}
	}

	return &apiv1.CreateUserResponse{
		UserId:  "user_123",
		Message: "User created successfully",
	}, nil
}

func (s *demoServer) UpdateUser(ctx context.Context, req *apiv1.UpdateUserRequest) (*apiv1.UpdateUserResponse, error) {
	log.Printf("UpdateUser called for ID: %s", req.UserId)

	// 模拟用户不存在
	if req.UserId == "404" {
		if s.mode == "monolithic" {
			return nil, errors.New("USER_NOT_FOUND", "用户不存在", 404)
		} else {
			return nil, status.Error(codes.NotFound, "user not found")
		}
	}

	// 参数验证
	if req.Name != "" && len(req.Name) < 2 {
		if s.mode == "monolithic" {
			return nil, errors.New("INVALID_NAME", "用户名至少需要2个字符", 400)
		} else {
			return nil, status.Error(codes.InvalidArgument, "name must be at least 2 characters")
		}
	}

	return &apiv1.UpdateUserResponse{
		Success: true,
		Message: "User updated successfully",
	}, nil
}

func (s *demoServer) DeleteUser(ctx context.Context, req *apiv1.DeleteUserRequest) (*apiv1.DeleteUserResponse, error) {
	log.Printf("DeleteUser called for ID: %s", req.UserId)

	// 模拟用户不存在
	if req.UserId == "404" {
		if s.mode == "monolithic" {
			return nil, errors.New("USER_NOT_FOUND", "用户不存在", 404)
		} else {
			return nil, status.Error(codes.NotFound, "user not found")
		}
	}

	// 模拟没有权限删除
	if req.UserId == "admin" {
		if s.mode == "monolithic" {
			return nil, errors.New("FORBIDDEN", "不能删除管理员账户", 403)
		} else {
			return nil, status.Error(codes.PermissionDenied, "cannot delete admin user")
		}
	}

	return &apiv1.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

func (s *demoServer) ListUsers(ctx context.Context, req *apiv1.ListUsersRequest) (*apiv1.ListUsersResponse, error) {
	log.Printf("ListUsers called: page=%d, pageSize=%d", req.Page, req.PageSize)

	// 参数验证
	if req.Page < 0 {
		if s.mode == "monolithic" {
			return nil, errors.New("INVALID_PAGE", "页码不能为负数", 400)
		} else {
			return nil, status.Error(codes.InvalidArgument, "page cannot be negative")
		}
	}

	if req.PageSize < 0 || req.PageSize > 100 {
		if s.mode == "monolithic" {
			return nil, errors.New("INVALID_PAGE_SIZE", "每页数量必须在1-100之间", 400)
		} else {
			return nil, status.Error(codes.InvalidArgument, "page_size must be between 1 and 100")
		}
	}

	// 默认值处理
	page := req.Page
	if page == 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	return &apiv1.ListUsersResponse{
		Users: []*apiv1.User{
			{
				UserId: "user_1",
				Name:   "Alice",
				Email:  "alice@example.com",
				Age:    30,
			},
			{
				UserId: "user_2",
				Name:   "Bob",
				Email:  "bob@example.com",
				Age:    25,
			},
		},
		Total:    2,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *demoServer) BatchOperation(ctx context.Context, req *apiv1.BatchRequest) (*apiv1.BatchResponse, error) {
	log.Printf("BatchOperation called: %s on %d items", req.Operation, len(req.Ids))

	// 参数验证
	if len(req.Ids) == 0 {
		if s.mode == "monolithic" {
			return nil, errors.New("INVALID_IDS", "ID列表不能为空", 400)
		} else {
			return nil, status.Error(codes.InvalidArgument, "ids cannot be empty")
		}
	}

	if len(req.Ids) > 100 {
		if s.mode == "monolithic" {
			return nil, errors.New("TOO_MANY_IDS", "批量操作最多支持100个ID", 400)
		} else {
			return nil, status.Error(codes.InvalidArgument, "batch operation supports maximum 100 ids")
		}
	}

	if req.Operation == "" {
		if s.mode == "monolithic" {
			return nil, errors.New("INVALID_OPERATION", "操作类型不能为空", 400)
		} else {
			return nil, status.Error(codes.InvalidArgument, "operation is required")
		}
	}

	// 模拟部分失败的情况
	var failedIds []string
	for _, id := range req.Ids {
		if id == "fail" {
			failedIds = append(failedIds, id)
		}
	}

	processed := int32(len(req.Ids) - len(failedIds))

	return &apiv1.BatchResponse{
		Success:   len(failedIds) == 0,
		Processed: processed,
		FailedIds: failedIds,
	}, nil
}

func main() {
	// 从环境变量或命令行参数获取运行模式
	mode := os.Getenv("RUN_MODE")
	if mode == "" {
		mode = "monolithic" // 默认为单体模式
	}

	if mode != "monolithic" && mode != "distributed" {
		log.Fatal("Invalid RUN_MODE. Must be 'monolithic' or 'distributed'")
	}

	srv := &demoServer{
		mode: mode,
	}

	if mode == "monolithic" {
		// 单体模式：同时启动 gRPC 和 HTTP
		go startGRPCServer(srv)
		startHTTPServer(srv, mode)
	} else {
		// 分布式模式：只启动 HTTP（假设 gRPC 服务在其他地方）
		startHTTPServer(srv, mode)
	}
}

func startGRPCServer(srv *demoServer) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	// gRPC 拦截器：将业务错误转换为 gRPC 错误
	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			// 如果是业务错误，转换为 gRPC 错误
			if bizErr, ok := err.(errors.BizError); ok {
				code := httpStatusToGRPCCode(bizErr.GetStatus())
				return nil, status.Error(code, bizErr.Error())
			}
		}
		return resp, err
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	apiv1.RegisterDemoServiceServer(s, srv)

	log.Println("🚀 gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func startHTTPServer(srv *demoServer, mode string) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"mode":   mode,
		})
	})

	// 自定义错误处理器（可选）
	customErrorHandler := func(c *gin.Context, err error) {
		// 这里可以添加日志记录、监控上报等
		log.Printf("Error occurred: %v", err)

		// 使用默认错误处理器的逻辑
		apiv1.DefaultErrorHandler(c, err)
	}

	// 注册 HTTP 服务
	apiv1.RegisterDemoServiceServerHTTPServer(srv, r,
		apiv1.WithErrorHandler(customErrorHandler))

	log.Println("===============================================")
	log.Println("🚀 Server starting on :8080")
	log.Printf("🔄 Mode: %s", mode)
	log.Println("===============================================")
	log.Println("📍 Health check:")
	log.Println("   GET http://localhost:8080/health")
	log.Println("")
	log.Println("📍 API endpoints:")
	log.Println("   GET    /api/v1/users/:user_id")
	log.Println("   POST   /api/v1/users")
	log.Println("   PUT    /api/v1/users/:user_id")
	log.Println("   DELETE /api/v1/users/:user_id")
	log.Println("   GET    /api/v1/users")
	log.Println("   POST   /api/v1/batch")

	if mode == "monolithic" {
		log.Println("")
		log.Println("📍 gRPC endpoint:")
		log.Println("   localhost:50051")
	}

	log.Println("")
	log.Println("📝 Test examples:")
	log.Println("   # 成功请求")
	log.Println("   curl http://localhost:8080/api/v1/users/123")
	log.Println("")
	log.Println("   # 404错误")
	log.Println("   curl http://localhost:8080/api/v1/users/404")
	log.Println("")
	log.Println("   # 403错误")
	log.Println("   curl http://localhost:8080/api/v1/users/403")
	log.Println("")
	log.Println("   # 创建用户")
	log.Println(`   curl -X POST http://localhost:8080/api/v1/users \`)
	log.Println(`        -H "Content-Type: application/json" \`)
	log.Println(`        -d '{"name":"John","email":"john@example.com","age":30}'`)
	log.Println("")
	log.Println("   # 参数错误")
	log.Println(`   curl -X POST http://localhost:8080/api/v1/users \`)
	log.Println(`        -H "Content-Type: application/json" \`)
	log.Println(`        -d '{"name":"","email":"test@example.com","age":30}'`)
	log.Println("")
	log.Println("   # 用户列表")
	log.Println("   curl 'http://localhost:8080/api/v1/users?page=1&page_size=10'")
	log.Println("===============================================")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func httpStatusToGRPCCode(status int) codes.Code {
	switch status {
	case 400:
		return codes.InvalidArgument
	case 401:
		return codes.Unauthenticated
	case 403:
		return codes.PermissionDenied
	case 404:
		return codes.NotFound
	case 409:
		return codes.AlreadyExists
	case 429:
		return codes.ResourceExhausted
	case 500:
		return codes.Internal
	case 501:
		return codes.Unimplemented
	case 503:
		return codes.Unavailable
	case 504:
		return codes.DeadlineExceeded
	default:
		return codes.Unknown
	}
}
