package middleware

import (
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

// AuthInterceptor 是一个 gRPC 拦截器，用于验证请求的身份认证信息
func AuthInterceptor(svc *service.BizConfigService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 排除登录等不需要认证的接口
		if info.FullMethod == "/ai.v1.BizConfigService/CreateBizConfig" {
			return handler(ctx, req)
		}

		// 从请求的上下文中获取 metadata（元数据）
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		// 从 metadata 中提取出 'authorization' 头部，包含认证信息（通常是 token）
		authHeader, ok := md["authorization"]
		if !ok || len(authHeader) == 0 {
			// 如果没有找到 authorization 头部，说明没有携带认证 token，返回认证失败
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		// 去掉 token 前缀的 "Bearer "，得到实际的 token 字符串
		tokenString := strings.TrimPrefix(authHeader[0], "Bearer ")

		// 使用 svc.ValidateToken 函数来验证 token 的合法性
		token, err := svc.ValidateToken(tokenString)
		if err != nil {
			// 如果验证失败，返回认证失败错误
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// 尝试从 token 中提取 claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			// 如果提取失败或 token 无效，返回认证失败错误
			return nil, status.Errorf(codes.Unauthenticated, "invalid token claims")
		}

		// 将claims添加到上下文
		newCtx := context.WithValue(ctx, "claims", claims)

		// 继续调用下一个拦截器或处理请求，传递更新后的上下文和请求
		return handler(newCtx, req)
	}
}
