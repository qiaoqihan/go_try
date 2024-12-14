package config

import (
	"finaltenzor/service/validator"
)

func init() {
	initConfig()
	initLogger()
	validator.InitValidator(Config.AppLanguage)
}
