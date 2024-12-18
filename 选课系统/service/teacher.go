package service

import (
	"finaltenzor/model"

	"gorm.io/gorm"
)

type TeacherService struct {
}

func (t *TeacherService) FindTeacherByName(name string) (*model.Teacher, error) {
	var teacher model.Teacher
	if err := model.DB.Where("name = ?", name).First(&teacher).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &teacher, nil
}

func (t *TeacherService) FindTeacherByID(id int64) (*model.Teacher, error) {
	var teacher model.Teacher
	if err := model.DB.Where("id = ?", id).First(&teacher).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &teacher, nil
}

func (t *TeacherService) RegisterTeacher(name string) (int64, error) {
	newTeacher := model.Teacher{Name: name}
	if err := model.DB.Create(&newTeacher).Error; err != nil {
		return 0, err
	}
	return newTeacher.ID, nil
}

func (t *TeacherService) FindOrRegisterTeacher(name string) (int64, error) {
	teacher, err := t.FindTeacherByName(name)
	if err != nil {
		return 0, err
	}
	if teacher != nil {
		return teacher.ID, nil
	}
	return t.RegisterTeacher(name)
}

func (t *TeacherService) GetTeacherNamesByCourses(CourseID int64) ([]string, error) {
	var teacherNames []string
	var teacherIDs []int64
	err := model.DB.Table("course_teacher").
		Where("course_id = ?", CourseID).
		Pluck("teacher_id", &teacherIDs).Error
	if err != nil {
		return nil, err
	}
	if len(teacherIDs) > 0 {
		err = model.DB.Table("teacher").
			Where("id IN ?", teacherIDs).
			Pluck("name", &teacherNames).Error
		if err != nil {
			return nil, err
		}
	}
	return teacherNames, nil
}
