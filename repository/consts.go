package repository

import "github.com/pkg/errors"


var (
	ErrorInvalidCredentials error = errors.New("invalid credentials")
	ErrorUserNotActivated   error = errors.New("user is not yet activated")
	ErrorInternalServer     error = errors.New("internal server error")
)

