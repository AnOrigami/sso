package model

import (
	"time"

	"gorm.io/gorm"
)

// 对应表模型
type Model struct {
	ID        uint           `gorm:"primaryKey;" json:"id"`
	CreatedAt time.Time      `gorm:"not null;" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;" json:"deleted_at"`
}

type Application struct {
	Model
	AppKey   string `gorm:"not null;" json:"app_key"`
	Name     string `gorm:"not null;" json:"name"`
	Site     string `gorm:"not null;unique;" json:"site"`
	Redirect string `gorm:"not null;unique;" json:"redirect"`
}

type Role struct {
	Model
	Name        string `gorm:"column:name;not null;unique;" json:"name"`
	Description string `gorm:"column:description;not null;" json:"description"`
}

type User struct {
	Model
	Username     string `gorm:"not null;unique;" json:"username"`
	PasswordHash string `gorm:"not null;" json:"password_hash"`
}

type UserRole struct {
	UserID uint `gorm:"not null;index:idx_user_role,unique;" json:"user_id"`
	RoleID uint `gorm:"not null;index:idx_user_role,unique;" json:"role_id"`
}
