package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PoolSnapshot struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ContractID     string         `gorm:"type:varchar(60);not null;index"               json:"contract_id"`
	Reserve0       string         `gorm:"type:varchar(40);not null"                     json:"reserve_0"`
	Reserve1       string         `gorm:"type:varchar(40);not null"                     json:"reserve_1"`
	Token0         string         `gorm:"type:varchar(12);not null"                     json:"token_0"`
	Token1         string         `gorm:"type:varchar(12);not null"                     json:"token_1"`
	LpTotalSupply  string         `gorm:"type:varchar(40);not null"                     json:"lp_total_supply"`
	TimestampLast  int64          `gorm:"not null"                                      json:"timestamp_last"`
	CreatedAt      time.Time      `                                                     json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index"                                         json:"-"`
}

func (p *PoolSnapshot) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
