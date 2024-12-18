package service

import (
	"errors"
	"finaltenzor/model"
	"strings"

	"gorm.io/gorm"
)

type Admin struct{}

// 添加课程
func (a *Admin) AddCourse(CourseName string, Capacity int, CourseTeachers []string, Time []model.CourseTime, Location string) (int64, error) {
	TeacherService := TeacherService{}
	var course model.Course
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Where("course_name = ?", CourseName).First(&course).Error; err == nil {
		return 0, errors.New("该课程已存在且未被删除")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return 0, err
	}
	for _, newTime := range Time {
		var count int64
		if err := tx.Model(&model.Course{}).
			Where("location = ? AND course_name != ? AND EXISTS (SELECT 1 FROM course_time WHERE course_time.course_id = course.course_id AND start_time < ? AND end_time > ?)",
				Location, CourseName, newTime.EndTime, newTime.StartTime).
			Count(&count).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
		if count > 0 {
			tx.Rollback()
			return 0, errors.New("该地点在指定时间内已被占用")
		}
	}
	for _, teacherName := range CourseTeachers {
		teacherID, err := TeacherService.FindOrRegisterTeacher(teacherName)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		for _, newTime := range Time {
			var count int64
			if err := tx.Model(&model.Course{}).
				Joins("JOIN course_teacher ON course_teacher.course_id = course.course_id").
				Where("course_teacher.teacher_id = ? AND EXISTS (SELECT 1 FROM course_time WHERE course_time.course_id = course.course_id AND start_time < ? AND end_time > ?)",
					teacherID, newTime.EndTime, newTime.StartTime).
				Count(&count).Error; err != nil {
				tx.Rollback()
				return 0, err
			}
			if count > 0 {
				tx.Rollback()
				return 0, errors.New("教师在指定时间内已被安排其他课程")
			}
		}
	}
	course = model.Course{
		CourseName:  CourseName,
		Capacity:    Capacity,
		CourseTimes: Time,
		Location:    Location,
	}
	if err := tx.Create(&course).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	var courseTeachers []model.CourseTeacher
	for _, teacherName := range CourseTeachers {
		teacherID, err := TeacherService.FindOrRegisterTeacher(teacherName)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		courseTeachers = append(courseTeachers, model.CourseTeacher{
			CourseID:  course.CourseID,
			TeacherID: teacherID,
		})
	}
	if len(courseTeachers) > 0 {
		if err := tx.Create(&courseTeachers).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	return course.CourseID, nil
}

// 删除课程
func (a *Admin) DeleteCourse(courseID int64) error {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var course model.Course
	if err := tx.First(&course, courseID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return errors.New("课程不存在")
		}
		tx.Rollback()
		return err
	}
	if err := tx.Where("course_id = ?", courseID).Delete(&model.CourseTime{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("course_id = ?", courseID).Delete(&model.CourseStudent{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("course_id = ?", courseID).Delete(&model.CourseTeacher{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&model.Course{}, courseID).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

// 更新课程
func (a *Admin) UpdateCourse(courseID int64, CourseName string, Capacity int, CourseTeachers []string, Time []model.CourseTime, Location string) error {
	TeacherService := TeacherService{}
	var course model.Course
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.First(&course, courseID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return errors.New("课程不存在")
		}
		tx.Rollback()
		return err
	}
	course.CourseName = CourseName
	course.Capacity = Capacity
	course.Location = Location
	if len(Time) > 0 {
		for _, timeItem := range Time {
			var conflictingCourses []model.Course
			if err := tx.Model(&model.Course{}).
				Joins("JOIN course_time ON course_time.course_id = course.course_id").
				Where("location = ? AND course.course_id != ? AND ((course_time.end_time > ?) AND (course_time.start_time < ?))",
					Location, courseID, timeItem.StartTime, timeItem.EndTime).Find(&conflictingCourses).Error; err != nil {
				tx.Rollback()
				return err
			}
			if len(conflictingCourses) > 0 {
				tx.Rollback()
				return errors.New("教室冲突，请检查课程时间安排")
			}
		}
		for i := range Time {
			Time[i].CourseID = course.CourseID
		}
		if err := tx.Where("course_id = ?", course.CourseID).Delete(&model.CourseTime{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Create(&Time).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if len(CourseTeachers) > 0 {
		var teacherIDs []int64
		for _, teacherName := range CourseTeachers {
			teacherID, err := TeacherService.FindOrRegisterTeacher(teacherName)
			if err != nil {
				tx.Rollback()
				return err
			}
			teacherIDs = append(teacherIDs, teacherID)
		}
		if len(teacherIDs) > 0 && len(Time) > 0 {
			var conflicts []model.CourseTeacher
			if err := tx.Model(&model.CourseTeacher{}).
				Joins("JOIN course_time ON course_time.course_id = course_teacher.course_id").
				Where("teacher_id IN (?) AND course_teacher.course_id != ? AND ((course_time.start_time < ? AND course_time.end_time > ?) OR (course_time.start_time < ? AND course_time.end_time > ?))",
					teacherIDs, courseID, Time[0].EndTime, Time[0].StartTime, Time[0].StartTime, Time[0].EndTime).Find(&conflicts).Error; err != nil {
				tx.Rollback()
				return err
			}
			if len(conflicts) > 0 {
				tx.Rollback()
				return errors.New("教师冲突，请检查课程时间安排")
			}
		}
		for _, teacherID := range teacherIDs {
			courseTeacher := model.CourseTeacher{
				CourseID:  course.CourseID,
				TeacherID: teacherID,
			}
			if err := tx.Create(&courseTeacher).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	if err := tx.Save(&course).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

// 查询课程
func (a *Admin) GetCourses(page int, limit int, courseName string, teachers []string, times []model.CourseTime, location string) ([]model.Course, int, error) {
	TeacherService := TeacherService{}
	var courses []model.Course
	var total int64
	query := model.DB.Model(&model.Course{}).Preload("CourseTimes")
	if courseName != "" {
		query = query.Where("course_name LIKE ?", "%"+courseName+"%")
	}
	if len(teachers) > 0 {
		var teacherIDs []int64
		for _, teacherName := range teachers {
			teacher, err := TeacherService.FindTeacherByName(teacherName)
			if err != nil {
				return nil, 0, err
			}
			if teacher == nil {
				return nil, 0, errors.New("教师未找到: " + teacherName)
			}
			teacherIDs = append(teacherIDs, teacher.ID)
		}
		if len(teacherIDs) > 0 {
			query = query.Joins("JOIN course_teacher ON course_teacher.course_id = course.course_id").
				Where("course_teacher.teacher_id IN (?)", teacherIDs)
		}
	}
	if len(times) > 0 {
		var timeConditions []string
		var timeArgs []interface{}
		for _, time := range times {
			timeConditions = append(timeConditions, "course_id IN (SELECT course_id FROM course_time WHERE start_time = ? AND end_time = ?)")
			timeArgs = append(timeArgs, time.StartTime, time.EndTime)
		}
		query = query.Where(strings.Join(timeConditions, " OR "), timeArgs...)
	}
	if location != "" {
		query = query.Where("location LIKE ?", "%"+location+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(limit).Offset((page - 1) * limit).Find(&courses).Error; err != nil {
		return nil, 0, err
	}
	return courses, int(total), nil
}

// 获取课程详情
func (a *Admin) GetCourseDetail(courseID int64) (*model.Course, error) {
	var course model.Course
	err := model.DB.Preload("CourseTimes").Where("course_id = ?", courseID).First(&course).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// 获取学生列表
func (a *Admin) GetStudentsList(page int, limit int, studentName string, studentID string) ([]model.User, map[string]int, error) {
	var students []model.User
	courseCounts := make(map[string]int)
	query := model.DB.Model(&model.User{})
	if studentName != "" {
		query = query.Where("user_name LIKE ?", "%"+studentName+"%")
	}
	if studentID != "" {
		query = query.Where("user_id = ?", studentID)
	}
	if err := query.Limit(limit).Offset((page - 1) * limit).Find(&students).Error; err != nil {
		return nil, nil, err
	}
	for _, student := range students {
		var count int64
		err := model.DB.Table("course_student").Where("student_id = ?", student.UserID).Count(&count).Error
		if err != nil {
			return nil, nil, err
		}
		courseCounts[student.UserID] = int(count)
	}
	return students, courseCounts, nil
}

// 获取学生详情
func (a *Admin) GetStudentDetail(studentID string) (*model.User, *[]model.Course, error) {
	var student model.User
	var courses []model.Course
	if err := model.DB.Where("user_id = ?", studentID).First(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("学生不存在")
		}
		return nil, nil, err
	}
	if err := model.DB.Table("course").
		Select("course.*").
		Joins("JOIN course_student ON course_student.course_id = course.course_id").
		Where("course_student.student_id = ?", studentID).
		Preload("CourseTimes").
		Find(&courses).Error; err != nil {
		return nil, nil, err
	}
	return &student, &courses, nil
}
