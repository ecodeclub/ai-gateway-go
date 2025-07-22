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

package dao

import (
	"context"
	"gorm.io/gorm/clause"
	"time"

	"gorm.io/gorm"
)

type ChatDAO struct {
	db *gorm.DB
}

func NewChatDAO(db *gorm.DB) *ChatDAO {
	return &ChatDAO{db: db}
}

func (dao *ChatDAO) Save(ctx context.Context, c Chat) error {
	c.Utime = time.Now().Unix()
	c.Ctime = time.Now().Unix()
	return dao.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"title", "utime"}),
		}).Create(&c).Error
}

func (dao *ChatDAO) GetByUid(ctx context.Context, uid int64, limit int64, offset int64) ([]Chat, error) {
	var chats []Chat
	err := dao.db.WithContext(ctx).Model(&Chat{}).Where("uid = ?", uid).
		Order("id DESC").
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&chats).Error
	if err != nil {
		return chats, err
	}
	return chats, nil
}

func (dao *ChatDAO) GetBySN(ctx context.Context, sn string) (Chat, error) {
	var chat Chat
	err := dao.db.WithContext(ctx).Model(&Chat{}).Where("sn = ?", sn).First(&chat).Error
	return chat, err
}

func (dao *ChatDAO) GetMessages(ctx context.Context, sn string) ([]Message, error) {
	var messages []Message
	err := dao.db.WithContext(ctx).Where("sn = ?", sn).
		Order("id DESC").
		Find(&messages).Error
	if err != nil {
		return []Message{}, err
	}
	return messages, nil
}

func (dao *ChatDAO) AddMessages(ctx context.Context, messages []Message) error {
	now := time.Now().Unix()
	for _, msg := range messages {
		msg.Ctime = now
		msg.Utime = now
	}

	return dao.db.WithContext(ctx).Create(&messages).Error
}

type Chat struct {
	ID    int64  `gorm:"primary_key;autoIncrement"`
	Sn    string `gorm:"uniqueIndex;column:sn;size:36"`
	Uid   int64  `gorm:"column:uid;index"`
	Title string `gorm:"column:title"`
	Ctime int64  `gorm:"column:ctime"`
	Utime int64  `gorm:"column:utime"`
}

type Message struct {
	ID            int64  `gorm:"primary_key;column:id"`
	ChatSN        string `gorm:"column:chat_sn;type:varchar(128);index"`
	Content       string `gorm:"column:content"`
	ReasonContent string `gorm:"column:reason_content"`
	Role          string `gorm:"column:role;type:varchar(128);"`
	Ctime         int64  `gorm:"column:ctime"`
	Utime         int64  `gorm:"column:utime"`
}
