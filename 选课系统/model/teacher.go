package model

type Teacher struct {
	Name string `gorm:"type:VARCHAR(128) NOT NULL;comment:教师姓名" json:"name"`

	BaseModel
}

func (Teacher) TableName() string {
	return "teacher"
}
