package keeper

import (
	"golang.org/x/crypto/bcrypt"
)

func validateLogin(login string) error {
	if login == "" {
		return ErrLoginNotValid
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return ErrPasswordNotValid
	}
	return nil
}

func hashPassword(password string) (string, error) {
	cost := 14
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
