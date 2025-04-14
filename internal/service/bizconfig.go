package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

var ErrBizConfigNotFound = repository.ErrBizConfigNotFound
var ErrQuotaExhausted = repository.ErrQuotaExhausted

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
		Quota:     req.Quota,
		UsedQuota: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

func (s *BizConfigService) GetByID(ctx context.Context, id string) (domain.BizConfig, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BizConfigService) Update(ctx context.Context, config domain.BizConfig) error {
	config.UpdatedAt = time.Now()
	return s.repo.Update(ctx, config)
}

func (s *BizConfigService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *BizConfigService) List(ctx context.Context, ownerID int64, ownerType string, page, pageSize int) ([]domain.BizConfig, int, error) {
	return s.repo.List(ctx, ownerID, ownerType, page, pageSize)
}

func (s *BizConfigService) CheckQuota(ctx context.Context, id string, requiredQuota int64) (bool, int64, error) {
	return s.repo.CheckAndUpdateQuota(ctx, id, requiredQuota)
}

func (s *BizConfigService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
}

func (s *BizConfigService) generateJWTToken(id string, ownerID int64, ownerType string) (string, error) {
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
