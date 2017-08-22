package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Token struct {
	gorm.Model
	Token     string    `gorm:"unique" json:"token"`
	ExpiredTo time.Time `json:"expired_to"`

	User   User `gorm:"ForeignKey:TGID;AssociationForeignKey:UserID"`
	UserID uint `json:"user_id"`
}

func (t *Token) Store(db *gorm.DB) error {
	if db.Create(t).Error != nil {
		return db.Error
	}
	return db.Error
}

func (t *Token) GetUserByToken(db *gorm.DB) *User {
	if !db.Model(&Token{}).Preload("User").
		Where("token = ? AND expired_to > ?", t.Token, time.Now()).
		First(t).RecordNotFound() {
		return &t.User
	}
	return nil
}