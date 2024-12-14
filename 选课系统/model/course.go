package model

import (
	"time"

	"gorm.io/gorm"
)

type Course struct {
	CourseID   int64  `gorm:"primaryKey;UNSIGNED;NOT NULL;comment:课程ID" json:"courseID"`
	CourseName string `gorm:"type:VARCHAR(128) NOT NULL;comment:课程名称" json:"courseName"`
	Capacity   int    `gorm:"type:INT NOT NULL;comment:课程容量" json:"capacity"`
	Location   string `gorm:"type:VARCHAR(128) NOT NULL;comment:上课地点" json:"location"`

	CreatedAt time.Time      `gorm:"type:DATETIME(3);NOT NULL;comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"type:DATETIME(3);NOT NULL;comment:更新时间" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"type:DATETIME(3);NULL;index;comment:删除时间" json:"deletedAt"`

	CourseTimes []CourseTime `json:"courseTimes"`
}

func (Course) TableName() string {
	return "course"
}
