package middleware

import (
	"errors"

	"finaltenzor/common"
	"finaltenzor/controller"

	"github.com/gin-gonic/gin"
)

func CheckRole(auth int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userSession := controller.SessionGet(c, "user")
		if userSession == nil {
			c.Error(common.ErrNew(errors.New("您未登录"), common.AuthErr))
			c.Abort()
			return
		}
		if userSession.(controller.UserSession).Level != auth {
			c.Error(common.ErrNew(errors.New("您没有权限"), common.LevelErr))
			c.Abort()
			return
		}
		c.Next()
	}
}

func CheckLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userSession := controller.SessionGet(c, "user")
		if userSession == nil {
			c.Error(common.ErrNew(errors.New("您未登录"), common.AuthErr))
			c.Abort()
			return
		}
		c.Next()
	}
}
