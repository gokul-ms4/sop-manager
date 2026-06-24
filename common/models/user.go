package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                  int            `json:"id" gorm:"primaryKey"`
	Name                string         `json:"name"`
	Email               string         `json:"email"`
	PhoneNumber         string         `json:"phone_number"`
	Password            string         `json:"password"`
	Avatar              string         `json:"avatar"`
	PasswordResetToken  string         `json:"-"`
	PasswordResetExpiry *time.Time     `json:"-"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}
