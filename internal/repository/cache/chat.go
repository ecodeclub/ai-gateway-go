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

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	NameSpace         = "conversation:%s"
	DefaultExpiration = 24 * time.Hour
)

type ChatCache struct {
	rdb redis.Cmdable
}

func NewChatCache(rdb redis.Cmdable) *ChatCache {
	return &ChatCache{rdb: rdb}
}

func (c *ChatCache) AddMessages(ctx context.Context, chatSN string, messages ...Message) error {
	if len(messages) == 0 {
		return nil
	}
	pipe := c.rdb.Pipeline()
	for _, msg := range messages {
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		pipe.RPush(ctx, c.key(chatSN), jsonMsg)
	}
	_, err := pipe.Exec(ctx)

	pipe.Expire(ctx, fmt.Sprintf(NameSpace, chatSN), DefaultExpiration)
	return err
}

// GetMessages 后续考虑分页
func (c *ChatCache) GetMessages(ctx context.Context, sn string) ([]Message, error) {
	messagesJSON, err := c.rdb.LRange(ctx, c.key(sn), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	messages := make([]Message, 0, len(messagesJSON))
	for _, jsonStr := range messagesJSON {
		var msg Message
		if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (c *ChatCache) key(sn string) string {
	return fmt.Sprintf(NameSpace, sn)
}

type Message struct {
	Role          string `json:"role"`
	Content       string `json:"content"`
	ReasonContent string `json:"reasonContent"`
}
