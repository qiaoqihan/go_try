package controller

import (
	"encoding/json"
	"finaltenzor/common"
	"finaltenzor/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Admin struct{}

// AddCourse 添加课程
func (a *Admin) AddCourse(c *gin.Context) {
	type timeform struct {
		StartTime string `json:"startTime" binding:"required"`
		EndTime   string `json:"endTime" binding:"required"`
	}
	var form struct {
		CourseId       int64      `json:"courseId" binding:"required"`
		CourseName     string     `json:"courseName" binding:"required"`
		Capacity       int        `json:"capacity" binding:"required"`
		CourseTeachers []string   `json:"teachers" binding:"required"`
		Time           []timeform `json:"time" binding:"required"`
		Location       string     `json:"location" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		logrus.Errorf("参数错误: %v", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var srvtime []model.CourseTime
	for _, timeItem := range form.Time {
		startTime, err := time.Parse("2006-01-02 15:04:05", timeItem.StartTime)
		if err != nil {
			logrus.Errorf("开始时间格式错误")
			c.Error(common.ErrNew(err, common.ParamErr))
			return
		}
		endTime, err := time.Parse("2006-01-02 15:04:05", timeItem.EndTime)
		if err != nil {
			logrus.Errorf("结束时间格式错误")
			c.Error(common.ErrNew(err, common.ParamErr))
			return
		}
		srvtime = append(srvtime, model.CourseTime{
			CourseID:  form.CourseId,
			StartTime: startTime,
			EndTime:   endTime,
		})
	}
	courseID, err := srv.AddCourse(form.CourseId, form.CourseName, form.Capacity, form.CourseTeachers, srvtime, form.Location)
	if err != nil {
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"id": courseID}))
}

// DeleteCourse 删除课程
func (a *Admin) DeleteCourse(c *gin.Context) {
	courseIdStr := c.Param("courseId")
	courseId, err := strconv.ParseInt(courseIdStr, 10, 64)
	if err != nil {
		logrus.Errorf("无效的 courseId: %v", courseIdStr)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	err = srv.DeleteCourse(courseId)
	if err != nil {
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// UpdateCourse 更新课程信息处理函数
func (a *Admin) UpdateCourse(c *gin.Context) {
	type timeform struct {
		StartTime string `form:"startTime"`
		EndTime   string `form:"endTime"`
	}
	var form struct {
		CourseId       int64      `json:"courseId" binding:"required"`
		CourseName     string     `json:"courseName"`
		Capacity       int        `json:"capacity"`
		CourseTeachers []string   `json:"teachers"`
		Time           []timeform `json:"time"`
		Location       string     `json:"location"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		logrus.Errorf("参数错误: %v", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var srvtime []model.CourseTime
	for _, timeItem := range form.Time {
		startTime, err := time.Parse("2006-01-02 15:04:05", timeItem.StartTime)
		if err != nil {
			logrus.Errorf("开始时间格式错误")
			c.Error(common.ErrNew(err, common.ParamErr))
			return
		}
		endTime, err := time.Parse("2006-01-02 15:04:05", timeItem.EndTime)
		if err != nil {
			logrus.Errorf("结束时间格式错误")
			c.Error(common.ErrNew(err, common.ParamErr))
			return
		}
		srvtime = append(srvtime, model.CourseTime{
			CourseID:  form.CourseId,
			StartTime: startTime,
			EndTime:   endTime,
		})
	}
	err := srv.UpdateCourse(form.CourseId, form.CourseName, form.Capacity, form.CourseTeachers, srvtime, form.Location)
	if err != nil {
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// GetCourses 查询课程
func (a *Admin) GetCourses(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	courseName := c.Query("courseName")
	teachers := c.QueryArray("teachers")
	location := c.Query("location")
	timeStrings := c.Query("time")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		logrus.Errorf("无效的页面参数: %v", pageStr)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		logrus.Errorf("无效的限制参数: %v", limitStr)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var times []model.CourseTime
	type TimeForm struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	}
	if timeStrings != "" {
		var timeForms []TimeForm
		if err := json.Unmarshal([]byte(timeStrings), &timeForms); err != nil {
			logrus.Errorf("时间参数格式错误: %v", timeStrings)
			c.Error(common.ErrNew(err, common.ParamErr))
			return
		}
		for _, timeItem := range timeForms {
			startTime, err := time.Parse("2006-01-02 15:04:05", timeItem.StartTime)
			if err != nil {
				logrus.Errorf("开始时间格式错误: %v", timeItem.StartTime)
				c.Error(common.ErrNew(err, common.ParamErr))
				return
			}
			endTime, err := time.Parse("2006-01-02 15:04:05", timeItem.EndTime)
			if err != nil {
				logrus.Errorf("结束时间格式错误: %v", timeItem.EndTime)
				c.Error(common.ErrNew(err, common.ParamErr))
				return
			}
			times = append(times, model.CourseTime{
				StartTime: startTime,
				EndTime:   endTime,
			})
		}
	}
	courses, total, err := srv.GetCourses(page, limit, courseName, teachers, times, location)
	if err != nil {
		logrus.Errorf("查询课程失败: %v", err)
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	type responseformat struct {
		CourseID       int64      `json:"id"`
		CourseName     string     `json:"courseName"`
		Capacity       int        `json:"capacity"`
		CourseTeachers []string   `json:"teachers"`
		Time           []TimeForm `json:"time"`
		Location       string     `json:"location"`
	}
	var response []responseformat
	for _, course := range courses {
		var timeForms []TimeForm
		for _, timeItem := range course.CourseTimes {
			timeForms = append(timeForms, TimeForm{
				StartTime: timeItem.StartTime.Format("2006-01-02 15:04:05"),
				EndTime:   timeItem.EndTime.Format("2006-01-02 15:04:05"),
			})
		}
		TeacherNames, err := srv.GetTeacherNamesByCourses(course.CourseID)
		if err != nil {
			logrus.Errorf("获取课程教师失败: %v", course.CourseID)
			c.Error(common.ErrNew(err, common.SysErr))
			return
		}
		response = append(response, responseformat{
			CourseID:       course.CourseID,
			CourseName:     course.CourseName,
			Capacity:       course.Capacity,
			CourseTeachers: TeacherNames,
			Time:           timeForms,
			Location:       course.Location,
		})
	}
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"size": total, "rows": response}))
}

// GetCourse 获取课程
func (a *Admin) GetCourseDetail(c *gin.Context) {
	courseIdStr := c.Param("courseId")
	couresID, err := strconv.ParseInt(courseIdStr, 10, 64)
	if err != nil {
		logrus.Errorf("无效的 courseId: %v", courseIdStr)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		logrus.Errorf("无效的页面参数: %v", pageStr)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		logrus.Errorf("无效的限制参数: %v", limitStr)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	course, err := srv.GetCourseDetail(page, limit, couresID)
	if err != nil {
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	type TimeForm struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	}
	type StudentForm struct {
		Name      string `json:"name"`
		StudentID string `json:"studentId"`
	}
	type responseformat struct {
		CourseID       int64         `json:"id"`
		CourseName     string        `json:"courseName"`
		Capacity       int           `json:"capacity"`
		Time           []TimeForm    `json:"time"`
		Location       string        `json:"location"`
		CourseTeachers []string      `json:"teachers"`
		TotalStudents  int           `json:"totalStudents"`
		Students       []StudentForm `json:"students"`
	}
	var timeForms []TimeForm
	for _, timeItem := range course.CourseTimes {
		timeForms = append(timeForms, TimeForm{
			StartTime: timeItem.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:   timeItem.EndTime.Format("2006-01-02 15:04:05"),
		})
	}
	students, err := srv.GetStudentsByCourse(couresID)
	if err != nil {
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	var studentForms []StudentForm
	for _, student := range students {
		studentForms = append(studentForms, StudentForm{
			Name:      student.User.UserName,
			StudentID: student.StudentID,
		})
	}
	TeacherNames, err := srv.GetTeacherNamesByCourses(course.CourseID)
	if err != nil {
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	response := responseformat{
		CourseID:       course.CourseID,
		CourseName:     course.CourseName,
		Capacity:       course.Capacity,
		Time:           timeForms,
		Location:       course.Location,
		CourseTeachers: TeacherNames,
		TotalStudents:  len(students),
		Students:       studentForms,
	}
	c.JSON(http.StatusOK, ResponseNew(c, response))
}

// GetStudentsList 获取学生列表处理函数
func (a *Admin) GetStudentsList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	studentName := c.Query("studentName")
	studentId := c.Query("studentId")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		logrus.Errorf("无效的页面参数: %v", pageStr)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		logrus.Errorf("无效的限制参数: %v", limitStr)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	type StudentFormat struct {
		StudentName  string `json:"studentName"`
		StudentID    string `json:"studentId"`
		TotalCourses int    `json:"totalCourses"`
	}
	students, courseCounts, err := srv.GetStudentsList(page, limit, studentName, studentId)
	if err != nil {
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	var studentForms []StudentFormat
	for _, student := range students {
		studentForms = append(studentForms, StudentFormat{
			StudentName:  student.User.UserName,
			StudentID:    student.StudentID,
			TotalCourses: courseCounts[student.StudentID],
		})
	}
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"students": studentForms}))
}

// GetStudentDetail 获取学生具体信息处理函数
func (a *Admin) GetStudentDetail(c *gin.Context) {
	studentId := c.Param("studentId")
	student, courses, err := srv.GetStudentDetail(studentId)
	if err != nil {
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	type CourseTimeFormat struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	}
	type CourseFormat struct {
		CourseID   int64              `json:"id"`
		CourseName string             `json:"courseName"`
		Teachers   []string           `json:"teacher"`
		Time       []CourseTimeFormat `json:"time"`
		Location   string             `json:"location"`
	}
	type ResponseFormat struct {
		StudentName string         `json:"studentName"`
		Courses     []CourseFormat `json:"courses"`
	}
	courseForms := make([]CourseFormat, len(*courses)) // 预分配切片
	for i, course := range *courses {
		Teacher, err := srv.GetTeacherNamesByCourses(course.CourseID)
		if err != nil {
			c.Error(common.ErrNew(err, common.SysErr))
			return
		}
		var timeForms []CourseTimeFormat
		for _, timeItem := range course.CourseTimes {
			timeForms = append(timeForms, CourseTimeFormat{
				StartTime: timeItem.StartTime.Format("2006-01-02 15:04:05"),
				EndTime:   timeItem.EndTime.Format("2006-01-02 15:04:05"),
			})
		}
		courseForms[i] = CourseFormat{
			CourseID:   course.CourseID,
			CourseName: course.CourseName,
			Teachers:   Teacher,
			Time:       timeForms,
			Location:   course.Location,
		}
	}
	response := ResponseFormat{
		StudentName: student.User.UserName,
		Courses:     courseForms,
	}
	c.JSON(http.StatusOK, ResponseNew(c, response))
}
