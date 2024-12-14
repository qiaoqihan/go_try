package service

type Service struct {
	User
	Admin
	TeacherService
	StudentService
}

func New() *Service {
	service := &Service{}
	return service
}
