package service

import (
	"finaltenzor/model"
)

type StudentService struct {
}

func (s *StudentService) GetStudentsByCourse(page int, limit int, courseID int64) ([]model.User, error) {
	var studentIDs []int
	err := model.DB.Table("course_student").Select("student_id").Where("course_id =?", courseID).Find(&studentIDs).Error
	if err != nil {
		return nil, err
	}
	var students []model.User
	err = model.DB.Where("user_id IN (?)", studentIDs).Limit(limit).Offset((page - 1) * limit).Find(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}
