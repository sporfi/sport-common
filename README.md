# sport-common

一个为sportsio应用提供的 Go 语言通用组件库，支持 go-zero 和 go-kratos 双框架，提供认证中间件、工具函数和常量定义等基础功能。

## ✨ 特性

- 🔐 **JWT 认证中间件** - 支持 go-zero 和 go-kratos 双框架
- 🎯 **路由白名单** - 灵活配置无需认证的公开接口
- ❄️ **雪花算法 ID 生成器** - 分布式唯一 ID 生成
- 🔧 **常量管理** - 统一管理项目常量和配置
- 🚀 **高性能** - 优化的中间件实现，最小化性能开销
- 📦 **开箱即用** - 简单集成，快速上手

## 📁 项目结构

```
sport-common/
├── pkg/
│   ├── constants/          # 常量定义
│   │   └── common.go       # 通用常量
│   ├── middleware/         # 中间件
│   │   ├── auth_middleware.go  # JWT认证中间件
│   │   └── jwt.go          # JWT工具函数
│   └── utils/              # 工具函数
│       └── snowflake.go    # 雪花算法ID生成器
├── go.mod                  # Go模块定义
├── go.sum                  # 依赖版本锁定
└── README.md              # 项目文档
```

## 📦 安装

```bash
go get -u github.com/sportsio/sport-common
```

### 依赖要求

- Go 1.21+
- 支持的框架：
    - [go-zero](https://github.com/zeromicro/go-zero) v1.5+
    - [go-kratos](https://github.com/go-kratos/kratos) v2.7+

## 🚀 快速开始

### 1. JWT 认证中间件

#### go-zero 使用方式

```go
import (
    "github.com/sportsio/sport-common/pkg/middleware"
)

func main() {
    // 创建认证中间件
    authMiddleware := middleware.NewAuthMiddleware("your-secret-key")
    
    // 应用到路由
    mux := http.NewServeMux()
    mux.HandleFunc("/api/user/profile", 
        authMiddleware.Handle(userProfileHandler))
}
```

#### go-kratos 使用方式

```go
import (
    "github.com/sportsio/sport-common/pkg/middleware"
    "github.com/go-kratos/kratos/v2/transport/http"
)

func NewHTTPServer() *http.Server {
    authMiddleware := middleware.NewAuthMiddleware("your-secret-key")
    
    opts := []http.ServerOption{
        http.Middleware(
            authMiddleware.Middleware(),
        ),
    }
    
    return http.NewServer(opts...)
}
```

### 2. 带白名单的认证中间件

```go
// 定义不需要认证的路径
whitelist := []string{
    "/api/v1/login",
    "/api/v1/register",
    "/health",
    "/swagger/*",
}

// 创建带白名单的中间件
authMiddleware := middleware.NewAuthMiddlewareWithWhitelist(
    "your-secret-key",
    whitelist,
)
```