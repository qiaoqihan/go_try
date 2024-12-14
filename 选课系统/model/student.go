package model

type Student struct {
	UserID    uint   `gorm:"type:INT UNSIGNED NOT NULL;unique;comment:用户ID" json:"userId"`
	User      User   `gorm:"foreignkey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	StudentID string `gorm:"type:VARCHAR(128) NOT NULL;unique;comment:学号" json:"studentId"`
}

func (Student) TableName() string {
	return "student"
}
