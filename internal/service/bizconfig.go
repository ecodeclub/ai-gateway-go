package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type BizConfigService struct {
	repo        *repository.BizConfigRepository
	jwtSecret   string
	tokenExpire time.Duration
}

func NewBizConfigService(repo *repository.BizConfigRepository, jwtSecret string, tokenExpire time.Duration) *BizConfigService {
	return &BizConfigService{
		repo:        repo,
		jwtSecret:   jwtSecret,
		tokenExpire: tokenExpire,
	}
}

func (s *BizConfigService) Create(ctx context.Context, req domain.BizConfig) (domain.BizConfig, string, error) {
	// 生成短 token，用于客户端存储，服务端也存一份（可以作为撤销等用途）
	token := generateRandomToken()

	config := domain.BizConfig{
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Token:     token,
		Config:    req.Config,
	}

	created, err := s.repo.Create(ctx, config)
	if err != nil {
		return domain.BizConfig{}, "", err
	}

	// 生成 JWT，给客户端用于后续权限校验
	jwtToken, err := s.generateJwtToken(created.ID)
	if err != nil {
		return domain.BizConfig{}, "", err
	}

	return created, jwtToken, nil
}

func (s *BizConfigService) GetByID(ctx context.Context, id int64) (domain.BizConfig, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BizConfigService) Update(ctx context.Context, config domain.BizConfig) error {
	return s.repo.Update(ctx, config)
}

func (s *BizConfigService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

type BizClaims struct {
	BizID int64
	jwt.RegisteredClaims
}

func (s *BizConfigService) generateJwtToken(id int64) (string, error) {
	claims := BizClaims{
		BizID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpire)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func generateRandomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
