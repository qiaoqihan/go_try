package service

import (
	"finaltenzor/model"
)

type StudentService struct {
}

func (s *StudentService) GetStudentsByCourse(courseID int64) ([]model.Student, error) {
	var studentIDs []int

	// 首先，查询中间表获得对应课程的学生ID
	err := model.DB.Table("course_student").Select("student_id").Where("course_id =?", courseID).Find(&studentIDs).Error
	if err != nil {
		return nil, err
	}

	// 然后，通过学生ID查询学生信息
	var students []model.Student
	err = model.DB.Table("student").Select("*").Where("student_id IN (?)", studentIDs).Find(&students).Error
	if err != nil {
		return nil, err
	}

	return students, nil
}
