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
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConversationDao struct {
	db *gorm.DB
}

func NewConversationDao(db *gorm.DB) *ConversationDao {
	return &ConversationDao{db: db}
}

func (dao *ConversationDao) Create(ctx context.Context, c Conversation) (Conversation, error) {
	c.Utime = time.Now().Unix()
	c.Ctime = time.Now().Unix()
	c.Sn = uuid.New().String()
	err := dao.db.WithContext(ctx).Create(&c).Error
	if err != nil {
		return Conversation{}, err
	}

	return c, nil
}

func (dao *ConversationDao) GetByUid(ctx context.Context, uid string, limit int64, offset int64) ([]Conversation, error) {
	var conversations []Conversation
	err := dao.db.WithContext(ctx).Model(&Conversation{}).Where("uid = ?", uid).
		Order("id DESC").
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&conversations).Error
	if err != nil {
		return conversations, err
	}
	return conversations, nil
}

func (dao *ConversationDao) GetById(ctx context.Context, id int64) (Conversation, error) {
	var conversation Conversation
	err := dao.db.WithContext(ctx).Model(&Conversation{}).Where("id = ?", id).First(&conversation).Error
	if err != nil {
		return conversation, err
	}
	return conversation, nil
}

func (dao *ConversationDao) GetMessages(ctx context.Context, sn string, limit int64, offset int64) ([]Message, error) {
	var messages []Message
	err := dao.db.WithContext(ctx).Where("sn = ?", sn).
		Order("id DESC").
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&messages).Error
	if err != nil {
		return []Message{}, err
	}
	return messages, nil
}

func (dao *ConversationDao) AddMessages(ctx context.Context, messages []Message) error {
	now := time.Now().Unix()
	for _, msg := range messages {
		msg.Ctime = now
		msg.Utime = now
	}

	return dao.db.WithContext(ctx).Create(&messages).Error
}

type Conversation struct {
	ID    int64  `gorm:"primary_key;autoIncrement"`
	Sn    string `gorm:"uniqueIndex;column:sn;size:36"`
	Uid   string `gorm:"column:uid;index"`
	Title string `gorm:"column:title"`
	Ctime int64  `gorm:"column:ctime"`
	Utime int64  `gorm:"column:utime"`
}

type Message struct {
	ID            int64  `gorm:"primary_key;column:id"`
	Sn            string `gorm:"column:sn;size:36;index"`
	Content       string `gorm:"column:content"`
	ReasonContent string `gorm:"column:reason_content"`
	Role          int32  `gorm:"column:role"`
	Ctime         int64  `gorm:"column:ctime"`
	Utime         int64  `gorm:"column:utime"`
}

func (Conversation) TableName() string {
	return "conversations"
}

func (Message) TableName() string {
	return "messages"
}

func InitConversation(db *gorm.DB) error {
	return db.AutoMigrate(&Conversation{}, &Message{})
}
