package model

type Admin struct {
	UserID  uint   `gorm:"type:INT UNSIGNED;NOT NULL;unique;comment:用户ID" json:"userid"`
	AdminID string `gorm:"type:VARCHAR(255);NOT NULL;unique;comment:管理员ID" json:"adminid"`
	Auth    uint   `gorm:"type:INT UNSIGNED;NOT NULL;comment:权限" json:"auth"`
	User    User   `gorm:"foreignkey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Admin) TableName() string {
	return "admin"
}
