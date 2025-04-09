package grpc

import (
	"context"
	"github.com/ecodeclub/ai-gateway-go/api/gen/prompt/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
)

type PromptServer struct {
	svc *service.PromptService
	promptv1.UnimplementedPromptServiceServer
}

func NewPromptServer(svc *service.PromptService) *PromptServer {
	return &PromptServer{svc: svc}
}

func (p *PromptServer) Add(ctx context.Context, req *promptv1.AddRequest) (*promptv1.AddResponse, error) {
	err := p.svc.Add(ctx, req.Biz, req.Pattern, req.Name, req.Description)
	return &promptv1.AddResponse{Res: err == nil}, err
}

func (p *PromptServer) Get(ctx context.Context, req *promptv1.GetRequest) (*promptv1.GetResponse, error) {
	res, err := p.svc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &promptv1.GetResponse{
		Name: res.Name, Biz: res.Biz,
		Pattern:     res.Pattern,
		Description: res.Description,
		CreateTime:  res.Ctime.UnixMilli(),
		UpdateTime:  res.Utime.UnixMilli(),
	}, nil
}

func (p *PromptServer) Delete(ctx context.Context, req *promptv1.DeleteRequest) (*promptv1.DeleteResponse, error) {
	err := p.svc.Delete(ctx, req.Id)
	return &promptv1.DeleteResponse{Res: err == nil}, err
}

func (p *PromptServer) Update(ctx context.Context, req *promptv1.UpdateRequest) (*promptv1.UpdateResponse, error) {
	err := p.svc.Update(ctx, req.Id, req.Name, req.Pattern, req.Description)
	return &promptv1.UpdateResponse{Res: err == nil}, err
}
