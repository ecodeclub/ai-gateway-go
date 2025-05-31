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

type ConversationCache struct {
	rdb    redis.Cmdable
	length int64
}

func NewConversationCache(rdb redis.Cmdable, length int64) *ConversationCache {
	return &ConversationCache{rdb: rdb, length: length}
}

func (c *ConversationCache) AddMessages(ctx context.Context, id string, messages []Message) error {
	pipe := c.rdb.Pipeline()

	for _, msg := range messages {
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		pipe.RPush(ctx, fmt.Sprintf(NameSpace, id), jsonMsg)
	}
	_, err := pipe.Exec(ctx)

	pipe.Expire(ctx, fmt.Sprintf(NameSpace, id), DefaultExpiration)
	return err
}

func (c *ConversationCache) GetMessage(ctx context.Context, id string) ([]Message, error) {
	length, err := c.rdb.LLen(ctx, fmt.Sprintf(NameSpace, id)).Result()
	if err != nil {
		return []Message{}, err
	}

	start := length
	if length > c.length {
		start = length - 20
	} else {
		start = 0
	}

	messagesJSON, err := c.rdb.LRange(ctx, id, start, -1).Result()
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

type Message struct {
	Role          int64  `json:"role"`
	Content       string `json:"content"`
	ReasonContent string `json:"reason_content"`
}
