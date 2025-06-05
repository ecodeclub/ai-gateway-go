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
	NameSpace         = "conversation:%s:%s"
	DefaultExpiration = 24 * time.Hour
)

type ConversationCache struct {
	rdb redis.Cmdable
}

func NewConversationCache(rdb redis.Cmdable) *ConversationCache {
	return &ConversationCache{rdb: rdb}
}

func (c *ConversationCache) AddMessages(ctx context.Context, id string, uid string, messages []Message) error {
	pipe := c.rdb.Pipeline()

	for _, msg := range messages {
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		pipe.RPush(ctx, c.key(uid, id), jsonMsg)
	}
	_, err := pipe.Exec(ctx)

	pipe.Expire(ctx, fmt.Sprintf(NameSpace, id, uid), DefaultExpiration)
	return err
}

func (c *ConversationCache) GetMessage(ctx context.Context, id string, uid string, limit int64, offset int64) ([]Message, error) {
	length, err := c.rdb.LLen(ctx, c.key(uid, id)).Result()
	if err != nil {
		return []Message{}, err
	}

	start := offset
	end := offset + limit - 1
	if end >= length {
		end = length - 1
	}

	if start >= length {
		return []Message{}, nil
	}

	messagesJSON, err := c.rdb.LRange(ctx, id, start, end).Result()
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

func (c *ConversationCache) key(uid string, cid string) string {
	return fmt.Sprintf(NameSpace, uid, cid)
}

type Message struct {
	Role          int64  `json:"role"`
	Content       string `json:"content"`
	ReasonContent string `json:"reason_content"`
}
