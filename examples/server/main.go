package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	apiv1 "github.com/JarrettGuo/protogin/gen/api/v1"
)

// 实现服务接口
type demoServer struct {
	apiv1.UnimplementedDemoServiceServer
}

func (s *demoServer) GetUser(ctx context.Context, req *apiv1.GetUserRequest) (*apiv1.GetUserResponse, error) {
	log.Printf("GetUser called with ID: %s", req.UserId)
	return &apiv1.GetUserResponse{
		UserId: req.UserId,
		Name:   "Test User",
		Email:  "test@example.com",
		Age:    25,
	}, nil
}

func (s *demoServer) CreateUser(ctx context.Context, req *apiv1.CreateUserRequest) (*apiv1.CreateUserResponse, error) {
	log.Printf("CreateUser called: %s", req.Name)
	return &apiv1.CreateUserResponse{
		UserId:  "user_123",
		Message: "User created successfully",
	}, nil
}

func (s *demoServer) UpdateUser(ctx context.Context, req *apiv1.UpdateUserRequest) (*apiv1.UpdateUserResponse, error) {
	log.Printf("UpdateUser called for ID: %s", req.UserId)
	return &apiv1.UpdateUserResponse{
		Success: true,
		Message: "User updated successfully",
	}, nil
}

func (s *demoServer) DeleteUser(ctx context.Context, req *apiv1.DeleteUserRequest) (*apiv1.DeleteUserResponse, error) {
	log.Printf("DeleteUser called for ID: %s", req.UserId)
	return &apiv1.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

func (s *demoServer) ListUsers(ctx context.Context, req *apiv1.ListUsersRequest) (*apiv1.ListUsersResponse, error) {
	log.Printf("ListUsers called: page=%d, pageSize=%d", req.Page, req.PageSize)
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
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *demoServer) BatchOperation(ctx context.Context, req *apiv1.BatchRequest) (*apiv1.BatchResponse, error) {
	log.Printf("BatchOperation called: %s on %d items", req.Operation, len(req.Ids))
	return &apiv1.BatchResponse{
		Success:   true,
		Processed: int32(len(req.Ids)),
		FailedIds: []string{},
	}, nil
}

func main() {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// 添加日志中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// 注册 HTTP 服务
	srv := &demoServer{}
	apiv1.RegisterDemoServiceServerHTTPServer(srv, r)

	log.Println("🚀 Server starting on :8080")
	log.Println("📍 Health check: http://localhost:8080/health")
	log.Println("📍 API endpoints:")
	log.Println("   GET    /api/v1/users")
	log.Println("   POST   /api/v1/users")
	log.Println("   GET    /api/v1/users/:user_id")
	log.Println("   PUT    /api/v1/users/:user_id")
	log.Println("   DELETE /api/v1/users/:user_id")
	log.Println("   POST   /api/v1/batch")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
