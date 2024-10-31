package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/playmixer/secret-keeper/internal/adapter/api/rest"
	"github.com/playmixer/secret-keeper/internal/adapter/keeperr"
	"github.com/playmixer/secret-keeper/internal/adapter/models"
	"github.com/playmixer/secret-keeper/internal/core/config"
	"github.com/playmixer/secret-keeper/internal/core/keeper"
	"github.com/playmixer/secret-keeper/internal/mocks/storage/database"
)

func TestServer_handlerRegistration(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		login    string
		password string
		status   int
		wontErr  error
	}{
		{
			name:     "created",
			login:    "user",
			password: "pass",
			status:   http.StatusCreated,
			wontErr:  nil,
		},
		{
			name:     "empty login",
			login:    "",
			password: "",
			status:   http.StatusBadRequest,
			wontErr:  keeper.ErrLoginNotValid,
		},
		{
			name:     "empty password",
			login:    "user",
			password: "",
			status:   http.StatusBadRequest,
			wontErr:  keeper.ErrPasswordNotValid,
		},
		{
			name:     "not unique",
			login:    "user",
			password: "pass",
			status:   http.StatusConflict,
			wontErr:  keeperr.ErrLoginNotUnique,
		},
		{
			name:     "error",
			login:    "user",
			password: "pass",
			status:   http.StatusInternalServerError,
			wontErr:  keeperr.ErrLoginNotUnique,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg, err := config.Init(false)
			assert.NoError(t, err)

			storeMock := database.NewMockStorage(ctrl)

			if tt.status == http.StatusConflict {
				storeMock.EXPECT().
					Registration(ctx, gomock.Any(), gomock.Any()).
					Return(tt.wontErr).
					Times(1)
			}
			if tt.status == http.StatusCreated {
				storeMock.EXPECT().
					Registration(ctx, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			}
			if tt.status == http.StatusInternalServerError {
				storeMock.EXPECT().
					Registration(ctx, gomock.Any(), gomock.Any()).
					Return(errors.New("another")).
					Times(1)
			}
			keep, err := keeper.New(storeMock, keeper.SetEncryptKey(cfg.EncryptKey))
			assert.NoError(t, err)

			server, err := rest.New(keep,
				rest.SetConfig(*cfg.Rest),
				rest.SetLogger(zap.NewNop()),
				rest.SetSecretKey("qwe"),
				rest.SetSSLEnable(true))
			assert.NoError(t, err)
			engin := server.Engin()

			w := httptest.NewRecorder()
			body := strings.NewReader(fmt.Sprintf(`{"login":%q, "password":%q}`, tt.login, tt.password))
			r := httptest.NewRequest(http.MethodPost, "/api/v0/auth/registration", body)

			engin.ServeHTTP(w, r)

			result := w.Result()

			assert.Equal(t, tt.status, result.StatusCode)

			err = result.Body.Close()
			assert.NoError(t, err)
		})
	}
}

func TestServer_handlerLogin(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		login    string
		password string
		status   int
		wontErr  error
	}{
		{
			name:     "ok",
			login:    "user",
			password: "user",
			status:   http.StatusOK,
			wontErr:  nil,
		},
		{
			name:     "empty request",
			login:    "",
			password: "",
			status:   http.StatusBadRequest,
			wontErr:  keeper.ErrLoginNotValid,
		},
		{
			name:     "empty login",
			login:    "",
			password: "",
			status:   http.StatusUnauthorized,
			wontErr:  keeper.ErrPasswordNotValid,
		},
		{
			name:     "empty password",
			login:    "123",
			password: "",
			status:   http.StatusUnauthorized,
			wontErr:  keeper.ErrPasswordNotValid,
		},
		{
			name:     "error",
			login:    "user",
			password: "pass",
			status:   http.StatusInternalServerError,
			wontErr:  errors.New("any"),
		},
		{
			name:     "not found",
			login:    "user",
			password: "pass",
			status:   http.StatusUnauthorized,
			wontErr:  keeperr.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg, err := config.Init(false)
			assert.NoError(t, err)

			storeMock := database.NewMockStorage(ctrl)

			if tt.status == http.StatusUnauthorized {
				if errors.Is(tt.wontErr, keeperr.ErrNotFound) {
					storeMock.EXPECT().
						GetUserByLogin(ctx, tt.login).
						Return(nil, tt.wontErr).
						Times(1)
				}
			}
			if tt.status == http.StatusOK {
				storeMock.EXPECT().
					GetUserByLogin(ctx, tt.login).
					Return(&models.User{
						Model:        gorm.Model{ID: 1},
						Login:        "user",
						PasswordHash: "$2a$14$M2qLheAVBq/0yqT6NBUleewVIjlhOY4EqzCfEdgg3M0vBvKJA6Ct.",
					}, nil).
					Times(1)
			}
			if tt.status == http.StatusInternalServerError {
				storeMock.EXPECT().
					GetUserByLogin(ctx, tt.login).
					Return(nil, tt.wontErr).
					Times(1)
			}
			keep, err := keeper.New(storeMock, keeper.SetEncryptKey(cfg.EncryptKey))
			assert.NoError(t, err)

			server, err := rest.New(keep)
			assert.NoError(t, err)
			engin := server.Engin()

			w := httptest.NewRecorder()
			reqBody := fmt.Sprintf(`{"login":%q, "password":%q}`, tt.login, tt.password)
			if tt.status == http.StatusBadRequest {
				reqBody = ""
			}
			body := strings.NewReader(reqBody)
			r := httptest.NewRequest(http.MethodPost, "/api/v0/auth/login", body)

			engin.ServeHTTP(w, r)

			result := w.Result()

			assert.Equal(t, tt.status, result.StatusCode)

			err = result.Body.Close()
			assert.NoError(t, err)
		})
	}
}

func TestServer_handlerGetDatas(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		token   string
		status  int
		wontErr error
	}{
		{
			name:    "not auth",
			token:   "",
			status:  http.StatusUnauthorized,
			wontErr: nil,
		},
		{
			name:    "data error",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			status:  http.StatusInternalServerError,
			wontErr: errors.New("any"),
		},
		{
			name:    "ok",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			status:  http.StatusOK,
			wontErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storeMock := database.NewMockStorage(ctrl)

			if tt.status == http.StatusInternalServerError {
				storeMock.EXPECT().
					GetMetaDatasByUserID(ctx, uint(1)).
					Return(nil, tt.wontErr).
					Times(1)
			}
			if tt.status == http.StatusOK {
				storeMock.EXPECT().
					GetMetaDatasByUserID(ctx, uint(1)).
					Return(&[]models.Secret{
						{
							Model:    gorm.Model{ID: 1},
							Title:    "test",
							DataType: models.CARD,
							Data:     []byte("test"),
						},
					}, tt.wontErr).
					Times(1)
			}
			keep, err := keeper.New(storeMock, keeper.SetEncryptKey("secret_key"))
			assert.NoError(t, err)

			server, err := rest.New(keep)
			assert.NoError(t, err)
			engin := server.Engin()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/v0/user/data", http.NoBody)
			r.Header.Add("Authorization", "Bearer "+tt.token)
			engin.ServeHTTP(w, r)

			result := w.Result()

			assert.Equal(t, tt.status, result.StatusCode)

			err = result.Body.Close()
			assert.NoError(t, err)
		})
	}
}

func TestServer_handlerGetData(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		id      string
		token   string
		status  int
		wontErr error
	}{
		{
			name:    "not auth",
			id:      "1",
			token:   "",
			status:  http.StatusUnauthorized,
			wontErr: nil,
		},
		{
			name:    "data error",
			id:      "1",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			status:  http.StatusInternalServerError,
			wontErr: errors.New("any"),
		},
		{
			name:    "ok",
			id:      "1",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			status:  http.StatusOK,
			wontErr: nil,
		},
		{
			name:    "bad id",
			id:      "1a2",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			status:  http.StatusBadRequest,
			wontErr: nil,
		},
		{
			name:    "no content",
			id:      "1",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			status:  http.StatusNoContent,
			wontErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storeMock := database.NewMockStorage(ctrl)

			if tt.status == http.StatusInternalServerError {
				storeMock.EXPECT().
					GetSecret(ctx, uint(1)).
					Return(&models.Secret{
						Model:    gorm.Model{ID: 1},
						Title:    "test",
						DataType: models.CARD,
						Data: []byte("\x108\x01su#\xa9̽\xf6\xc45sD\xe5\x11S\x9e\xf2\xb8\xe3\x18(\xe7f+\n" +
							"\xcb\x03c\x1aEވ\xb8\x04\xe9:\xe1\x9dh\x16R\x05'\x8d\x89\x00\x95\xc6^\xe4\xe0\x15\x92%" +
							"}\xc6ƴ\x86X#\xcbF\xf7\xb2\xb7\xc1\x1di\x8b \x18\x86>\xa0\xd4,\x12\x00\xd3%[\\\x1b(" +
							"\x8b0ţp H\aM\x19\x9d\x1e-\x9f\xc8\xc9b\x1a~\xe6e\x96\xb7xA\f\xc0}X"),
					}, tt.wontErr).
					Times(1)
			}
			if tt.status == http.StatusOK {
				storeMock.EXPECT().
					GetSecret(ctx, uint(1)).
					Return(&models.Secret{
						Model:    gorm.Model{ID: 1},
						Title:    "test",
						DataType: models.CARD,
						Data: []byte("\x108\x01su#\xa9̽\xf6\xc45sD\xe5\x11S\x9e\xf2\xb8\xe3\x18(\xe7f+\n" +
							"\xcb\x03c\x1aEވ\xb8\x04\xe9:\xe1\x9dh\x16R\x05'\x8d\x89\x00\x95\xc6^\xe4\xe0\x15\x92%" +
							"}\xc6ƴ\x86X#\xcbF\xf7\xb2\xb7\xc1\x1di\x8b \x18\x86>\xa0\xd4,\x12\x00\xd3%[\\\x1b(" +
							"\x8b0ţp H\aM\x19\x9d\x1e-\x9f\xc8\xc9b\x1a~\xe6e\x96\xb7xA\f\xc0}X"),
					}, tt.wontErr).
					Times(1)
			}
			if tt.status == http.StatusNoContent {
				storeMock.EXPECT().
					GetSecret(ctx, uint(1)).
					Return(nil, tt.wontErr).
					Times(1)
			}
			keep, err := keeper.New(storeMock, keeper.SetEncryptKey("RZLMAOIOuljexYLh5S47O9kfVI7O1Ll0"))
			assert.NoError(t, err)

			server, err := rest.New(keep)
			assert.NoError(t, err)
			engin := server.Engin()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/v0/user/data/"+tt.id, http.NoBody)
			r.Header.Add("Authorization", "Bearer "+tt.token)
			engin.ServeHTTP(w, r)

			result := w.Result()

			assert.Equal(t, tt.status, result.StatusCode)

			err = result.Body.Close()
			assert.NoError(t, err)
		})
	}
}

func TestServer_handlerNewData(t *testing.T) {
	card := models.Card{
		Title:  "test",
		Number: "12341231231",
		PIN:    "123",
		CVV:    "123",
		Expiry: "",
	}
	bCard, err := json.Marshal(card)
	assert.NoError(t, err)
	ctx := context.Background()
	tests := []struct {
		name    string
		token   string
		request *rest.THandlerNewDataRequest
		status  int
		wontErr error
	}{
		{
			name:  "not auth",
			token: "",
			request: &rest.THandlerNewDataRequest{
				Title:    "test",
				DataType: models.CARD,
				Data:     []byte("test"),
			},
			status:  http.StatusUnauthorized,
			wontErr: nil,
		},
		{
			name:  "data error",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			request: &rest.THandlerNewDataRequest{
				Title:    "test",
				DataType: models.CARD,
				Data:     bCard,
			},
			status:  http.StatusInternalServerError,
			wontErr: errors.New("any"),
		},
		{
			name:    "bad request",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			request: nil,
			status:  http.StatusBadRequest,
			wontErr: nil,
		},
		{
			name:  "ok",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			request: &rest.THandlerNewDataRequest{
				Title:    "test",
				DataType: models.CARD,
				Data:     bCard,
			},
			status:  http.StatusOK,
			wontErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storeMock := database.NewMockStorage(ctrl)

			if tt.status == http.StatusInternalServerError {
				storeMock.EXPECT().
					NewSecret(ctx, gomock.Any()).
					Return(&models.Secret{
						Model:    gorm.Model{ID: 1},
						Title:    "test",
						DataType: models.CARD,
						Data:     []byte("test"),
					}, tt.wontErr).
					Times(1)
			}
			if tt.status == http.StatusOK {
				storeMock.EXPECT().
					NewSecret(ctx, gomock.Any()).
					Return(&models.Secret{
						Model:    gorm.Model{ID: 1},
						Title:    "test",
						DataType: models.CARD,
						Data:     []byte("test"),
					}, tt.wontErr).
					Times(1)
			}
			keep, err := keeper.New(storeMock, keeper.SetEncryptKey("RZLMAOIOuljexYLh5S47O9kfVI7O1Ll0"))
			assert.NoError(t, err)

			server, err := rest.New(keep)
			assert.NoError(t, err)
			engin := server.Engin()

			w := httptest.NewRecorder()
			bBody := []byte{}
			if tt.request != nil {
				bBody, err = json.Marshal(*tt.request)
				assert.NoError(t, err)
			}
			if tt.status == http.StatusBadRequest {
				bBody = []byte{}
			}
			body := bytes.NewReader(bBody)
			r := httptest.NewRequest(http.MethodPost, "/api/v0/user/data", body)
			r.Header.Add("Authorization", "Bearer "+tt.token)
			engin.ServeHTTP(w, r)

			result := w.Result()

			assert.Equal(t, tt.status, result.StatusCode)

			err = result.Body.Close()
			assert.NoError(t, err)
		})
	}
}

func TestServer_handlerDelData(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		token   string
		id      string
		status  int
		wontErr error
	}{
		{
			name:    "not auth",
			token:   "",
			id:      "1",
			status:  http.StatusUnauthorized,
			wontErr: nil,
		},
		{
			name:    "data error",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			id:      "1",
			status:  http.StatusInternalServerError,
			wontErr: errors.New("any"),
		},
		{
			name:    "bad request",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			id:      "1a",
			status:  http.StatusBadRequest,
			wontErr: nil,
		},
		{
			name:    "ok",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			id:      "1",
			status:  http.StatusOK,
			wontErr: nil,
		},
		{
			name:    "no content",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			id:      "1",
			status:  http.StatusNoContent,
			wontErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storeMock := database.NewMockStorage(ctrl)
			if tt.status != http.StatusBadRequest && tt.status != http.StatusUnauthorized {
				storeMock.EXPECT().
					DelSecret(ctx, uint(1)).
					Return(tt.wontErr).
					Times(1)
			}

			keep, err := keeper.New(storeMock, keeper.SetEncryptKey("RZLMAOIOuljexYLh5S47O9kfVI7O1Ll0"))
			assert.NoError(t, err)

			server, err := rest.New(keep)
			assert.NoError(t, err)
			engin := server.Engin()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, "/api/v0/user/data/"+tt.id, http.NoBody)
			r.Header.Add("Authorization", "Bearer "+tt.token)
			engin.ServeHTTP(w, r)

			result := w.Result()

			assert.Equal(t, tt.status, result.StatusCode)

			err = result.Body.Close()
			assert.NoError(t, err)
		})
	}
}

func TestServer_handlerUpdData(t *testing.T) {
	card := models.Card{
		Title:  "test",
		Number: "12341231231",
		PIN:    "123",
		CVV:    "123",
		Expiry: "",
	}
	bCard, err := json.Marshal(card)
	assert.NoError(t, err)
	ctx := context.Background()
	tests := []struct {
		name    string
		token   string
		id      string
		request *rest.THandlerUpdDataRequest
		status  int
		wontErr error
	}{
		{
			name:  "not auth",
			token: "",
			id:    "1",
			request: &rest.THandlerUpdDataRequest{
				Title:    "test",
				DataType: models.CARD,
				Data:     bCard,
			},
			status:  http.StatusUnauthorized,
			wontErr: nil,
		},
		{
			name:  "data error",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			id:    "1",
			request: &rest.THandlerUpdDataRequest{
				Title:    "test",
				DataType: models.CARD,
				Data:     bCard,
			},
			status:  http.StatusInternalServerError,
			wontErr: errors.New("any"),
		},
		{
			name:    "bad request",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			id:      "1",
			request: nil,
			status:  http.StatusBadRequest,
			wontErr: nil,
		},
		{
			name:  "ok",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.DLIZZPxXkBkwKLXIa-HdWu4zBV5ZaUDf6E4_Q-X-jG0",
			id:    "1",
			request: &rest.THandlerUpdDataRequest{
				Title:    "test",
				DataType: models.CARD,
				Data:     bCard,
			},
			status:  http.StatusOK,
			wontErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storeMock := database.NewMockStorage(ctrl)

			if tt.status == http.StatusInternalServerError {
				storeMock.EXPECT().
					UpdSecret(ctx, gomock.Any()).
					Return(&models.Secret{
						Model:    gorm.Model{ID: 1},
						Title:    "test",
						DataType: models.CARD,
						Data: []byte(`\x108\x01su#\xa9̽\xf6\xc45sD\xe5\x11S\x9e\xf2\xb8\xe3\x18
						(\xe7f+\n\xcb\x03c\x1aEވ\xb8\x04\xe9:\xe1\x9dh\x16R\x05'\x8d\x89\x00\x95
						\xc6^\xe4\xe0\x15\x92%}\xc6ƴ\x86X#\xcbF\xf7\xb2\xb7\xc1\x1di\x8b \x18
						\x86>\xa0\xd4,\x12\x00\xd3%[\\\x1b(\x8b0ţp H\aM\x19\x9d\x1e-\x9f\xc8
						\xc9b\x1a~\xe6e\x96\xb7xA\f\xc0}X`),
					}, tt.wontErr).
					Times(1)
			}
			if tt.status == http.StatusOK {
				storeMock.EXPECT().
					UpdSecret(ctx, gomock.Any()).
					Return(&models.Secret{
						Model:    gorm.Model{ID: 1},
						Title:    "test",
						DataType: models.CARD,
						Data: []byte(`\x108\x01su#\xa9̽\xf6\xc45sD\xe5\x11S\x9e\xf2\xb8\xe3\x18
						(\xe7f+\n\xcb\x03c\x1aEވ\xb8\x04\xe9:\xe1\x9dh\x16R\x05'\x8d\x89\x00\x95
						\xc6^\xe4\xe0\x15\x92%}\xc6ƴ\x86X#\xcbF\xf7\xb2\xb7\xc1\x1di\x8b \x18
						\x86>\xa0\xd4,\x12\x00\xd3%[\\\x1b(\x8b0ţp H\aM\x19\x9d\x1e-\x9f\xc8
						\xc9b\x1a~\xe6e\x96\xb7xA\f\xc0}X`),
					}, tt.wontErr).
					Times(1)
			}
			keep, err := keeper.New(storeMock, keeper.SetEncryptKey("RZLMAOIOuljexYLh5S47O9kfVI7O1Ll0"))
			assert.NoError(t, err)

			server, err := rest.New(keep)
			assert.NoError(t, err)
			engin := server.Engin()

			w := httptest.NewRecorder()
			bBody := []byte{}
			if tt.request != nil {
				bBody, err = json.Marshal(*tt.request)
				assert.NoError(t, err)
			}
			if tt.status == http.StatusBadRequest {
				bBody = []byte{}
			}
			body := bytes.NewReader(bBody)
			r := httptest.NewRequest(http.MethodPut, "/api/v0/user/data/"+tt.id, body)
			r.Header.Add("Authorization", "Bearer "+tt.token)
			engin.ServeHTTP(w, r)

			result := w.Result()

			assert.Equal(t, tt.status, result.StatusCode)

			err = result.Body.Close()
			assert.NoError(t, err)
		})
	}
}
