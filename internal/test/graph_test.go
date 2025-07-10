// Copyright 2021 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
  "bytes"
  "encoding/json"
  "errors"
  "net/http"
  "net/http/httptest"
  "testing"
  "time"

  "github.com/ecodeclub/ai-gateway-go/internal/repository"
  "github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
  "github.com/ecodeclub/ai-gateway-go/internal/service"
  "github.com/ecodeclub/ai-gateway-go/internal/test/mocks"
  "github.com/ecodeclub/ginx/session"
  "github.com/gin-gonic/gin"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "github.com/stretchr/testify/suite"
  "github.com/yumosx/got/pkg/config"
  "go.uber.org/mock/gomock"
  "gorm.io/gorm"
)

type GraphTestSuite struct {
  suite.Suite
  db     *gorm.DB
  server *gin.Engine
}

func TestNode(t *testing.T) {
  suite.Run(t, new(GraphTestSuite))
}

func (n *GraphTestSuite) SetupSuite() {
  dbConfig := config.NewConfig(
    config.WithDBName("ai_gateway_platform"),
    config.WithUserName("root"),
    config.WithPassword("root"),
    config.WithHost("127.0.0.1"),
    config.WithPort("13306"),
  )

  db, err := config.NewDB(dbConfig)
  require.NoError(n.T(), err)
  err = dao.InitGraphTable(db)
  require.NoError(n.T(), err)
  n.db = db
  d := dao.NewGraphDAO(db)
  repo := repository.NewGraphRepo(d)
  svc := service.NewGraphService(repo)
  handler := admin.NewGraphHandler(svc)
  server := gin.Default()
  handler.PrivateRoutes(server)
  n.server = server
}

func (n *GraphTestSuite) TearDownTest() {
  err := n.db.Exec("TRUNCATE TABLE edges").Error
  require.NoError(n.T(), err)
  err = n.db.Exec("TRUNCATE TABLE nodes").Error
  require.NoError(n.T(), err)
  err = n.db.Exec("TRUNCATE TABLE graphs").Error
  require.NoError(n.T(), err)
}

func (n *GraphTestSuite) TestSaveNode() {
  t := n.T()

  ctrl := gomock.NewController(t)
  defer ctrl.Finish()

  testcases := []struct {
    name     string
    before   func()
    after    func()
    reqBody  string
    wantCode int
  }{
    {
      name: "创建 node",
      before: func() {
        sess := mocks.NewMockSession(ctrl)
        sess.EXPECT().Claims().Return(session.Claims{
          Uid:  1,
          Data: map[string]string{},
        }).AnyTimes()
        provider := mocks.NewMockProvider(ctrl)
        session.SetDefaultProvider(provider)
        provider.EXPECT().Get(gomock.Any()).Return(sess, nil)
      },
      after: func() {
        var node dao.Node
        err := n.db.Where("id = ?", 1).First(&node).Error
        require.NoError(t, err)
        assert.Equal(t, int64(1), node.GraphID)
        assert.True(t, node.Utime > 0)
        assert.True(t, node.Ctime > 0)
      },
      reqBody:  `{"graph_id": 1}`,
      wantCode: http.StatusOK,
    },
    {
      name: "更新 node",
      before: func() {
        sess := mocks.NewMockSession(ctrl)
        sess.EXPECT().Claims().Return(session.Claims{
          Uid:  1,
          Data: map[string]string{},
        }).AnyTimes()
        provider := mocks.NewMockProvider(ctrl)
        session.SetDefaultProvider(provider)
        provider.EXPECT().Get(gomock.Any()).Return(sess, nil)
        // 插入一条数据
        err := n.db.Create(&dao.Node{GraphID: 1, Ctime: time.Now().UnixMilli(), Utime: time.Now().UnixMilli()}).Error
        require.NoError(t, err)
      },
      after: func() {
        var node dao.Node
        err := n.db.Where("id = ?", 1).First(&node).Error
        require.NoError(t, err)
        assert.Equal(t, int64(2), node.GraphID)
        assert.True(t, node.Utime > 0)
        assert.True(t, node.Ctime > 0)
      },
      reqBody:  `{"id": 1, "graph_id": 2}`,
      wantCode: http.StatusOK,
    },
  }
  for _, tc := range testcases {
    t.Run(tc.name, func(t *testing.T) {
      tc.before()
      req, err := http.NewRequest(http.MethodPost, "/node/save", bytes.NewBuffer([]byte(tc.reqBody)))
      require.NoError(t, err)
      req.Header.Set("Content-Type", "application/json")
      resp := httptest.NewRecorder()
      n.server.ServeHTTP(resp, req)
      tc.after()
    })
  }
}

func (n *GraphTestSuite) TestSaveEdge() {
  t := n.T()

  ctrl := gomock.NewController(t)
  defer ctrl.Finish()

  testcases := []struct {
    name     string
    before   func()
    after    func()
    reqBody  string
    wantCode int
  }{
    {
      name: "创建 edge",
      before: func() {
        sess := mocks.NewMockSession(ctrl)
        sess.EXPECT().Claims().Return(session.Claims{
          Uid:  1,
          Data: map[string]string{},
        }).AnyTimes()
        provider := mocks.NewMockProvider(ctrl)
        session.SetDefaultProvider(provider)
        provider.EXPECT().Get(gomock.Any()).Return(sess, nil)
      },
      after: func() {
        var edge dao.Edge
        err := n.db.Where("id = ?", 1).First(&edge).Error
        require.NoError(t, err)
        assert.Equal(t, int64(1), edge.GraphID)
        assert.Equal(t, int64(1), edge.SourceID)
        assert.Equal(t, int64(2), edge.TargetID)
        assert.True(t, edge.Utime > 0)
        assert.True(t, edge.Ctime > 0)
      },
      reqBody:  `{"graph_id": 1, "source_id": 1, "target_id": 2}`,
      wantCode: http.StatusOK,
    },
    {
      name: "更新edge",
      before: func() {
        sess := mocks.NewMockSession(ctrl)
        sess.EXPECT().Claims().Return(session.Claims{
          Uid:  1,
          Data: map[string]string{},
        }).AnyTimes()
        provider := mocks.NewMockProvider(ctrl)
        session.SetDefaultProvider(provider)
        provider.EXPECT().Get(gomock.Any()).Return(sess, nil)
        // 插入一条数据
        err := n.db.Create(&dao.Edge{GraphID: 1, SourceID: 1, TargetID: 2, Ctime: time.Now().UnixMilli(), Utime: time.Now().UnixMilli()}).Error
        require.NoError(t, err)
      },
      after: func() {
        var edge dao.Edge
        err := n.db.Where("id = ?", 1).First(&edge).Error
        require.NoError(t, err)
        assert.Equal(t, int64(1), edge.GraphID)
        assert.Equal(t, int64(2), edge.SourceID)
        assert.Equal(t, int64(3), edge.TargetID)

        assert.True(t, edge.Ctime > 0)
        assert.True(t, edge.Utime > 0)
      },
      reqBody:  `{"id": 1, "graph_id": 1, "source_id": 2, "target_id": 3}`,
      wantCode: http.StatusOK,
    },
  }
  for _, tc := range testcases {
    t.Run(tc.name, func(t *testing.T) {
      tc.before()
      req, err := http.NewRequest(http.MethodPost, "/edge/save", bytes.NewBuffer([]byte(tc.reqBody)))
      require.NoError(t, err)
      req.Header.Set("Content-Type", "application/json")
      resp := httptest.NewRecorder()
      n.server.ServeHTTP(resp, req)
      tc.after()
    })
  }
}

func (n *GraphTestSuite) TestGetGraph() {
  t := n.T()

  ctrl := gomock.NewController(t)
  defer ctrl.Finish()

  testcases := []struct {
    name    string
    before  func()
    reqBody string
  }{
    {
      name: "获取graph",
      before: func() {
        sess := mocks.NewMockSession(ctrl)
        sess.EXPECT().Claims().Return(session.Claims{
          Uid:  1,
          Data: map[string]string{},
        }).AnyTimes()
        provider := mocks.NewMockProvider(ctrl)
        session.SetDefaultProvider(provider)
        provider.EXPECT().Get(gomock.Any()).Return(sess, nil)

        err := n.db.Create(&dao.Node{GraphID: 1, Ctime: time.Now().UnixMilli(), Utime: time.Now().UnixMilli()}).Error
        require.NoError(t, err)
        err = n.db.Create(&dao.Edge{GraphID: 1, SourceID: 1, TargetID: 2, Ctime: time.Now().UnixMilli(), Utime: time.Now().UnixMilli()}).Error
        require.NoError(t, err)
        err = n.db.Create(&dao.Graph{Metadata: "test"}).Error
        require.NoError(t, err)
      },
      reqBody: `{"id": 1}`,
    },
  }
  for _, tc := range testcases {
    t.Run(tc.name, func(t *testing.T) {
      tc.before()
      req, err := http.NewRequest(http.MethodPost, "/graph/detail", bytes.NewBuffer([]byte(tc.reqBody)))
      require.NoError(t, err)
      req.Header.Set("Content-Type", "application/json")
      resp := httptest.NewRecorder()
      n.server.ServeHTTP(resp, req)
      var result Result[admin.GraphVO]
      err = json.NewDecoder(resp.Body).Decode(&result)
      require.NoError(t, err)
      assert.Equal(t, len(result.Data.Edges), 1)
      assert.Equal(t, len(result.Data.Nodes), 1)
      assert.Equal(t, int(result.Data.Edges[0].ID), 1)
      assert.Equal(t, int(result.Data.Nodes[0].ID), 1)
    })
  }
}

func (n *GraphTestSuite) TestDeleteNode() {
  t := n.T()
  ctrl := gomock.NewController(t)
  defer ctrl.Finish()
  testcases := []struct {
    name    string
    before  func()
    after   func()
    reqBody string
  }{
    {
      name: "删除node",
      before: func() {
        sess := mocks.NewMockSession(ctrl)
        sess.EXPECT().Claims().Return(session.Claims{
          Uid: 1,
        }).AnyTimes()
        provider := mocks.NewMockProvider(ctrl)
        session.SetDefaultProvider(provider)
        provider.EXPECT().Get(gomock.Any()).Return(sess, nil)
        err := n.db.Create(&dao.Node{GraphID: 1, Ctime: time.Now().UnixMilli(), Utime: time.Now().UnixMilli()}).Error
        require.NoError(t, err)
      },
      after: func() {
        var node dao.Node
        err := n.db.Where("id = ?", 1).First(&node).Error
        require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
      },
      reqBody: `{"id": 1}`,
    },
  }

  for _, tc := range testcases {
    t.Run(tc.name, func(t *testing.T) {
      tc.before()
      req, err := http.NewRequest(http.MethodPost, "/node/delete", bytes.NewBuffer([]byte(tc.reqBody)))
      require.NoError(t, err)
      req.Header.Set("Content-Type", "application/json")
      resp := httptest.NewRecorder()
      n.server.ServeHTTP(resp, req)
      tc.after()
    })
  }
}

func (n *GraphTestSuite) TestDeleteEdge() {
  t := n.T()
  ctrl := gomock.NewController(t)
  defer ctrl.Finish()
  testcases := []struct {
    name    string
    before  func()
    after   func()
    reqBody string
  }{
    {
      name: "删除edge",
      before: func() {
        sess := mocks.NewMockSession(ctrl)
        sess.EXPECT().Claims().Return(session.Claims{
          Uid: 1,
        }).AnyTimes()
        provider := mocks.NewMockProvider(ctrl)
        session.SetDefaultProvider(provider)
        provider.EXPECT().Get(gomock.Any()).Return(sess, nil)
        err := n.db.Create(&dao.Edge{GraphID: 1, SourceID: 1, TargetID: 2, Ctime: time.Now().UnixMilli(), Utime: time.Now().UnixMilli()}).Error
        require.NoError(t, err)
        var edge1 dao.Edge
        n.db.Create(&edge1)
        require.NoError(t, err)
      },
      after: func() {
        var edge dao.Edge
        err := n.db.Where("id = ?", 1).First(&edge).Error
        require.Equal(t, err, gorm.ErrRecordNotFound)
      },
      reqBody: `{"id": 1}`,
    },
  }

  for _, tc := range testcases {
    t.Run(tc.name, func(t *testing.T) {
      tc.before()
      req, err := http.NewRequest(http.MethodPost, "/edge/delete", bytes.NewBuffer([]byte(tc.reqBody)))
      require.NoError(t, err)
      req.Header.Set("Content-Type", "application/json")
      resp := httptest.NewRecorder()
      n.server.ServeHTTP(resp, req)
      tc.after()
    })
  }
}
