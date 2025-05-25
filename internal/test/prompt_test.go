package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ai-gateway-go/internal/test/mocks"
	"github.com/ecodeclub/ai-gateway-go/internal/web"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// PromptTestSuite 是一个测试套件，用于测试提示相关的所有操作
// 包括提示的增删改查、版本管理、发布、删除版本等业务逻辑
type PromptTestSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

// TestPrompt 运行关于提示操作的测试用例
func TestPrompt(t *testing.T) {
	suite.Run(t, new(PromptTestSuite))
}

// SetupSuite 初始化测试环境，配置数据库连接和表结构
func (s *PromptTestSuite) SetupSuite() {
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:13316)/ai_gateway_go?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s"))
	require.NoError(s.T(), err)
	err = dao.InitTable(db)
	require.NoError(s.T(), err)
	s.db = db
	d := dao.NewPromptDAO(db)
	repo := repository.NewPromptRepo(d)
	svc := service.NewPromptService(repo)
	handler := web.NewHandler(svc)
	server := gin.Default()
	handler.PrivateRoutes(server)
	s.server = server
}

// TearDownTest 每个测试用例执行后清理数据
func (s *PromptTestSuite) TearDownTest() {
	err := s.db.Exec("TRUNCATE TABLE prompts").Error
	require.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE prompt_versions").Error
	require.NoError(s.T(), err)
}

// TestAdd 测试新增提示及其初始版本的功能
func (s *PromptTestSuite) TestAdd() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	testCases := []struct {
		name     string
		reqBody  string
		wantCode int
		wantReq  string
		before   func()
		after    func()
	}{
		{
			name: "成功",
			before: func() {
				sess := mocks.NewMockSession(ctrl)
				sess.EXPECT().Claims().Return(session.Claims{
					Uid: 1,
					Data: map[string]string{
						"owner_type": "personal",
					},
				}).AnyTimes()

				provider := mocks.NewMockProvider(ctrl)
				session.SetDefaultProvider(provider)
				provider.EXPECT().Get(gomock.Any()).Return(sess, nil)
			},
			after: func() {
				t := s.T()
				var res dao.Prompt
				err := s.db.Where("id = ?", 1).First(&res).Error
				require.NoError(t, err)
				assert.Equal(t, "test", res.Name)
				assert.Equal(t, "test", res.Description)
				assert.Equal(t, int64(1), res.Owner)
				assert.Equal(t, "personal", res.OwnerType)
				assert.Equal(t, uint8(1), res.Status)
				assert.True(t, res.Ctime > 0)
				assert.True(t, res.Utime > 0)

				var version dao.PromptVersion
				err = s.db.Where("id = ?", 1).First(&version).Error
				require.NoError(t, err)
				assert.Equal(t, int64(1), version.PromptID)
				assert.Equal(t, "", version.Label)
				assert.Equal(t, "test", version.Content)
				assert.Equal(t, "test", version.SystemContent)
				assert.Equal(t, uint8(1), res.Status)
				assert.Equal(t, float32(9.9), version.Temperature)
				assert.Equal(t, float32(0.9), version.TopN)
				assert.Equal(t, 1000, version.MaxTokens)
				assert.True(t, version.Ctime > 0)
				assert.True(t, version.Utime > 0)
			},
			reqBody: `{
  "name": "test",
  "content": "test",
  "description": "test",
  "system_content": "test",
  "temperature": 9.9,
  "top_n":0.9,
  "max_tokens":1000
}`,
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodPost, "/prompt/add", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			tc.after()
		})
	}
}

// TestGet 测试获取提示信息的功能
func (s *PromptTestSuite) TestGet() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	testCases := []struct {
		name       string
		reqBody    string
		wantCode   int
		wantResult web.PromptVO
		wantReq    string
		before     func()
		after      func()
	}{
		{
			name: "成功",
			before: func() {
				now := time.Now().UnixMilli()
				err := s.db.Create(&dao.Prompt{
					Name:          "test",
					Description:   "test",
					Owner:         1,
					OwnerType:     "personal",
					ActiveVersion: 1,
					Status:        1,
					Ctime:         now,
					Utime:         now,
				}).Error
				require.NoError(s.T(), err)
				s.db.Create(&dao.PromptVersion{
					PromptID:      1,
					Label:         "test",
					Content:       "test",
					SystemContent: "test",
					Temperature:   9.9,
					TopN:          0.9,
					MaxTokens:     1000,
					Status:        1,
					Ctime:         now,
					Utime:         now,
				})
			},
			after:    func() {},
			wantCode: http.StatusOK,
			wantResult: web.PromptVO{
				ID:            1,
				Name:          "test",
				Description:   "test",
				Owner:         1,
				OwnerType:     "personal",
				ActiveVersion: 1,
				Versions: []web.PromptVersionVO{{
					ID:            1,
					Label:         "test",
					Content:       "test",
					SystemContent: "test",
					Temperature:   9.9,
					TopN:          0.9,
					MaxTokens:     1000,
					Status:        1,
				}},
			},
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodGet, "/prompt/1", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			var result Result[web.PromptVO]
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.True(t, result.Data.CreateTime > 0)
			assert.True(t, result.Data.UpdateTime > 0)
			result.Data.CreateTime = 0
			result.Data.UpdateTime = 0
			assert.True(t, len(result.Data.Versions) == 1)
			result.Data.Versions[0].CreateTime = 0
			result.Data.Versions[0].UpdateTime = 0
			assert.Equal(t, tc.wantResult, result.Data)
			tc.after()
		})
	}
}

// TestUpdatePrompt 测试更新提示基本信息的功能
func (s *PromptTestSuite) TestUpdatePrompt() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	testCases := []struct {
		name     string
		reqBody  string
		wantCode int
		wantReq  string
		before   func()
		after    func()
	}{
		{
			name: "成功",
			before: func() {
				now := time.Now().UnixMilli()
				err := s.db.Create(&dao.Prompt{
					Name:          "test",
					Description:   "test",
					Owner:         1,
					OwnerType:     "personal",
					ActiveVersion: 1,
					Status:        1,
					Ctime:         now,
					Utime:         now,
				}).Error
				require.NoError(s.T(), err)
			},
			after: func() {
				t := s.T()
				var res dao.Prompt
				err := s.db.Where("id = ?", 1).First(&res).Error
				require.NoError(t, err)
				assert.Equal(t, "aaa", res.Name)
				assert.Equal(t, "aaa", res.Description)
				assert.Equal(t, int64(1), res.Owner)
				assert.Equal(t, "personal", res.OwnerType)
				assert.Equal(t, uint8(1), res.Status)
				assert.True(t, res.Ctime > 0)
				assert.True(t, res.Utime > res.Ctime)
			},
			reqBody: `{
				"id": 1,
				"name": "aaa",
				"description": "aaa"
			}`,
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodPost, "/prompt/update", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			tc.after()
		})
	}
}

// TestUpdateVersion 测试更新提示版本信息的功能
func (s *PromptTestSuite) TestUpdateVersion() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	testCases := []struct {
		name     string
		reqBody  string
		wantCode int
		wantReq  string
		before   func()
		after    func()
	}{
		{
			name: "成功",
			before: func() {
				now := time.Now().UnixMilli()
				err := s.db.Create(&dao.PromptVersion{
					PromptID:      1,
					Label:         "test",
					Content:       "test",
					SystemContent: "test",
					Temperature:   9.9,
					TopN:          0.9,
					MaxTokens:     1000,
					Status:        1,
					Ctime:         now,
					Utime:         now,
				}).Error
				require.NoError(s.T(), err)
			},
			after: func() {
				t := s.T()
				var version dao.PromptVersion
				err := s.db.Where("id = ?", 1).First(&version).Error
				require.NoError(t, err)
				assert.Equal(t, int64(1), version.PromptID)
				assert.Equal(t, "aaa", version.Content)
				assert.Equal(t, "test", version.Label)
				assert.Equal(t, "test", version.SystemContent)
				assert.Equal(t, uint8(1), version.Status)
				assert.Equal(t, float32(1.0), version.TopN)
				assert.True(t, version.Utime > version.Ctime)
			},
			reqBody: `{
				"version_id": 1,
				"content": "aaa",
				"top_n": 1.0
			}`,
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodPost, "/prompt/update/version", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			tc.after()
		})
	}
}

// TestDelete 测试删除整个提示（软删除）的功能
func (s *PromptTestSuite) TestDelete() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	testCases := []struct {
		name     string
		reqBody  string
		wantCode int
		wantReq  string
		before   func()
		after    func()
	}{
		{
			name: "删除整个prompt-成功",
			before: func() {
				now := time.Now().UnixMilli()
				err := s.db.Create(&dao.Prompt{
					Name:          "test",
					Description:   "test",
					Owner:         1,
					OwnerType:     "personal",
					ActiveVersion: 1,
					Status:        1,
					Ctime:         now,
					Utime:         now,
				}).Error
				require.NoError(s.T(), err)
				s.db.Create(&dao.PromptVersion{
					PromptID:      1,
					Label:         "test",
					Content:       "test",
					SystemContent: "test",
					Temperature:   9.9,
					TopN:          0.9,
					MaxTokens:     1000,
					Status:        1,
					Ctime:         now,
					Utime:         now,
				})
			},
			after: func() {
				t := s.T()
				var res dao.Prompt
				err := s.db.Where("id = ?", 1).First(&res).Error
				require.NoError(t, err)
				assert.Equal(t, uint8(0), res.Status)
				assert.True(t, res.Utime > res.Ctime)

				var version dao.PromptVersion
				err = s.db.Where("id = ?", 1).First(&version).Error
				require.NoError(t, err)
				assert.Equal(t, uint8(0), version.Status)
				assert.True(t, res.Utime > version.Ctime)
			},
			reqBody: `{
				"id": 1
			}`,
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodPost, "/prompt/delete", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			tc.after()
		})
	}
}

// TestDeleteVersion 测试删除特定提示版本的功能
func (s *PromptTestSuite) TestDeleteVersion() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	testCases := []struct {
		name     string
		reqBody  string
		wantCode int
		wantReq  string
		before   func()
		after    func()
	}{
		{
			name: "删除version-成功",
			before: func() {
				now := time.Now().UnixMilli()
				err := s.db.Create(&dao.Prompt{
					Name:          "test",
					Description:   "test",
					Owner:         1,
					OwnerType:     "personal",
					ActiveVersion: 1,
					Status:        1,
					Ctime:         now,
					Utime:         now,
				}).Error
				require.NoError(s.T(), err)
				s.db.Create(&dao.PromptVersion{
					PromptID:      1,
					Label:         "test",
					Content:       "test",
					SystemContent: "test",
					Temperature:   9.9,
					TopN:          0.9,
					MaxTokens:     1000,
					Status:        1,
					Ctime:         now,
					Utime:         now,
				})
			},
			after: func() {
				t := s.T()
				var res dao.Prompt
				err := s.db.Where("id = ?", 1).First(&res).Error
				require.NoError(t, err)
				assert.Equal(t, uint8(1), res.Status)
				assert.True(t, res.Utime == res.Ctime)

				var version dao.PromptVersion
				err = s.db.Where("id = ?", 1).First(&version).Error
				require.NoError(t, err)
				assert.Equal(t, uint8(0), version.Status)
				assert.True(t, version.Utime > version.Ctime)
			},
			reqBody: `{
				"version_id": 1
			}`,
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodPost, "/prompt/delete/version", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			tc.after()
		})
	}
}

// TestPublish 测试发布特定提示版本的功能
func (s *PromptTestSuite) TestPublish() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	testCases := []struct {
		name     string
		reqBody  string
		wantCode int
		wantReq  string
		before   func()
		after    func()
	}{
		{
			name: "成功",
			before: func() {
				now := time.Now().UnixMilli()
				err := s.db.Create(&dao.Prompt{
					Name:          "test",
					Description:   "test",
					Owner:         1,
					OwnerType:     "personal",
					ActiveVersion: 0, // 没有发布版本
					Status:        1,
					Ctime:         now,
					Utime:         now,
				}).Error
				require.NoError(s.T(), err)
				s.db.Create(&dao.PromptVersion{
					PromptID:      1,
					Label:         "", // 没有语义化版本
					Content:       "test",
					SystemContent: "test",
					Temperature:   9.9,
					TopN:          0.9,
					MaxTokens:     1000,
					Status:        1,
					Ctime:         now,
					Utime:         now,
				})
			},
			after: func() {
				t := s.T()
				var res dao.Prompt
				err := s.db.Where("id = ?", 1).First(&res).Error
				require.NoError(t, err)
				assert.True(t, res.ActiveVersion == 1)
				assert.True(t, res.Utime > res.Ctime)

				var version dao.PromptVersion
				err = s.db.Where("id = ?", 1).First(&version).Error
				require.NoError(t, err)
				assert.Equal(t, int64(1), version.PromptID)
				assert.Equal(t, "v1.0.0", version.Label)
				assert.True(t, version.Utime > version.Ctime)
			},
			reqBody: `{
				"version_id": 1,
				"label": "v1.0.0"
			}`,
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodPost, "/prompt/publish", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			tc.after()
		})
	}
}

// TestFork 测试复制（fork）提示版本的功能
func (s *PromptTestSuite) TestFork() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	testCases := []struct {
		name     string
		reqBody  string
		wantCode int
		wantReq  string
		before   func()
		after    func()
	}{
		{
			name: "成功",
			before: func() {
				now := time.Now().UnixMilli()
				err := s.db.Create(&dao.Prompt{
					Name:          "test",
					Description:   "test",
					Owner:         1,
					OwnerType:     "personal",
					ActiveVersion: 1,
					Status:        1,
					Ctime:         now,
					Utime:         now,
				}).Error
				require.NoError(s.T(), err)
				s.db.Create(&dao.PromptVersion{
					PromptID:      1,
					Label:         "v1.0.0",
					Content:       "test",
					SystemContent: "test",
					Temperature:   9.9,
					TopN:          0.9,
					MaxTokens:     1000,
					Status:        1,
					Ctime:         now,
					Utime:         now,
				})
			},
			after: func() {
				t := s.T()
				var versions []dao.PromptVersion
				err := s.db.Where("prompt_id = ?", 1).Find(&versions).Error
				require.NoError(t, err)
				require.True(t, len(versions) == 2) // fork 之后有两个版本
				assert.Equal(t, "test", versions[0].Content)
				assert.Equal(t, "test", versions[0].SystemContent)
				assert.Equal(t, "v1.0.0", versions[0].Label)
				assert.Equal(t, float32(0.9), versions[0].TopN)
				assert.Equal(t, float32(9.9), versions[0].Temperature)
				assert.Equal(t, 1000, versions[0].MaxTokens)
				assert.True(t, versions[0].Ctime > 0)
				assert.True(t, versions[0].Utime > 0)

				assert.Equal(t, "test", versions[1].Content)
				assert.Equal(t, "test", versions[1].SystemContent)
				// fork 后的 label 为空
				assert.Equal(t, "", versions[1].Label)
				assert.Equal(t, float32(0.9), versions[1].TopN)
				assert.Equal(t, float32(9.9), versions[1].Temperature)
				assert.Equal(t, 1000, versions[1].MaxTokens)
				assert.True(t, versions[0].Ctime > 0)
				assert.True(t, versions[0].Utime > 0)
			},
			reqBody: `{
				"version_id": 1
			}`,
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodPost, "/prompt/fork", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			tc.after()
		})
	}
}

// Result 定义了统一的返回结果格式
// 用于测试中解析 HTTP 响应结果
type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
