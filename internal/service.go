package internal

type Service struct {
	s *Storage
}

func NewService(s *Storage) *Service {
	return &Service{s: s}
}

func (s *Service) Ping() string {
	return "pong"
}
