package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primary_key;AUTO_INCREMENT" comment:"用户ID" json:"id"`
	UserName  string         `gorm:"type:VARCHAR(128) NOT NULL;comment:用户名" json:"studentName"`
	Password  string         `gorm:"type:VARCHAR(128) NOT NULL;comment:密码" json:"-"`
	Role      string         `gorm:"type:VARCHAR(128) NOT NULL;comment:角色" json:"role"`
	CreatedAt time.Time      `gorm:"type:DATETIME(3);NOT NULL;comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"type:DATETIME(3);NOT NULL;comment:更新时间" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"type:DATETIME(3);NULL;index;comment:删除时间" json:"deletedAt"`
}

func (User) TableName() string {
	return "user"
}
