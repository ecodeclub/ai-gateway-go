package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

var ErrBizConfigNotFound = repository.ErrBizConfigNotFound

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
	// 生成随机token
	token := generateRandomToken()

	// 创建配置
	config := domain.BizConfig{
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Token:     hashToken(token),
		Config:    req.Config,
	}

	created, err := s.repo.Create(ctx, config)
	if err != nil {
		return domain.BizConfig{}, "", err
	}

	// 生成JWT token
	jwtToken, err := s.generateJWTToken(created.ID, created.OwnerID, created.OwnerType)
	if err != nil {
		return domain.BizConfig{}, "", err
	}

	return created, jwtToken, nil
}

func (s *BizConfigService) GetByID(ctx context.Context, id int64) (domain.BizConfig, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BizConfigService) Update(ctx context.Context, config domain.BizConfig) error {
	config.UpdatedAt = time.Now().UnixMilli()
	return s.repo.Update(ctx, config)
}

func (s *BizConfigService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *BizConfigService) generateJWTToken(id int64, ownerID int64, ownerType string) (string, error) {
	claims := jwt.MapClaims{
		"biz_id":     id,
		"owner_id":   ownerID,
		"owner_type": ownerType,
		"exp":        time.Now().Add(s.tokenExpire).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func generateRandomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing token: %v", err)
		return ""
	}
	return string(hashed)
}
