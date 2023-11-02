package model

import (
	"time"
)

// TUsers 用户表
type TUsers struct {
	UserID    int64     `gorm:"primaryKey;column:user_id" json:"user_id"`
	UserName  string    `gorm:"unique;column:user_name" json:"user_name"` // 用户名
	Password  string    `gorm:"column:password" json:"password"`          // 密码哈希
	Position  float64   `gorm:"column:position" json:"position"`          // 位置
	Money     float64   `gorm:"column:money" json:"money"`                // 金额
	IsVip     int64     `gorm:"column:is_vip" json:"is_vip"`              // 已否为VIP, 1-是,0-否
	UUID      string    `gorm:"column:uuid" json:"uuid"`                  // UUID
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName get sql table name.获取数据库表名
func (m *TUsers) TableName() string {
	return "t_users"
}

// TUsersColumns get sql column name.获取数据库列名
var TUsersColumns = struct {
	UserID    string
	UserName  string
	Password  string
	Position  string
	Money     string
	IsVip     string
	UUID      string
	CreatedAt string
	UpdatedAt string
}{
	UserID:    "user_id",
	UserName:  "user_name",
	Password:  "password",
	Position:  "position",
	Money:     "money",
	IsVip:     "is_vip",
	UUID:      "uuid",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}
