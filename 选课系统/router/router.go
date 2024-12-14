package router

import (
	"finaltenzor/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.Use(middleware.Error)
	r.Use(middleware.GinLogger(), middleware.GinRecovery(true))
	apiRouter := r.Group("/api")
	{
		adminRouter := apiRouter.Group("/admin")
		{
			adminRouter.Use(middleware.CheckRole(2))
			{
				adminRouter.POST("/courses", ctr.Admin.AddCourse)                   // 添加课程
				adminRouter.DELETE("/courses/:courseId", ctr.Admin.DeleteCourse)    // 根据课程编号删除一门课程
				adminRouter.PUT("/courses", ctr.Admin.UpdateCourse)                 // 更新课程信息
				adminRouter.GET("/courses", ctr.Admin.GetCourses)                   // 获取所有的课程列表
				adminRouter.GET("/courses/:courseId", ctr.Admin.GetCourseDetail)    // 获取一门课的详情
				adminRouter.GET("/students", ctr.Admin.GetStudentsList)             // 获取学生列表
				adminRouter.GET("/students/:studentId", ctr.Admin.GetStudentDetail) // 获取某个学生具体信息
			}
		}
		userRouter := apiRouter.Group("/user")
		{
			userRouter.POST("/register", ctr.User.Register) // 学生注册
			userRouter.POST("", ctr.User.Login)             // 登录
			userRouter.Use(middleware.CheckLogin())
			{
				userRouter.DELETE("", ctr.User.Logout)     // 退出登录
				userRouter.GET("", ctr.User.GetUserStatus) // 获取当前用户状态
			}
			userRouter.Use(middleware.CheckRole(1))
			{
				userRouter.POST("/courses", ctr.User.GrabCourse)                 // 抢课
				userRouter.DELETE("/courses/:courseId", ctr.User.GiveUpCourse)   // 放弃选择这门课
				userRouter.GET("/courses-selected", ctr.User.ViewGrabbedCourses) // 查看自己已经抢到的课
				userRouter.GET("/schedule", ctr.User.GetSchedule)                // 获取用户当前已选课形成的课表
			}
			userRouter.GET("/courses", ctr.User.GetCoursesList)            // 获取课程列表
			userRouter.GET("/courses/:courseId", ctr.User.GetCourseDetail) // 根据课程编号获取某个课程详情
		}
	}
}
