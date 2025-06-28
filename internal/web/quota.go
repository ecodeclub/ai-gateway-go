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

package web

import (
	"time"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

type QuotaHandler struct {
	svc *service.QuotaService
}

func NewQuotaHandler(svc *service.QuotaService) *QuotaHandler {
	return &QuotaHandler{svc: svc}
}

func (q *QuotaHandler) PrivateRoutes(_ *gin.Engine) {}

func (q *QuotaHandler) PublicRoutes(server *gin.Engine) {
	group := server.Group("/quota")
	group.POST("/create", ginx.BS(q.CreateTempQuota))
	group.POST("/create_tmp", ginx.BS(q.CreateTempQuota))
	group.POST("/deduct", ginx.BS(q.Deduct))
	group.POST("/get", ginx.S(q.GetQuota))
	group.POST("/get_tmp", ginx.S(q.GetTempQuota))
}

func (q *QuotaHandler) CreateQuota(ctx *ginx.Context, req QuotaRequest, sess session.Session) (ginx.Result, error) {
	uid := sess.Claims().Uid
	err := q.svc.CreateQuota(ctx, domain.Quota{Amount: req.Amount, Uid: uid})
	if err != nil {
		return systemErrorResult, nil
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (q *QuotaHandler) CreateTempQuota(ctx *ginx.Context, req QuotaRequest, sess session.Session) (ginx.Result, error) {
	uid := sess.Claims().Uid

	if req.StartTime == "" || req.EndTime == "" {
		return systemErrorResult, nil
	}

	start, _ := q.toTimestamp(req.StartTime)
	end, _ := q.toTimestamp(req.EndTime)

	err := q.svc.CreateTempQuota(ctx, domain.TempQuota{Amount: req.Amount, Uid: uid, StartTime: start, EndTime: end})
	if err != nil {
		return systemErrorResult, nil
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (q *QuotaHandler) GetQuota(ctx *ginx.Context, sees session.Session) (ginx.Result, error) {
	uid := sees.Claims().Uid

	quota, err := q.svc.GetQuota(ctx, uid)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg:  "ok",
		Data: QuotaResponse{Amount: quota.Amount},
	}, nil
}

func (q *QuotaHandler) GetTempQuota(ctx *ginx.Context, sees session.Session) (ginx.Result, error) {
	uid := sees.Claims().Uid
	quotaList, err := q.svc.GetTempQuota(ctx, uid)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg:  "ok",
		Data: q.toQuotaResponse(quotaList),
	}, nil
}

func (q *QuotaHandler) UpdateQuota(ctx *ginx.Context, req QuotaRequest, sees session.Session) (ginx.Result, error) {
	uid := sees.Claims().Uid

	err := q.svc.UpdateQuota(ctx, domain.Quota{Uid: uid, Amount: req.Amount})
	if err != nil {
		return systemErrorResult, nil
	}

	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (q *QuotaHandler) Deduct(ctx *ginx.Context, req QuotaRequest, sees session.Session) (ginx.Result, error) {
	uid := sees.Claims().Uid
	err := q.svc.Deduct(ctx, uid, req.Amount)
	if err != nil {
		return systemErrorResult, nil
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (q *QuotaHandler) toTimestamp(timeStr string) (int64, error) {
	const layout = "2006-01-02 15:04:05"
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func (q *QuotaHandler) toQuotaResponse(tempQuotaList []domain.TempQuota) []QuotaResponse {
	return slice.Map[domain.TempQuota, QuotaResponse](tempQuotaList, func(idx int, src domain.TempQuota) QuotaResponse {
		return QuotaResponse{
			Amount:    src.Amount,
			StartTime: time.Unix(src.StartTime, 0).Format("2006-01-02 15:04:05"),
			EndTime:   time.Unix(src.EndTime, 0).Format("2006-01-02 15:04:05"),
		}
	})
}

type QuotaRequest struct {
	Amount    int64  `json:"amount,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
}

type QuotaResponse struct {
	Amount    int64  `json:"amount,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
}
