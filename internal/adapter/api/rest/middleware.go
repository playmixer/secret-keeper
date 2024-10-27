package rest

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/playmixer/secret-keeper/pkg/jwt"
)

func (s *Server) middlewareAuthorization(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	jwtManager := jwt.New(s.secretKey)

	token := strings.Replace(authHeader, "Bearer ", "", 1)
	if ok, err := jwtManager.Verify(token); err != nil || !ok {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		c.Abort()
	}

	c.Next()
}
