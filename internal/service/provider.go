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
)

type ProviderService struct {
	repo      *repository.ProviderRepo
	secretKey string
}

func NewProviderService(repo *repository.ProviderRepo, key string) *ProviderService {
	return &ProviderService{repo: repo, secretKey: key}
}

func (p *ProviderService) SaveProvider(ctx context.Context, provider domain.Provider) (int64, error) {
	provider.ApiKey, _ = p.Encrypt(p.secretKey)
	return p.repo.SaveProvider(ctx, provider)
}

func (p *ProviderService) SaveModel(ctx context.Context, model domain.Model) (int64, error) {
	return p.repo.SaveModel(ctx, model)
}

func (p *ProviderService) GetProviders(ctx context.Context) ([]domain.Provider, error) {
	return p.repo.GetProviders(ctx)
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
