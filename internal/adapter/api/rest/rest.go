package rest

//	@title		«GophKeeper»
//	@version	1.0
//	@BasePath	/api/v0

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "github.com/playmixer/secret-keeper/docs"
	"github.com/playmixer/secret-keeper/internal/adapter/models"
	"github.com/playmixer/secret-keeper/pkg/jwt"
)

var (
	msgErrorCloseBody = "failed close body"
)

// Keeper - координатор.
type Keeper interface {
	Registration(ctx context.Context, login string, password string) error
	Login(ctx context.Context, login, password string) (*models.User, error)
	GetMetaDatasByUserID(ctx context.Context, userID uint) (*[]models.Secret, error)
	GetSecret(ctx context.Context, ID uint) (*models.Secret, error)
	NewSecret(ctx context.Context,
		data *[]byte, title string, dataType models.DataType, updateDT int64, userID uint) (*models.Secret, error)
	UpdSecret(ctx context.Context, id uint,
		data *[]byte, title string, dataType models.DataType, updateDT int64, userID uint) (*models.Secret, error)
	DelSecret(ctx context.Context, id uint) error
}

// Server - сервер.
type Server struct {
	srv       *http.Server
	log       *zap.Logger
	keeper    Keeper
	secretKey []byte
	sslEnable bool
}

type option func(*Server)

// SetConfig устанавливаем конфигурации сервера.
func SetConfig(cfg Config) option {
	return func(s *Server) {
		s.srv.Addr = cfg.Address
	}
}

func SetLogger(log *zap.Logger) option {
	return func(s *Server) {
		s.log = log
	}
}

func SetSecretKey(secret string) option {
	return func(s *Server) {
		s.secretKey = []byte(secret)
	}
}

func SetSSLEnable(enable bool) option {
	return func(s *Server) {
		s.sslEnable = enable
	}
}

// New создаём рест сервер.
func New(keeper Keeper, options ...option) (*Server, error) {
	s := &Server{
		srv:    &http.Server{},
		keeper: keeper,
		log:    zap.NewNop(),
	}

	for _, opt := range options {
		opt(s)
	}

	return s, nil
}

// Engin - ручки сервера.
func (s *Server) Engin() *gin.Engine {
	r := gin.Default()
	api := r.Group("/api/v0")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/registration", s.handlerRegistration)
			auth.POST("/login", s.handlerLogin)
		}
		user := api.Group("/user")
		user.Use(s.middlewareAuthorization)
		{
			user.GET("/data", s.handlerGetDatas)
			user.GET("/data/:id", s.handlerGetData)
			user.POST("/data", s.handlerNewData)
			user.PUT("/data/:id", s.handlerUpdData)
			user.DELETE("/data/:id", s.handlerDelData)
		}
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

// Run - старт сервера.
func (s *Server) Run() error {
	s.srv.Handler = s.Engin()
	switch s.sslEnable {
	case false:
		if err := s.srv.ListenAndServe(); err != nil {
			return fmt.Errorf("server has failed: %w", err)
		}
	case true:
		if err := s.srv.ListenAndServeTLS("./cert/gophkeeper.crt", "./cert/gophkeeper.key"); err != nil {
			return fmt.Errorf("server has failed: %w", err)
		}
	}

	return nil
}

func (s *Server) readBody(c *gin.Context) ([]byte, int) {
	bBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.log.Error("failed read body", zap.Error(err))
		return []byte{}, http.StatusBadRequest
	}
	defer func() {
		if err := c.Request.Body.Close(); err != nil {
			s.log.Error(msgErrorCloseBody, zap.Error(err))
		}
	}()
	return bBody, 0
}

func (s *Server) authParams(c *gin.Context) (map[string]string, error) {
	var err error
	var res map[string]string
	authHeader := c.Request.Header.Get("Authorization")
	jwtManager := jwt.New(s.secretKey)

	token := strings.Replace(authHeader, "Bearer ", "", 1)
	if res, err = jwtManager.GetParams(token); err != nil {
		return res, fmt.Errorf("failed get params from token: %w", err)
	}

	return res, nil
}

func (s *Server) authUserID(c *gin.Context) (uint, error) {
	params, err := s.authParams(c)
	if err != nil {
		return 0, fmt.Errorf("failed get params from authorization: %w", err)
	}
	if strUserID, ok := params["user_id"]; ok {
		userID, err := strconv.Atoi(strUserID)
		if err != nil {
			return 0, fmt.Errorf("failed convert user_id to int: %w", err)
		}
		return uint(userID), nil
	}
	return 0, errors.New("user_id not found from authorization")
}
