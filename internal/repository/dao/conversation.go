package dao

import (
	"context"
	"time"

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
	err := dao.db.WithContext(ctx).Create(&c).Error
	if err != nil {
		return Conversation{}, err
	}

	return Conversation{}, nil
}

func (dao *ConversationDao) GetMessages(ctx context.Context, id int64) ([]Message, error) {
	var msgs []Message
	err := dao.db.WithContext(ctx).Where("cid = ?", id).Find(&msgs).Error
	if err != nil {
		return []Message{}, err
	}
	return msgs, nil
}

func (dao *ConversationDao) CreateMsgs(ctx context.Context, msgs []Message) error {
	now := time.Now().Unix()
	for _, msg := range msgs {
		msg.Ctime = now
		msg.Utime = now
	}

	return dao.db.WithContext(ctx).Create(&msgs).Error
}

type Conversation struct {
	ID    int64  `gorm:"primary_key;column:id"`
	Uid   int64  `gorm:"column:uid;index"`
	Title string `gorm:"column:title"`
	Ctime int64  `gorm:"column:ctime"`
	Utime int64  `gorm:"column:utime"`
}

type Message struct {
	ID            int64  `gorm:"primary_key;column:id"`
	CID           int64  `gorm:"column:cid;index"`
	Content       string `gorm:"column:content"`
	ReasonContent string `gorm:"column:reason_content"`
	Role          int64  `gorm:"column:role"`
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
