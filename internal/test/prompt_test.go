package test

import (
	"bytes"
	"encoding/json"
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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type PromptTestSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

func TestPrompt(t *testing.T) {
	suite.Run(t, new(PromptTestSuite))
}

func (s *PromptTestSuite) SetupSuite() {
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:13316)/ai_gateway?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s"))
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

func (s *PromptTestSuite) TearDownTest() {
	err := s.db.Exec("TRUNCATE TABLE prompts").Error
	require.NoError(s.T(), err)
}

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
				assert.Equal(t, "test", res.Content)
				assert.Equal(t, "test", res.Description)
				assert.Equal(t, int64(1), res.Owner)
				assert.Equal(t, "personal", res.OwnerType)
				assert.Equal(t, uint8(1), res.Status)
				assert.True(t, res.Ctime > 0)
				assert.True(t, res.Utime > 0)
			},
			reqBody: `{
				"name": "test",
				"content": "test",
				"description": "test"
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

func (s *PromptTestSuite) TestGet() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	testCases := []struct {
		name       string
		reqBody    string
		wantCode   int
		wantResult web.GetVO
		wantReq    string
		before     func()
		after      func()
	}{
		{
			name: "成功",
			before: func() {
				now := time.Now().UnixMilli()
				err := s.db.Create(&dao.Prompt{
					Name:        "test",
					Content:     "test",
					Description: "test",
					Owner:       1,
					OwnerType:   "personal",
					Status:      1,
					Ctime:       now,
					Utime:       now,
				}).Error
				require.NoError(s.T(), err)
			},
			after: func() {

			},
			wantCode: http.StatusOK,
			wantResult: web.GetVO{
				Name:        "test",
				Content:     "test",
				Description: "test",
				Owner:       1,
				OwnerType:   "personal",
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
			var result Result[web.GetVO]
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.True(t, result.Data.CreateTime > 0)
			assert.True(t, result.Data.UpdateTime > 0)
			result.Data.CreateTime = 0
			result.Data.UpdateTime = 0
			assert.Equal(t, tc.wantResult, result.Data)
			tc.after()
		})
	}
}

func (s *PromptTestSuite) TestUpdate() {
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
					Name:        "test",
					Content:     "test",
					Description: "test",
					Owner:       1,
					OwnerType:   "personal",
					Status:      1,
					Ctime:       now,
					Utime:       now,
				}).Error
				require.NoError(s.T(), err)
			},
			after: func() {
				t := s.T()
				var res dao.Prompt
				err := s.db.Where("id = ?", 1).First(&res).Error
				require.NoError(t, err)
				assert.Equal(t, "aaa", res.Name)
				assert.Equal(t, "aaa", res.Content)
				assert.Equal(t, "aaa", res.Description)
				assert.Equal(t, int64(1), res.Owner)
				assert.Equal(t, "personal", res.OwnerType)
				assert.Equal(t, uint8(1), res.Status)
				assert.True(t, res.Ctime > 0)
				assert.True(t, res.Utime > res.Ctime)
			},
			reqBody: `{
				"name": "aaa",
				"content": "aaa",
				"description": "aaa"
			}`,
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodPost, "/prompt/1", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			tc.after()
		})
	}
}

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
			name: "成功",
			before: func() {
				now := time.Now().UnixMilli()
				err := s.db.Create(&dao.Prompt{
					Name:        "test",
					Content:     "test",
					Description: "test",
					Owner:       1,
					OwnerType:   "personal",
					Status:      1,
					Ctime:       now,
					Utime:       now,
				}).Error
				require.NoError(s.T(), err)
			},
			after: func() {
				t := s.T()
				var res dao.Prompt
				err := s.db.Where("id = ?", 1).First(&res).Error
				require.NoError(t, err)
				assert.Equal(t, uint8(0), res.Status)
				assert.True(t, res.Utime > res.Ctime)
			},
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before()
			req, err := http.NewRequest(http.MethodDelete, "/prompt/1", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			tc.after()
		})
	}
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
