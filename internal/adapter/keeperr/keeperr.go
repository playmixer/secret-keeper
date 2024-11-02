package keeperr

import "errors"

var (
	ErrNotFound = errors.New("not found")

	ErrLoginNotUnique            = errors.New("login not unique")
	ErrLoginOrPasswordNotCorrect = errors.New("login or password not correct")
)
