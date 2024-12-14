package model

type CourseStudent struct {
	CourseID  int64  `gorm:"type:INT UNSIGNED NOT NULL;comment:课程ID" json:"courseId"`
	StudentID string `gorm:"type:VARCHAR(20) NOT NULL;comment:学生ID" json:"studentId"`
}

func (CourseStudent) TableName() string {
	return "course_student"
}
