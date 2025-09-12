package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

const (
	// ContextUserIDKey 用户ID在context中的key
	ContextUserIDKey = "user_id"
	// BearerPrefix Bearer token前缀
	BearerPrefix = "Bearer "
)

type AuthMiddleware struct {
	secret string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{
		secret: secret,
	}
}

// Handle 用于go-zero框架的HTTP中间件
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从Header获取Token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeErrorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// 验证token并获取用户ID
		userID, err := m.validateTokenAndGetUserID(authHeader)
		if err != nil {
			writeErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		// 将用户ID存入上下文
		ctx := context.WithValue(r.Context(), ContextUserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// Middleware 用于go-kratos框架的中间件
func (m *AuthMiddleware) Middleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 获取transport信息
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, errors.New(401, "unauthorized", "未授权")
			}

			// 根据不同的transport类型获取token
			var token string
			switch tr.Kind() {
			case transport.KindHTTP:
				token = tr.RequestHeader().Get("Authorization")
			case transport.KindGRPC:
				token = tr.RequestHeader().Get("authorization")
			default:
				return nil, errors.New(401, "unauthorized", "unsupported transfer protocol")
			}

			if token == "" {
				return nil, errors.New(401, "unauthorized", "token is empty")
			}

			// 验证token并获取用户ID
			userID, err := m.validateTokenAndGetUserID(token)
			if err != nil {
				return nil, errors.New(401, "unauthorized", err.Error())
			}

			// 将用户ID存入上下文
			ctx = context.WithValue(ctx, ContextUserIDKey, userID)
			return handler(ctx, req)
		}
	}
}

// validateTokenAndGetUserID 验证token并返回用户ID
func (m *AuthMiddleware) validateTokenAndGetUserID(authHeader string) (uint64, error) {
	// 解析Token
	tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
	if tokenString == authHeader {
		// 如果没有Bearer前缀，直接使用原始token
		tokenString = authHeader
	}

	claims, err := ParseToken(tokenString, m.secret)
	if err != nil {
		return 0, err
	}

	userID := claims.UserID

	return userID, nil
}

// GetUserIDFromContext 从context中获取用户ID
func GetUserIDFromContext(ctx context.Context) (uint64, error) {
	userID, ok := ctx.Value(ContextUserIDKey).(uint64)
	if !ok {
		return 0, fmt.Errorf("用户ID不存在")
	}
	return userID, nil
}

// writeErrorResponse 写入错误响应（用于go-zero）
func writeErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"code": %d, "message": "%s"}`, code, message)))
}

// 白名单路由处理
type AuthMiddlewareWithWhitelist struct {
	*AuthMiddleware
	whitelist map[string]bool // 路径白名单
}

func NewAuthMiddlewareWithWhitelist(secret string, whitelist []string) *AuthMiddlewareWithWhitelist {
	wl := make(map[string]bool)
	for _, path := range whitelist {
		wl[path] = true
	}
	return &AuthMiddlewareWithWhitelist{
		AuthMiddleware: NewAuthMiddleware(secret),
		whitelist:      wl,
	}
}

// Middleware 带白名单的Kratos中间件
func (m *AuthMiddlewareWithWhitelist) Middleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			// 检查是否在白名单中
			if tr.Kind() == transport.KindHTTP {
				if ht, ok := tr.(transport.Transporter); ok {
					path := ht.Operation()
					if m.whitelist[path] {
						return handler(ctx, req)
					}
				}
			}

			// 不在白名单中，执行认证
			return m.AuthMiddleware.Middleware()(handler)(ctx, req)
		}
	}
}

// 使用示例：

// 1. 在go-zero中使用:
// func setupGoZeroMiddleware(mux *http.ServeMux, authMiddleware *AuthMiddleware) {
//     // 应用到需要认证的路由
//     mux.HandleFunc("/api/user/profile", authMiddleware.Handle(userProfileHandler))
// }

// 2. 在go-kratos HTTP服务中使用:
// func NewHTTPServer(authMiddleware *AuthMiddleware) *http.Server {
//     var opts = []http.ServerOption{
//         http.Middleware(
//             authMiddleware.Middleware(),
//         ),
//     }
//     srv := http.NewServer(opts...)
//     return srv
// }

// 3. 在go-kratos gRPC服务中使用:
// func NewGRPCServer(authMiddleware *AuthMiddleware) *grpc.Server {
//     var opts = []grpc.ServerOption{
//         grpc.Middleware(
//             authMiddleware.Middleware(),
//         ),
//     }
//     srv := grpc.NewServer(opts...)
//     return srv
// }
