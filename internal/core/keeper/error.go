package keeper

import "errors"

var (
	ErrPasswordNotValid = errors.New("password is not valid")
	ErrLoginNotValid    = errors.New("login is not valid")
)
