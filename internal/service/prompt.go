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

	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
)

type PromptService struct {
	repo *repository.PromptRepo
}

func NewPromptService(repo *repository.PromptRepo) *PromptService {
	return &PromptService{repo: repo}
}

func (s *PromptService) Add(ctx context.Context, prompt domain.Prompt, version domain.PromptVersion) error {
	return s.repo.Create(ctx, prompt, version)
}

func (s *PromptService) Get(ctx context.Context, id int64) (domain.Prompt, error) {
	return s.repo.Get(ctx, id)
}

func (s *PromptService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *PromptService) DeleteVersion(ctx context.Context, versionID int64) error {
	return s.repo.DeleteVersion(ctx, versionID)
}

func (s *PromptService) UpdateInfo(ctx context.Context, prompt domain.Prompt) error {
	return s.repo.UpdateInfo(ctx, prompt)
}

func (s *PromptService) UpdateVersion(ctx context.Context, version domain.PromptVersion) error {
	return s.repo.UpdateVersion(ctx, version)
}

func (s *PromptService) Publish(ctx context.Context, versionID int64, label string) error {
	return s.repo.UpdateActiveVersion(ctx, versionID, label)
}

func (s *PromptService) Fork(ctx context.Context, versionID int64) error {
	prompt, err := s.repo.GetByVersionID(ctx, versionID)
	if err != nil {
		return err
	}
	newVersion := domain.PromptVersion{
		Content:       prompt.Versions[0].Content,
		SystemContent: prompt.Versions[0].SystemContent,
		Temperature:   prompt.Versions[0].Temperature,
		TopN:          prompt.Versions[0].TopN,
		MaxTokens:     prompt.Versions[0].MaxTokens,
	}
	return s.repo.InsertVersion(ctx, prompt.ID, newVersion)
}
