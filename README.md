# protoc-gen-gin

将 Protocol Buffers 服务定义转换为 Gin HTTP 处理器的 protoc 插件。

## 项目介绍

`protoc-gen-gin` 是一个 protoc 插件，通过解析 proto 文件中的 Google API HTTP 注解，自动生成 Gin 框架的 RESTful API 处理代码。它让你能够用一份 proto 定义同时支持 gRPC 和 HTTP 服务。

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

### 生成代码使用

```go
package main

import (
    "context"
    "github.com/gin-gonic/gin"
    pb "your/package/path"
)

type server struct {
    pb.UnimplementedUserServiceServer
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // 实现业务逻辑
    return &pb.GetUserResponse{
        UserId: req.UserId,
        Name: "John Doe",
    }, nil
}

func main() {
    r := gin.Default()
    srv := &server{}
    
    // 注册 HTTP 处理器
    pb.RegisterUserServiceHTTPServer(srv, r)
    
    r.Run(":8080")
}
```

## 路线图

- [x] 基础 HTTP 方法支持
- [x] 路径参数绑定
- [x] Query 和 Body 参数自动绑定
- [ ] 支持自定义错误处理
- [ ] 支持中间件配置
- [ ] 支持流式 RPC
- [ ] 支持文件上传下载
- [ ] 生成 OpenAPI 文档
- [ ] 支持请求验证
- [ ] 支持自定义响应包装

## License

MIT