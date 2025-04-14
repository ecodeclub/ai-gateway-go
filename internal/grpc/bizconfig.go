package grpc

import (
	"context"
	"errors"
	ai "github.com/ecodeclub/ai-gateway-go/api/gen/ai/v1"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BizConfigServer struct {
	svc *service.BizConfigService
	ai.UnimplementedBizConfigServiceServer
}

func NewBizConfigServer(svc *service.BizConfigService) *BizConfigServer {
	return &BizConfigServer{svc: svc}
}

func (s *BizConfigServer) CreateBizConfig(ctx context.Context, req *ai.CreateBizConfigRequest) (*ai.BizConfigResponse, error) {
	config := domain.BizConfig{
		OwnerID:   req.GetOwnerId(),
		OwnerType: req.GetOwnerType(),
		Config:    req.GetConfig(),
		Quota:     req.GetQuota(),
	}

	created, token, err := s.svc.Create(ctx, config)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create biz config: %v", err)
	}

	return &ai.BizConfigResponse{
		Config: s.toProtoBizConfig(created),
		Token:  token,
	}, nil
}

func (s *BizConfigServer) GetBizConfig(ctx context.Context, req *ai.GetBizConfigRequest) (*ai.BizConfigResponse, error) {
	config, err := s.svc.GetByID(ctx, req.GetId())
	if err == service.ErrBizConfigNotFound {
		return nil, status.Errorf(codes.NotFound, "biz config not found")
	}

	return &ai.BizConfigResponse{
		Config: s.toProtoBizConfig(config),
	}, nil
}

func (s *BizConfigServer) UpdateBizConfig(ctx context.Context, req *ai.UpdateBizConfigRequest) (*ai.BizConfigResponse, error) {
	existing, err := s.svc.GetByID(ctx, req.GetId())
	if err == service.ErrBizConfigNotFound {
		return nil, status.Errorf(codes.NotFound, "biz config not found")
	}

	existing.Config = req.GetConfig()
	existing.Quota = req.GetQuota()

	if err := s.svc.Update(ctx, existing); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update biz config: %v", err)
	}

	updated, err := s.svc.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get updated biz config: %v", err)
	}

	return &ai.BizConfigResponse{
		Config: s.toProtoBizConfig(updated),
	}, nil
}

func (s *BizConfigServer) DeleteBizConfig(ctx context.Context, req *ai.DeleteBizConfigRequest) (*ai.DeleteBizConfigResponse, error) {
	if err := s.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete biz config: %v", err)
	}

	return &ai.DeleteBizConfigResponse{Success: true}, nil
}

func (s *BizConfigServer) ListBizConfigs(ctx context.Context, req *ai.ListBizConfigsRequest) (*ai.ListBizConfigsResponse, error) {
	configs, total, err := s.svc.List(ctx, req.GetOwnerId(), req.GetOwnerType(), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list biz configs: %v", err)
	}

	var pbConfigs []*ai.BizConfig
	for _, config := range configs {
		pbConfigs = append(pbConfigs, s.toProtoBizConfig(config))
	}

	return &ai.ListBizConfigsResponse{
		Configs: pbConfigs,
		Total:   int32(total),
	}, nil
}

func (s *BizConfigServer) CheckQuota(ctx context.Context, req *ai.CheckQuotaRequest) (*ai.CheckQuotaResponse, error) {
	allowed, remaining, err := s.svc.CheckQuota(ctx, req.GetId(), req.GetRequiredQuota())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBizConfigNotFound):
			return nil, status.Errorf(codes.NotFound, "biz config not found")
		case errors.Is(err, service.ErrQuotaExhausted):
			return nil, status.Errorf(codes.ResourceExhausted, "quota exhausted")
		default:
			return nil, status.Errorf(codes.Internal, "failed to check quota: %v", err)
		}
	}

	return &ai.CheckQuotaResponse{
		Allowed:        allowed,
		RemainingQuota: remaining,
	}, nil
}

func (s *BizConfigServer) UpdateQuota(ctx context.Context, req *ai.UpdateQuotaRequest) (*ai.UpdateQuotaResponse, error) {
	// 这里简化处理，实际应该调用服务层方法更新配额
	config, err := s.svc.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "biz config not found: %v", err)
	}

	config.UsedQuota += req.GetUsedQuota()
	if err := s.svc.Update(ctx, config); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update quota: %v", err)
	}

	remaining := config.Quota - config.UsedQuota
	return &ai.UpdateQuotaResponse{
		Success:        true,
		RemainingQuota: remaining,
	}, nil
}

func (s *BizConfigServer) toProtoBizConfig(config domain.BizConfig) *ai.BizConfig {
	return &ai.BizConfig{
		Id:        config.ID,
		OwnerId:   config.OwnerID,
		OwnerType: config.OwnerType,
		Token:     config.Token,
		Config:    config.Config,
		Quota:     config.Quota,
		UsedQuota: config.UsedQuota,
		CreatedAt: timestamppb.New(config.CreatedAt),
		UpdatedAt: timestamppb.New(config.UpdatedAt),
	}
}
