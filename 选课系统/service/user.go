package service

import (
	"errors"
	"finaltenzor/model"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct{}

// RegisterStudent 注册新学生
func (us *User) RegisterStudent(student *model.Student) error {
	existingStudent := model.Student{}
	if err := model.DB.Where("student_id = ?", student.StudentID).First(&existingStudent).Error; err == nil {
		return errors.New("学生ID已存在")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(student.User.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	student.User.Password = string(hashedPassword)
	student.User.Role = "student"
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Create(&student.User).Error; err != nil {
		tx.Rollback()
		return err
	}
	student.UserID = uint(student.User.ID)
	if err := tx.Create(student).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

// LoginStudent 用户登录
func (us *User) LoginStudent(studentID, password string) (string, string, error) {
	var admin model.Admin
	if err := model.DB.Preload("User").Where("admin_id = ?", studentID).First(&admin).Error; err == nil && admin.User.UserName != "" {
		if admin.User.Password != password {
			return admin.User.UserName, "admin", errors.New("密码错误")
		}
		return admin.User.UserName, "admin", nil
	}
	var student model.Student
	if err := model.DB.Preload("User").Where("student_id = ?", studentID).First(&student).Error; err == nil && student.User.UserName != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(student.User.Password), []byte(password)); err != nil {
			return student.User.UserName, "student", errors.New("密码错误")
		}
		return student.User.UserName, "student", nil
	}
	return " ", "", errors.New("用户不存在")
}

// GetCurrentUserStatus 获取当前用户状态
func (us *User) GetCurrentUserStatus(studentID int) (string, string, error) {
	var admin model.Admin
	if err := model.DB.Preload("User").Where("admin_id = ?", studentID).First(&admin).Error; err == nil && admin.User.UserName != "" {
		return admin.User.UserName, admin.AdminID, nil
	}
	var student model.Student
	if err := model.DB.Preload("User").Where("student_id = ?", studentID).First(&student).Error; err == nil && student.User.UserName != "" {
		return student.User.UserName, student.StudentID, nil
	}
	return " ", "", errors.New("用户不存在")
}

// GrabCourse 抢课
func (us *User) GrabCourse(studentID int, courseID int64) error {
	var course model.Course
	if err := model.DB.Where("course_id = ?", courseID).First(&course).Error; err != nil {
		return errors.New("课程未找到")
	}
	var student model.Student
	if err := model.DB.Where("student_id = ?", studentID).First(&student).Error; err != nil {
		return errors.New("学生未找到")
	}
	var courseStudent model.CourseStudent
	if err := model.DB.Where("student_id = ? AND course_id = ?", student.StudentID, courseID).First(&courseStudent).Error; err == nil {
		return errors.New("学生已经抢过该课程")
	}
	var schedule []model.Course
	err := model.DB.Preload("CourseTimes").
		Joins("JOIN course_student ON course_student.course_id = course.course_id").
		Where("course_student.student_id = ?", studentID).
		Find(&schedule).Error
	if err == nil {
		for _, existingCourse := range schedule {
			for _, courseTime := range existingCourse.CourseTimes {
				for _, newCourseTime := range course.CourseTimes {
					if courseTime.StartTime.Before(newCourseTime.EndTime) && newCourseTime.StartTime.Before(courseTime.EndTime) {
						return errors.New("学生课程时间冲突，无法选择该课程")
					}
				}
			}
		}
	}
	var count int64
	if err := model.DB.Model(&model.CourseStudent{}).Where("course_id = ?", courseID).Count(&count).Error; err != nil {
		return err
	}
	if count >= int64(course.Capacity) {
		return errors.New("课程容量已满，无法选择该课程")
	}
	courseStudent = model.CourseStudent{
		StudentID: student.StudentID,
		CourseID:  courseID,
	}
	if err := model.DB.Create(&courseStudent).Error; err != nil {
		return err
	}
	return nil
}

// GetGrabbedCourses 获取抢到的课程列表
func (us *User) GetGrabbedCourses(studentID int) ([]model.Course, int, error) {
	var courses []model.Course
	err := model.DB.Preload("CourseTimes").
		Joins("JOIN course_student ON course_student.course_id = course.course_id").
		Where("course_student.student_id = ?", studentID).
		Find(&courses).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []model.Course{}, 0, nil
		}
		return nil, 0, err
	}
	return courses, len(courses), nil
}

// GetUserSchedule 获取用户的课表
func (us *User) GetUserSchedule(studentID int) ([]model.Course, error) {
	var courses []model.Course
	err := model.DB.Preload("CourseTimes").
		Joins("JOIN course_student ON course_student.course_id = course.course_id").
		Where("course_student.student_id = ?", studentID).
		Find(&courses).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户未选课")
		}
		return nil, err
	}
	return courses, nil
}

// GetCoursesList 获取课程列表
func (us *User) GetCoursesList(page int, limit int, courseName string, teachers []string, location string, times []model.CourseTime) ([]model.Course, int, error) {
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

// UserGetCourseDetail 获取课程详情
func (us *User) UserGetCourseDetail(courseID int64) (*model.Course, error) {
	var course model.Course
	if err := model.DB.Preload("CourseTimes").Where("course_id = ?", courseID).First(&course).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

// GiveUpCourse 放弃课程
func (us *User) GiveUpCourse(studentID int, courseID string) error {
	var courseStudent model.CourseStudent
	if err := model.DB.Where("student_id = ? AND course_id = ?", studentID, courseID).First(&courseStudent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("学生未选这门课程")
		}
		return err
	}
	if err := model.DB.Where("student_id = ? AND course_id = ?", studentID, courseID).Delete(&courseStudent).Error; err != nil {
		return err
	}
	return nil
}
