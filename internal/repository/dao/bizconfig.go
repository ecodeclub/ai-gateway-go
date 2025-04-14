package dao

import (
	"errors"
	"github.com/ecodeclub/ai-gateway-go/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var (
	ErrBizConfigNotFound = errors.New("biz config not found")
	ErrQuotaExhausted    = errors.New("quota exhausted")
)

type BizConfig struct {
	ID        string    `gorm:"column:id;type:varchar(36);primaryKey"`
	OwnerID   int64     `gorm:"column:owner_id;type:bigint;not null"`
	OwnerType string    `gorm:"column:owner_type;type:varchar(20);not null"`
	Token     string    `gorm:"column:token;type:varchar(64);not null"`
	Config    string    `gorm:"column:config;type:text"`
	Quota     int64     `gorm:"column:quota;type:bigint;not null"`
	UsedQuota int64     `gorm:"column:used_quota;type:bigint;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func (BizConfig) TableName() string {
	return "biz_configs"
}

type BizConfigDAO struct {
	db *gorm.DB
}

func NewBizConfigDAO(db *gorm.DB) *BizConfigDAO {
	return &BizConfigDAO{db: db}
}

// Create 创建记录，返回创建后的对象（含 ID 和 Token）
func (d *BizConfigDAO) Insert(ctx context.Context, bc *BizConfig) (*BizConfig, error) {
	if bc.ID == "" {
		bc.ID = uuid.New().String()
	}
	if bc.Token == "" {
		bc.Token = uuid.New().String()
	}
	err := d.db.WithContext(ctx).Create(bc).Error
	if err != nil {
		return nil, err
	}
	return bc, nil
}

// GetByID 根据ID查询
func (d *BizConfigDAO) GetByID(ctx context.Context, id string) (*BizConfig, error) {
	var bc BizConfig
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&bc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrBizConfigNotFound
	}
	return &bc, err
}

// Update 更新记录
func (d *BizConfigDAO) Update(ctx context.Context, bc *BizConfig) error {
	return d.db.WithContext(ctx).Save(bc).Error
}

// Delete 删除记录
func (d *BizConfigDAO) Delete(ctx context.Context, id string) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&BizConfig{}).Error
}

// List 列表分页查询
func (d *BizConfigDAO) List(ctx context.Context, ownerID int64, ownerType string, page, pageSize int) ([]domain.BizConfig, int, error) {
	var count int64
	var configs []BizConfig

	tx := d.db.WithContext(ctx).Model(&BizConfig{}).Where("owner_id = ? AND owner_type = ?", ownerID, ownerType)

	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Offset((page - 1) * pageSize).Limit(pageSize).Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	// 转换为 domain.BizConfig
	var res []domain.BizConfig
	for _, c := range configs {
		res = append(res, domain.BizConfig{
			ID:        c.ID,
			OwnerID:   c.OwnerID,
			OwnerType: c.OwnerType,
			Token:     c.Token,
			Config:    c.Config,
			Quota:     c.Quota,
			UsedQuota: c.UsedQuota,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}

	return res, int(count), nil
}

func (d *BizConfigDAO) CheckAndUpdateQuota(ctx context.Context, id string, requiredQuota int64) (bool, int64, error) {
	var quotaOK bool
	var newRemaining int64

	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var bc BizConfig
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", id).
			First(&bc).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrBizConfigNotFound
			}
			return err
		}

		remaining := bc.Quota - bc.UsedQuota
		if remaining < requiredQuota {
			return ErrQuotaExhausted
		}

		bc.UsedQuota += requiredQuota
		if err := tx.Save(&bc).Error; err != nil {
			return err
		}

		quotaOK = true
		newRemaining = bc.Quota - bc.UsedQuota
		return nil
	})

	return quotaOK, newRemaining, err
}

// GetRemainingQuota 查询剩余配额
func (d *BizConfigDAO) GetRemainingQuota(ctx context.Context, id string) (int64, error) {
	var bc BizConfig
	if err := d.db.WithContext(ctx).Select("quota", "used_quota").Where("id = ?", id).First(&bc).Error; err != nil {
		return 0, err
	}
	return bc.Quota - bc.UsedQuota, nil
}
