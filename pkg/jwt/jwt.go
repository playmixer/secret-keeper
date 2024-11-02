package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt"
)

type JWT struct {
	secret []byte
}

func New(secret []byte) *JWT {
	return &JWT{
		secret: secret,
	}
}

func (s *JWT) Create(params map[string]string) (string, error) {
	mapParams := jwt.MapClaims{}
	for k, v := range params {
		mapParams[k] = v
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapParams)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed signe token: %w", err)
	}

	return tokenString, nil
}

func (s *JWT) Verify(signedData string) (bool, error) {
	token, err := jwt.Parse(signedData, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unknown signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return false, fmt.Errorf("failed parse jwt token: %w", err)
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true, nil
	}

	return false, nil
}

func (s *JWT) GetParams(signedData string) (map[string]string, error) {
	res := map[string]string{}
	token, err := jwt.Parse(signedData, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unknown signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return res, fmt.Errorf("failed parse jwt token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		for k, v := range claims {
			res[k], ok = v.(string)
			if !ok {
				return res, errors.New("not as string")
			}
		}
		return res, nil
	}

	return res, nil
}
