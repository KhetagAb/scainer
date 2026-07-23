package gateway

import "errors"

var (
	ErrNoLogin         = errors.New("ejudge gateway: login missing in context")
	ErrTeacherNotFound = errors.New("ejudge gateway: teacher not found")
	ErrBootstrap       = errors.New("ejudge gateway: bootstrap failed")
)
