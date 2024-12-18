package controller

import (
	"encoding/json"
	"errors"
	"finaltenzor/common"
	"finaltenzor/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type User struct{}

// Register - 学生注册
func (u *User) Register(c *gin.Context) {
	var form struct {
		StudentName string `json:"name" binding:"required"`
		Password    string `json:"password" binding:"required"`
		StudentID   string `json:"studentId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		logrus.Errorf("参数绑定错误: %v", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	student := model.User{
		UserID:   form.StudentID,
		UserName: form.StudentName,
		Password: form.Password,
	}
	err := srv.RegisterStudent(&student)
	if err != nil {
		logrus.Errorf("注册学生失败: %v", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"message": "注册成功"}))
}

// Login - 登录
func (u *User) Login(c *gin.Context) {
	if session := SessionGet(c, "user"); session != nil {
		c.Error(common.ErrNew(errors.New("请勿重复登陆"), common.AuthErr))
		return
	}
	var form struct {
		StudentID string `json:"studentId" binding:"required"`
		Password  string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		logrus.Errorf("参数绑定错误: %v", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	UserName, auth, err := srv.LoginStudent(form.StudentID, form.Password)
	if err != nil {
		logrus.Errorf("登录失败: %v", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	isUser := false
	if auth == 2 {
		userSession := UserSession{
			UserID:   form.StudentID,
			Username: UserName,
			Level:    2,
		}
		SessionSet(c, "user", userSession)
		isUser = true
	}
	if auth == 1 {
		userSession := UserSession{
			UserID:   form.StudentID,
			Username: UserName,
			Level:    1,
		}
		SessionSet(c, "user", userSession)
		isUser = true
	}
	if !isUser {
		c.Error(common.ErrNew(errors.New("您没有注册或登录权限"), common.AuthErr))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// Logout - 退出登录
func (u *User) Logout(c *gin.Context) {
	SessionClear(c)
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// GetUserStatus - 获取当前用户状态
func (u *User) GetUserStatus(c *gin.Context) {
	userSession := SessionGet(c, "user")
	studentID := userSession.(UserSession).UserID
	UserName, UserID, err := srv.GetCurrentUserStatus(studentID)
	if err != nil {
		logrus.Errorf("获取用户状态失败: %v", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"username": UserName, "userId": UserID}))
}

// GetCoursesList - 获取课程列表
func (u *User) GetCoursesList(c *gin.Context) {
	type QueryParams struct {
		Page       int      `form:"page" binding:"required,gt=0"`
		Limit      int      `form:"limit" binding:"required,gt=0"`
		CourseName string   `form:"courseName"`
		Teachers   []string `form:"teachers"`
		Location   string   `form:"location"`
		Time       string   `form:"time"`
	}
	var params QueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logrus.Errorf("参数错误: %v", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	var times []model.CourseTime
	type TimeForm struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	}
	if params.Time != "" {
		var timeForms []TimeForm
		if err := json.Unmarshal([]byte(params.Time), &timeForms); err != nil {
			logrus.Errorf("时间参数格式错误: %v", params.Time)
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
	courses, total, err := srv.GetCourses(params.Page, params.Limit, params.CourseName, params.Teachers, times, params.Location)
	if err != nil {
		c.Error(common.ErrNew(err, common.OpErr))
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
			c.Error(common.ErrNew(err, common.OpErr))
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
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"total": total, "courses": response}))
}

// GetCourseDetail - 获取课程详情
func (u *User) GetCourseDetail(c *gin.Context) {
	courseIdStr := c.Param("courseId")
	CourseID, err := strconv.ParseInt(courseIdStr, 10, 64)
	if err != nil {
		logrus.Errorf("无效的课程ID参数: %v", courseIdStr)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	course, err := srv.UserGetCourseDetail(CourseID)
	if err != nil {
		logrus.Errorf("获取课程详情失败: %v", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	type TimeForm struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	}
	type responseformat struct {
		CourseID       int64      `json:"id"`
		CourseName     string     `json:"courseName"`
		Capacity       int        `json:"capacity"`
		CourseTeachers []string   `json:"teachers"`
		Time           []TimeForm `json:"time"`
		Location       string     `json:"location"`
	}
	var response responseformat
	var timeForms []TimeForm
	for _, timeItem := range course.CourseTimes {
		timeForms = append(timeForms, TimeForm{
			StartTime: timeItem.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:   timeItem.EndTime.Format("2006-01-02 15:04:05"),
		})
	}
	TeacherNames, err := srv.GetTeacherNamesByCourses(course.CourseID)
	if err != nil {
		c.Error(common.ErrNew(err, common.OpErr))
		return
	}
	response = responseformat{
		CourseID:       course.CourseID,
		CourseName:     course.CourseName,
		Capacity:       course.Capacity,
		CourseTeachers: TeacherNames,
		Time:           timeForms,
		Location:       course.Location,
	}
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"course": response}))
}

// GrabCourse - 抢课
func (u *User) GrabCourse(c *gin.Context) {
	var form struct {
		CourseID int64 `form:"courseId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		logrus.Errorf("参数绑定错误: %v", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	userSession := SessionGet(c, "user")
	studentID := userSession.(UserSession).UserID
	err := srv.GrabCourse(studentID, form.CourseID)
	if err != nil {
		c.Error(common.ErrNew(err, common.OpErr))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// ViewGrabbedCourses - 查看自己已经抢到的课
func (u *User) ViewGrabbedCourses(c *gin.Context) {
	userSession := SessionGet(c, "user")
	studentID := userSession.(UserSession).UserID
	courses, total, err := srv.GetGrabbedCourses(studentID)
	if err != nil {
		c.Error(common.ErrNew(err, common.OpErr))
		return
	}
	type TimeForm struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
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
			c.Error(common.ErrNew(err, common.OpErr))
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
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"total": total, "courses": response}))
}

// GetSchedule - 获取用户当前已选课形成的课表
func (u *User) GetSchedule(c *gin.Context) {
	userSession := SessionGet(c, "user")
	studentID := userSession.(UserSession).UserID
	schedule, err := srv.GetUserSchedule(studentID)
	if err != nil {
		logrus.Errorf("获取课表失败: %v", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	type timeForm struct {
		StartTime time.Time `json:"startTime"`
		EndTime   time.Time `json:"endTime"`
	}
	type CourseTime struct {
		ID         *int64     `json:"id,omitempty"`
		CourseName *string    `json:"courseName,omitempty"`
		Teachers   []string   `json:"teachers"`
		Time       []timeForm `json:"time"`
		Location   *string    `json:"location,omitempty"`
	}
	coursesByDay := make(map[string][]CourseTime)
	daysOfWeek := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	for _, day := range daysOfWeek {
		coursesByDay[day] = []CourseTime{} // 初始化每一天为一个空的课程列表
	}
	for _, course := range schedule {
		teacherNames, err := srv.GetTeacherNamesByCourses(course.CourseID)
		if err != nil {
			c.Error(common.ErrNew(err, common.OpErr))
			return
		}
		for _, timeItem := range course.CourseTimes {
			courseTime := CourseTime{
				ID:         &course.CourseID,
				CourseName: &course.CourseName,
				Teachers:   teacherNames,
				Time: []timeForm{
					{
						StartTime: timeItem.StartTime,
						EndTime:   timeItem.EndTime,
					},
				},
				Location: &course.Location,
			}
			dayStr := timeItem.StartTime.Weekday().String()[:3]
			coursesByDay[dayStr] = append(coursesByDay[dayStr], courseTime)
		}
	}
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"courses": coursesByDay}))
}

// GiveUpCourse - 退课
func (u *User) GiveUpCourse(c *gin.Context) {
	userSession := SessionGet(c, "user")
	studentID := userSession.(UserSession).UserID
	CourseID := c.Param("courseId")
	err := srv.GiveUpCourse(studentID, CourseID)
	if err != nil {
		logrus.Errorf("退课失败: %v", err)
		c.Error(common.ErrNew(err, common.OpErr))
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}
