package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"user_id"`
	EngineID   string         `gorm:"type:varchar(36);uniqueIndex;not null"          json:"engine_id"`
	Sender     string         `gorm:"type:varchar(60);not null"                      json:"sender"`
	Recipient  string         `gorm:"type:varchar(60);not null"                      json:"recipient"`
	Amount     uint64         `gorm:"not null"                                       json:"amount"`
	Token      string         `gorm:"type:varchar(12);not null"                      json:"token"`
	Status     string         `gorm:"type:varchar(20);not null;default:'pending'"    json:"status"`
	TxHash     string         `gorm:"type:varchar(256)"                              json:"tx_hash,omitempty"`
	FeeStroops uint64         `                                                      json:"fee_stroops"`
	Urgency    string         `gorm:"type:varchar(20)"                               json:"urgency"`
	Error      string         `gorm:"type:text"                                      json:"error,omitempty"`
	CreatedAt  time.Time      `                                                      json:"created_at"`
	UpdatedAt  time.Time      `                                                      json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                                          json:"-"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
