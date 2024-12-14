package model

import (
	"time"
)

type CourseTime struct {
	CourseID  int64     `gorm:"type:INT UNSIGNED NOT NULL;comment:课程ID" json:"courseId"`
	StartTime time.Time `gorm:"type:DATETIME NOT NULL;comment:开始时间" json:"startTime"`
	EndTime   time.Time `gorm:"type:DATETIME NOT NULL;comment:结束时间" json:"endTime"`
}

func (CourseTime) TableName() string {
	return "course_time"
}
