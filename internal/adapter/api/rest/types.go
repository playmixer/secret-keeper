package rest

import (
	"github.com/playmixer/secret-keeper/internal/adapter/models"
)

type tResultResponse struct {
	Message string `json:"message"`
	Status  bool   `json:"status"`
}

type tResultErrorResponse struct {
	Error  string `json:"error"`
	Status bool   `json:"status"`
}

type tHandlerRegistrationRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type tHandlerRegistrationResponse struct {
	tResultResponse
}

type tHandlerLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type tHandlerLoginResponse struct {
	AccessToken string `json:"access_token"`
	tResultResponse
}

type tHandlerGetData struct {
	Title     string          `json:"title"`
	DataType  models.DataType `json:"data_type"`
	UpdatedAt int64           `json:"update_dt"`
	ID        uint            `json:"id"`
	IsDeleted bool            `json:"is_deleted"`
}

type tHalderGetDatasResponse struct {
	tResultResponse
	Data []tHandlerGetData `json:"data"`
}

type THandlerNewDataRequest struct {
	Title    string          `json:"title"`
	DataType models.DataType `json:"data_type"`
	Data     []byte          `json:"data"`
	UpdateDT int64           `json:"update_dt"`
}

type tNewData struct {
	Title    string          `json:"title"`
	DataType models.DataType `json:"data_type"`
	ID       uint            `json:"id"`
	UpdateDT int64           `json:"update_dt"`
}

type THandlerNewDataResponse struct {
	tResultResponse
	Data tNewData `json:"data"`
}

type tGetData struct {
	Title     string          `json:"title"`
	DataType  models.DataType `json:"data_type"`
	Data      []byte          `json:"data"`
	ID        uint            `json:"id"`
	UpdateDT  int64           `json:"update_dt"`
	IsDeleted bool            `json:"is_deleted"`
}

type THandlerGetDataResponse struct {
	tResultResponse
	Data tGetData `json:"data"`
}

type THandlerUpdDataRequest struct {
	Title     string          `json:"title"`
	DataType  models.DataType `json:"data_type"`
	Data      []byte          `json:"data"`
	IsDeleted bool            `json:"is_deleted"`
	UpdateDT  int64           `json:"update_dt"`
}

type THandlerUpdDataResponse struct {
	tResultResponse
}

type THandlerDelDataResponse struct {
	tResultResponse
}
