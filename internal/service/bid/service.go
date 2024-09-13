package bid

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) List() {
	return
}
