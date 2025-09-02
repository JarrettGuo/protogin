# protoc-gen-gin

将 Protocol Buffers 服务定义转换为 Gin HTTP 处理器的 protoc 插件。

## 项目介绍

`protoc-gen-gin` 是一个 protoc 插件，通过解析 proto 文件中的 Google API HTTP 注解，自动生成 Gin 框架的 RESTful API 处理代码。它让你能够用一份 proto 定义同时支持 gRPC 和 HTTP 服务，并且支持单体应用和微服务架构的无缝切换。

## 特性

- 🚀 **双协议支持**: 一份 proto 定义，同时生成 gRPC 和 HTTP 接口
- 🔄 **架构灵活**: 支持单体应用和微服务架构，可无缝切换
- 🛡️ **智能错误处理**: 自动转换 gRPC 错误码到 HTTP 状态码
- 🎯 **自动参数绑定**: 智能识别路径参数、Query 参数和 Body 参数
- 🔧 **可扩展设计**: 支持自定义错误处理器和中间件
- 📦 **统一响应格式**: 内置标准化的 JSON 响应格式

## 前置要求

- Go 1.21+
- Protocol Buffers 编译器 (protoc)
- buf (可选，但推荐)

## 快速开始

### 安装插件

```bash
go install github.com/JarrettGuo/protoc-gen-gin@latest
```

### 使用 protoc 命令

```bash
# 直接使用 protoc
protoc -I. -I$GOPATH/src \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --gin_out=. --gin_opt=paths=source_relative \
  api.proto

# 或使用 buf (推荐)
buf generate
```

## Proto 文件定义规则

### 基本示例

```protobuf
syntax = "proto3";

package api.v1;

import "google/api/annotations.proto";

// UserService 用户服务
service UserService {
  // GetUser 获取单个用户
  // 路径参数会自动绑定到请求结构体
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = {
      get: "/api/v1/users/{user_id}"  // {user_id} 会转换为 :user_id
    };
  }
  
  // CreateUser 创建用户
  // POST/PUT/PATCH 请求自动使用 JSON 绑定
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (google.api.http) = {
      post: "/api/v1/users"
      body: "*"  // 使用整个请求体
    };
  }
  
  // UpdateUser 更新用户
  // 支持同时使用路径参数和请求体
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
    option (google.api.http) = {
      put: "/api/v1/users/{user_id}"
      body: "*"
    };
  }
  
  // DeleteUser 删除用户
  // DELETE/GET 请求自动使用 Query 参数绑定
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse) {
    option (google.api.http) = {
      delete: "/api/v1/users/{user_id}"
    };
  }
  
  // ListUsers 列表查询
  // Query 参数会自动映射到请求结构体字段
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse) {
    option (google.api.http) = {
      get: "/api/v1/users"  // ?page=1&page_size=10
    };
  }
  
  // BatchUpdate 批量操作
  // 支持多个 HTTP 绑定到同一个 RPC 方法
  rpc BatchUpdate(BatchRequest) returns (BatchResponse) {
    option (google.api.http) = {
      post: "/api/v1/users/batch"
      body: "*"
      additional_bindings {
        post: "/api/v1/users/batch/async"
        body: "*"
      }
    };
  }
}
```

### 支持的 HTTP 方法

- `get`: GET 请求，参数通过 Query 传递
- `post`: POST 请求，参数通过 Body 传递
- `put`: PUT 请求，参数通过 Body 传递
- `delete`: DELETE 请求，参数通过 Query 传递
- `patch`: PATCH 请求，参数通过 Body 传递
- `custom`: 自定义 HTTP 方法

### 参数绑定规则

| HTTP 方法 | 绑定方式 | 示例 |
|----------|---------|------|
| GET | Query 参数 | `/users?name=john` |
| POST/PUT/PATCH | JSON Body | `{"name": "john"}` |
| DELETE | Query 参数 | `/users/123?force=true` |
| 路径参数 | URL 路径 | `/users/{user_id}` → `/users/123` |

## 错误处理

### 自动错误码转换

插件会自动将 gRPC 错误码转换为对应的 HTTP 状态码：

| gRPC 错误码 | HTTP 状态码 | 描述 |
|------------|------------|------|
| OK | 200 | 成功 |
| InvalidArgument | 400 | 参数错误 |
| NotFound | 404 | 资源不存在 |
| PermissionDenied | 403 | 权限不足 |
| Unauthenticated | 401 | 未认证 |
| AlreadyExists | 409 | 资源冲突 |
| ResourceExhausted | 429 | 请求过多 |
| Internal | 500 | 内部错误 |
| Unavailable | 503 | 服务不可用 |

### 业务错误支持（单体应用）

```go
import "github.com/JarrettGuo/protogin/pkg/errors"

// 创建业务错误
err := errors.New("USER_NOT_FOUND", "用户不存在", 404)

// 在服务实现中返回
func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    user, err := s.userRepo.FindByID(req.UserId)
    if err != nil {
        return nil, errors.New("USER_NOT_FOUND", "用户不存在", 404)
    }
    return &pb.GetUserResponse{...}, nil
}
```

### 自定义错误处理器

```go
// 定义自定义错误处理器
customErrorHandler := func(c *gin.Context, err error) {
    // 自定义错误处理逻辑
    log.Printf("Error: %v", err)
    
    c.JSON(500, gin.H{
        "success": false,
        "message": err.Error(),
        "timestamp": time.Now().Unix(),
    })
}

// 注册时使用
pb.RegisterUserServiceHTTPServer(srv, r, 
    pb.WithErrorHandler(customErrorHandler))
```

## 使用示例

### 单体应用模式

同时启动 gRPC 和 HTTP 服务，直接访问数据库：

```go
package main

import (
    "context"
    "net"
    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"
    pb "your/package/path"
    "github.com/JarrettGuo/protogin/pkg/errors"
)

type server struct {
    pb.UnimplementedUserServiceServer
    db *sql.DB  // 直接访问数据库
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // 单体模式：返回业务错误
    user, err := s.db.QueryRow("SELECT * FROM users WHERE id = ?", req.UserId)
    if err == sql.ErrNoRows {
        return nil, errors.New("USER_NOT_FOUND", "用户不存在", 404)
    }
    // ...
}

func main() {
    srv := &server{db: initDB()}
    
    // 启动 gRPC 服务
    go func() {
        lis, _ := net.Listen("tcp", ":50051")
        s := grpc.NewServer()
        pb.RegisterUserServiceServer(s, srv)
        s.Serve(lis)
    }()
    
    // 启动 HTTP 服务
    r := gin.Default()
    pb.RegisterUserServiceHTTPServer(srv, r)
    r.Run(":8080")
}
```

### 微服务模式

作为独立的 API Gateway，调用后端 gRPC 服务：

```go
func main() {
    // 连接到 gRPC 服务
    conn, _ := grpc.Dial("user-service:50051", grpc.WithInsecure())
    client := pb.NewUserServiceClient(conn)
    
    r := gin.Default()
    
    // 代理实现
    proxy := &proxyServer{client: client}
    pb.RegisterUserServiceHTTPServer(proxy, r)
    
    r.Run(":8080")
}
```

## 统一响应格式

生成的代码会自动包装响应为统一格式：

**成功响应:**
```json
{
    "code": "SUCCESS",
    "message": "ok",
    "data": {
        "user_id": "123",
        "name": "John"
    },
    "success": true
}
```

**错误响应:**
```json
{
    "code": "USER_NOT_FOUND",
    "message": "用户不存在",
    "success": false
}
```

## 高级特性

### 选项模式配置

```go
// 支持多种选项配置
pb.RegisterUserServiceHTTPServer(srv, r,
    pb.WithErrorHandler(customErrorHandler),
    pb.WithMiddleware(authMiddleware),
    pb.WithTimeout(30*time.Second),
)
```

### gRPC 拦截器集成

单体模式下自动转换业务错误为 gRPC 错误：

```go
interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    resp, err := handler(ctx, req)
    if bizErr, ok := err.(errors.BizError); ok {
        // 自动转换业务错误为 gRPC 错误
        return nil, status.Error(toGRPCCode(bizErr.GetStatus()), bizErr.Error())
    }
    return resp, err
}
```

## 路线图

- [x] 基础 HTTP 方法支持
- [x] 路径参数绑定
- [x] Query 和 Body 参数自动绑定
- [x] 支持自定义错误处理
- [x] 支持单体和微服务双模式
- [x] 智能错误码转换
- [x] 统一响应格式
- [ ] 支持中间件配置
- [ ] 支持流式 RPC
- [ ] 支持文件上传下载
- [ ] 生成 OpenAPI 文档
- [ ] 支持请求验证
- [ ] 支持限流和熔断

## 贡献

欢迎提交 Issue 和 Pull Request！

## License

MIT
