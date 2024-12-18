package middleware

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"finaltenzor/common"
	"finaltenzor/controller"

	vl "finaltenzor/service/validator"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func Error(c *gin.Context) {
	c.Next()
	if len(c.Errors) != 0 {
		err := c.Errors.Last().Err
		switch err := err.(type) {
		case validator.ValidationErrors:
			var errs string
			for _, v := range err.Translate(vl.Trans) {
				errs = fmt.Sprintf("%v,%v", errs, v)
			}
			errorHandle(c, strings.Replace(errs, ",", "", 1))
		case *strconv.NumError, *json.UnmarshalTypeError, *time.ParseError, *xml.SyntaxError:
			errorHandle(c, errors.New("错误或非法的传入参数"))
		default:
			errorHandle(c, err)
		}
	}
}

func errorHandle(c *gin.Context, err any) {
	errMsg := fmt.Sprintf("%v: %v\n", common.ErrorMapper[uint64(c.Errors.Last().Type)], err)

	var statusCode int
	switch c.Errors.Last().Type {
	case common.ParamErr:
		statusCode = http.StatusBadRequest // 400
	case common.SysErr:
		statusCode = http.StatusInternalServerError // 500
	case common.OpErr:
		statusCode = http.StatusForbidden // 403
	case common.AuthErr:
		statusCode = http.StatusUnauthorized // 401
	case common.LevelErr:
		statusCode = http.StatusNotAcceptable // 406
	default:
		statusCode = http.StatusInternalServerError // 500
	}
	c.JSON(statusCode, controller.Response{
		Success: false,
		Message: errMsg,
		Code:    uint64(c.Errors.Last().Type),
	})
}
