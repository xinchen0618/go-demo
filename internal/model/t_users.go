package model

import (
	"time"
)

// TUsers 用户表
type TUsers struct {
	UserID    int64     `gorm:"primaryKey;column:user_id;type:bigint;not null" json:"user_id"`
	UserName  string    `gorm:"column:user_name;type:varchar(50);not null;default:''" json:"user_name"` // 用户名
	Password  string    `gorm:"column:password;type:char(38);not null;default:''" json:"password"`      // 密码
	Position  float64   `gorm:"column:position;type:float;not null;default:0" json:"position"`          // 位置
	Money     float64   `gorm:"column:money;type:decimal(10,2);not null;default:0.00" json:"money"`     // 金额
	IsVip     int64     `gorm:"column:is_vip;type:tinyint(1);not null;default:0" json:"is_vip"`         // 是否VIP,1-是,0-否
	UUID      string    `gorm:"column:uuid;type:varchar(50);not null;default:''" json:"uuid"`           // UUID
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
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
