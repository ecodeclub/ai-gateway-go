// Copyright 2025 ecodeclub
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

package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	"github.com/gotomicro/ego/core/elog"
	"golang.org/x/sync/errgroup"
)

type ProviderService struct {
	repo      *repository.ProviderRepository
	secretKey string
	logger    *elog.Component
}

func NewProviderService(repo *repository.ProviderRepository, secretKey string) *ProviderService {
	return &ProviderService{
		repo:      repo,
		secretKey: secretKey,
		logger:    elog.DefaultLogger.With(elog.FieldComponent("ProviderService")),
	}
}

func (p *ProviderService) SaveProvider(ctx context.Context, provider domain.Provider) (int64, error) {
	var err error
	provider.APIKey, err = p.Encrypt(provider.APIKey)
	if err != nil {
		return 0, err
	}
	return p.repo.SaveProvider(ctx, provider)
}

func (p *ProviderService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(p.secretKey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (p *ProviderService) ListProviders(ctx context.Context, offset, limit int) ([]domain.Provider, int64, error) {
	var (
		eg    errgroup.Group
		res   []domain.Provider
		total int64
	)
	eg.Go(func() error {
		var err error
		res, err = p.repo.ListProviders(ctx, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = p.repo.CountProviders(ctx)
		return err
	})
	err := eg.Wait()
	p.decryptProviders(res)
	return res, total, err
}

func (p *ProviderService) decryptProviders(providers []domain.Provider) {
	for i := range providers {
		decrypt, err1 := p.Decrypt(providers[i].APIKey)
		if err1 != nil {
			p.logger.Error("decrypt 失败", elog.FieldErr(err1), elog.Int64("providerId", providers[i].ID))
			decrypt = ""
		}
		providers[i].APIKey = decrypt
		d := providers[i]
		d.Models = nil
		for j := range providers[i].Models {
			providers[i].Models[j].Provider = d
		}
	}
}

func (p *ProviderService) Decrypt(ciphertext string) (string, error) {
	block, err := aes.NewCipher([]byte(p.secretKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("malformed ciphertext")
	}

	nonce, encrypted := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (p *ProviderService) ProviderDetail(ctx context.Context, id int64) (domain.Provider, error) {
	provider, err := p.repo.GetProvider(ctx, id)
	if err != nil {
		return domain.Provider{}, err
	}
	providers := []domain.Provider{provider}
	p.decryptProviders(providers)
	return providers[0], nil
}

func (p *ProviderService) SaveModel(ctx context.Context, model domain.Model) (int64, error) {
	return p.repo.SaveModel(ctx, model)
}

func (p *ProviderService) ModelDetail(ctx context.Context, id int64) (domain.Model, error) {
	m, err := p.repo.GetModel(ctx, id)
	if err != nil {
		return domain.Model{}, err
	}
	providers := []domain.Provider{m.Provider}
	p.decryptProviders(providers)
	m.Provider = providers[0]
	return m, nil
}
