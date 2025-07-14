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
	"errors"

	"github.com/ecodeclub/ai-gateway-go/errs"
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

func (q *QuotaHandler) PrivateRoutes(server *gin.Engine) {
	group := server.Group("/quota")
	group.POST("/save", ginx.BS(q.AddQuota))
	group.POST("/get", ginx.S(q.GetQuota))

	tmp := server.Group("/tmp")
	tmp.POST("/save", ginx.BS(q.CreateTempQuota))
	tmp.POST("/get", ginx.S(q.GetTempQuota))

	server.POST("/deduct", ginx.BS(q.Deduct))
}

func (q *QuotaHandler) AddQuota(ctx *ginx.Context, req QuotaRequest, sess session.Session) (ginx.Result, error) {
	uid := sess.Claims().Uid
	err := q.svc.AddQuota(ctx, domain.Quota{Amount: req.Amount, Uid: uid, Key: req.Key})
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (q *QuotaHandler) CreateTempQuota(ctx *ginx.Context, req QuotaRequest, sess session.Session) (ginx.Result, error) {
	uid := sess.Claims().Uid

	if req.StartTime == 0 || req.EndTime == 0 {
		return invalidParamResult, errs.ErrInvalidParam
	}

	err := q.svc.CreateTempQuota(ctx, domain.TempQuota{
		Amount:    req.Amount,
		Key:       req.Key,
		Uid:       uid,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return systemErrorResult, err
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

func (q *QuotaHandler) Deduct(ctx *ginx.Context, req QuotaRequest, sees session.Session) (ginx.Result, error) {
	uid := sees.Claims().Uid
	err := q.svc.Deduct(ctx, uid, req.Amount, req.Key)
	if err != nil {
		// 检查是否是余额不足错误
		if errors.Is(err, errs.ErrInsufficientBalance) {
			return insufficientBalanceResult, nil
		}
		// 其他系统错误
		return systemErrorResult, nil
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (q *QuotaHandler) toQuotaResponse(tempQuotaList []domain.TempQuota) []QuotaResponse {
	return slice.Map[domain.TempQuota, QuotaResponse](tempQuotaList, func(idx int, src domain.TempQuota) QuotaResponse {
		return QuotaResponse{
			Amount:    src.Amount,
			StartTime: src.StartTime,
			EndTime:   src.EndTime,
		}
	})
}

type QuotaRequest struct {
	Amount    int64  `json:"amount,omitempty"`
	Key       string `json:"key,omitempty"`
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
}

type QuotaResponse struct {
	Amount    int64  `json:"amount,omitempty"`
	Key       string `json:"key"`
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
}
