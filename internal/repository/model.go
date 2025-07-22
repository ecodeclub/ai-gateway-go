// Copyright 2023 ecodeclub
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

package repository

import (
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"strconv"
)

type ModelRepository struct {
}

func NewModelRepository() *ModelRepository {
	return &ModelRepository{}
}

func (repo *ModelRepository) FindById(id int64) (domain.Model, error) {
	// TODO 找到并且组合 model, provider
	return domain.Model{
		ID:   id,
		Name: "model" + strconv.FormatInt(id, 10),
		Provider: domain.Provider{
			ID:   id,
			Name: "provider" + strconv.FormatInt(id, 10),
		},
	}, nil
}
