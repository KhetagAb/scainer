package gateway

import "errors"

var (
	ErrNoLogin          = errors.New("ejudge gateway: login missing in context")
	ErrTeacherNotFound  = errors.New("ejudge gateway: teacher not found")
	ErrAPIKeyProvision  = errors.New("ejudge gateway: api key provision failed")
)
