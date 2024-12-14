package model

type CourseTeacher struct {
	CourseID  int64 `gorm:"type:INT UNSIGNED NOT NULL;comment:课程ID" json:"courseId"`
	TeacherID int64 `gorm:"type:INT UNSIGNED NOT NULL;comment:教师ID" json:"teacherId"`
}

func (CourseTeacher) TableName() string {
	return "course_teacher"
}
