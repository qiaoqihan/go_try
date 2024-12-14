package controller

type Controller struct {
	User
	Admin
}

func New() *Controller {
	Controller := &Controller{}
	return Controller
}
