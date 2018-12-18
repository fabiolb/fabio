package names

import "net/url"

type Service struct {
	Service   string
	Host      string
	Path      string
	TargetURL *url.URL
}

func (s Service) String() string {
	return s.Service
}
