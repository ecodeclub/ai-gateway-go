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

package admin

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/server/egin"
)

type PromptHandler struct {
	svc *service.PromptService
}

func NewPromptHandler(svc *service.PromptService) *PromptHandler {
	res := &PromptHandler{svc: svc}
	return res
}

func (h *PromptHandler) PrivateRoutes(server *egin.Component) {
	prompt := server.Group("/prompt")
	prompt.POST("/add", ginx.BS(h.Add))
	prompt.GET("/:id", ginx.W(h.Get))
	prompt.POST("/delete", ginx.B(h.Delete))
	prompt.POST("/delete/version", ginx.B(h.DeleteVersion))
	prompt.POST("/update", ginx.B(h.UpdatePrompt))
	prompt.POST("/update/version", ginx.B(h.UpdateVersion))
	prompt.POST("/publish", ginx.B(h.Publish))
	prompt.POST("/fork", ginx.B(h.Fork))
}

func (h *PromptHandler) PublicRoutes(_ *gin.Engine) {}

func (h *PromptHandler) Add(ctx *ginx.Context, req AddReq, sess session.Session) (ginx.Result, error) {
	uid := sess.Claims().Uid
	// 这里我假设 owner_type 也存储在 jwt token 里
	ownerType, err := sess.Claims().Get("owner_type").String()
	if err != nil {
		return ginx.Result{}, ginx.ErrUnauthorized
	}
	prompt := domain.Prompt{
		Name:        req.Name,
		Description: req.Description,
		Owner:       uid,
		OwnerType:   domain.OwnerType(ownerType),
	}
	version := domain.PromptVersion{
		Content:       req.Content,
		SystemContent: req.SystemContent,
		Temperature:   req.Temperature,
		TopN:          req.TopN,
		MaxTokens:     req.MaxTokens,
	}
	err = h.svc.Add(ctx, prompt, version)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *PromptHandler) Get(ctx *ginx.Context) (ginx.Result, error) {
	id, err := ctx.Param("id").AsInt64()
	if err != nil {
		return ginx.Result{}, ginx.ErrNoResponse
	}
	res, err := h.svc.Get(ctx, id)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Data: newPromptVO(res),
	}, nil
}

// Delete 删除整个 prompt
func (h *PromptHandler) Delete(ctx *ginx.Context, req DeleteReq) (ginx.Result, error) {
	err := h.svc.Delete(ctx, req.ID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *PromptHandler) DeleteVersion(ctx *ginx.Context, req DeleteVersionReq) (ginx.Result, error) {
	err := h.svc.DeleteVersion(ctx, req.VersionID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// UpdatePrompt 更新 prompt 的基本信息
func (h *PromptHandler) UpdatePrompt(ctx *ginx.Context, req UpdatePromptReq) (ginx.Result, error) {
	prompt := domain.Prompt{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
	}
	err := h.svc.UpdateInfo(ctx, prompt)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *PromptHandler) UpdateVersion(ctx *ginx.Context, req UpdateVersionReq) (ginx.Result, error) {
	version := domain.PromptVersion{
		ID:            req.VersionID,
		Content:       req.Content,
		SystemContent: req.SystemContent,
		Temperature:   req.Temperature,
		TopN:          req.TopN,
		MaxTokens:     req.MaxTokens,
	}
	err := h.svc.UpdateVersion(ctx, version)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *PromptHandler) Publish(ctx *ginx.Context, req PublishReq) (ginx.Result, error) {
	err := h.svc.Publish(ctx, req.VersionID, req.Label)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// Fork 新增一个版本
func (h *PromptHandler) Fork(ctx *ginx.Context, req ForkReq) (ginx.Result, error) {
	err := h.svc.Fork(ctx, req.VersionID)
	if err != nil {
		return systemErrorResult, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}
