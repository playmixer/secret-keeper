package uiapi

import (
	"github.com/playmixer/secret-keeper/internal/adapter/models"
)

type tResultResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
	Status  bool   `json:"status"`
}

type tSignInRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type tSignInResponse struct {
	AccessToken string `json:"access_token"`
	tResultResponse
}

type tRegistrationRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type tRegistrationResponse struct {
	AccessToken string `json:"access_token"`
	tResultResponse
}

type tHandlerGetData struct {
	Title     string          `json:"title"`
	DataType  models.DataType `json:"data_type"`
	ID        uint            `json:"id"`
	UpdatedAt int64           `json:"update_dt"`
	IsDeleted bool            `json:"is_deleted"`
}

type tHalderGetDatasResponse struct {
	tResultResponse
	Data []tHandlerGetData `json:"data"`
}
