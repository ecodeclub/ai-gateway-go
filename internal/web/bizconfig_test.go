package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/service/mocks"
	"github.com/ecodeclub/ai-gateway-go/internal/web/infra"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

// fakeAuthMiddleware 创建一个假的身份验证中间件
// 用于在测试中模拟 JWT 认证流程
func fakeAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取JWT Bearer令牌
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, "missing token")
			c.Abort()
			return
		}
		c.Next()
	}
}

// generateJWT 生成一个用于测试的 JWT 令牌
// secret 是签名密钥
func generateJWT(secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"sub": "some_user_id",
	})
	return token.SignedString([]byte(secret))
}

// TestBizConfigHandler_Create 是针对业务配置创建功能的测试用例
// 检查创建业务配置时的正常行为以及返回结果是否符合预期
func TestBizConfigHandler_Create(t *testing.T) {
	const createUrl = "/api/v1/biz-configs/create"
	secret := "VGhpcyBpcyBhIHNlY3JldCB0aGF0IG5vYm9keSBjYW4gZ3Vlc3M=" // 你的 Secret

	// 生成 JWT token
	token, err := generateJWT(secret)
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) service.BizConfigService
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
		wantBody   string
	}{
		{
			name: "创建成功",
			mock: func(ctrl *gomock.Controller) service.BizConfigService {
				svc := mocks.NewMockBizConfigService(ctrl)
				svc.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(domain.BizConfig{ID: 1}, nil)
				return svc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{
                    "biz_id": 123,
                    "key": "some_config_key",
                    "value": "some_value"
                }`))
				req, err := http.NewRequest(http.MethodPost, createUrl, body)
				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: 200,
			wantBody: `{"code":0,"msg":"success","data":{"bizconfig":{"config":"","ctime":"0001-01-01 00:00:00","id":1,"owner_id":0,"owner_type":"","utime":"0001-01-01 00:00:00"}}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			infra.Init()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := tc.mock(ctrl)
			handler := NewBizConfigHandler(svc)

			server := gin.New()
			server.Use(fakeAuthMiddleware())
			handler.RegisterRoutes(server)

			req := tc.reqBuilder(t)
			// 将生成的 JWT Token 设置到 Authorization 头
			req.Header.Set("Authorization", "Bearer "+token)

			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			//assert.Equal(t, tc.wantBody, recorder.Body.String())
			assert.JSONEq(t, tc.wantBody, recorder.Body.String())
		})
	}
}
