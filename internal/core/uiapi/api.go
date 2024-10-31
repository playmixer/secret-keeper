package uiapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.uber.org/zap"

	"github.com/playmixer/secret-keeper/internal/adapter/api/rest"
	"github.com/playmixer/secret-keeper/internal/adapter/models"
)

var (
	errMessageFailedReadBody   = "failed read body"
	errMessageFailedUnmarshal  = "failed unmarshal"
	errMessageFailedCreateCard = "failed create card"
	errMessageFailedRequest    = "failed request"
	errMessageFailedCloseBody  = "failed close body response"

	formatStringError = "%s: %w"
)

func (k *keepClient) EventAuthorization(login, password string) error {
	req := tSignInRequest{
		Login:    login,
		Password: password,
	}

	bReq, err := json.Marshal(req)
	if err != nil {
		k.log.Error("failed marshal request", zap.Error(err))
		return fmt.Errorf("failed marshal request: %w", err)
	}

	r, err := k.newRequest(http.MethodPost, k.apiURL+"/api/v0/auth/login", &bReq)
	if err != nil {
		k.log.Error("failed create request", zap.Error(err))
		return fmt.Errorf("failed create request: %w", err)
	}

	res, err := io.ReadAll(r.Body)
	if err != nil {
		k.log.Error(errMessageFailedReadBody, zap.Error(err))
		return fmt.Errorf(formatStringError, errMessageFailedReadBody, err)
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			k.log.Error("failed close body", zap.Error(err))
		}
	}()

	if r.StatusCode == http.StatusUnauthorized {
		return errors.New("неверные данные авторизации")
	}

	result := tSignInResponse{}
	err = json.Unmarshal(res, &result)
	if err != nil {
		k.log.Error(errMessageFailedUnmarshal, zap.Error(err))
		return fmt.Errorf(formatStringError, errMessageFailedUnmarshal, err)
	}

	err = k.store.Open(login)
	if err != nil {
		k.log.Error("failed open store", zap.Error(err))
		return fmt.Errorf("failed open store: %w", err)
	}

	k.token = result.AccessToken
	return nil
}

func (k *keepClient) EventLogout() error {
	k.log.Debug("event logout", zap.String("token", k.token))
	if k.token == "" {
		return nil
	}

	err := k.store.Close()
	if err != nil {
		k.log.Error("failed close store", zap.Error(err))
		return fmt.Errorf("failed close store: %w", err)
	}
	k.log.Debug("store closed")
	k.token = ""
	return nil
}

func (k *keepClient) EventRegistration(login, password, password2 string) error {
	if password != password2 {
		return errors.New("повторный пароль не совпадает")
	}
	req := tRegistrationRequest{
		Login:    login,
		Password: password,
	}
	bReq, err := json.Marshal(req)
	if err != nil {
		k.log.Error("failed marshal request", zap.Error(err))
		return fmt.Errorf("failed marshal request: %w", err)
	}

	r, err := k.newRequest(http.MethodPost, k.apiURL+"/api/v0/auth/registration", &bReq)
	if err != nil {
		k.log.Error("failed create request", zap.Error(err))
		return fmt.Errorf("failed create request: %w", err)
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			k.log.Error("failed close body", zap.Error(err))
		}
	}()

	res, err := io.ReadAll(r.Body)
	if err != nil {
		k.log.Error(errMessageFailedReadBody, zap.Error(err))
		return fmt.Errorf(formatStringError, errMessageFailedReadBody, err)
	}

	result := tRegistrationResponse{}
	err = json.Unmarshal(res, &result)
	if err != nil {
		k.log.Error(errMessageFailedUnmarshal, zap.Error(err))
		return fmt.Errorf(formatStringError, errMessageFailedUnmarshal, err)
	}

	if r.StatusCode != http.StatusCreated {
		if r.StatusCode == http.StatusBadRequest {
			return errors.New("Ошибка запроса: " + result.Error)
		}
		if r.StatusCode == http.StatusConflict {
			return errors.New("логин уже занят")
		}
		return errors.New("Ошибка: " + result.Error)
	}

	return nil
}

func (k *keepClient) EventNewCard(eID uint, title, number, cvv, pin, date string) error {
	card := &models.Card{
		Title:  title,
		Number: number,
		PIN:    pin,
		CVV:    cvv,
		Expiry: date,
	}

	bDate, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed marshal card: %w", err)
	}
	_, err = k.store.NewData(eID, 0, title, models.CARD, &bDate)
	if err != nil {
		k.log.Error(errMessageFailedCreateCard, zap.Error(err))
		return fmt.Errorf(formatStringError, errMessageFailedCreateCard, err)
	}

	return nil
}

func (k *keepClient) EventEditCard(id int64, title, number, cvv, pin, date string) error {
	m, _, err := k.store.GetData(id)
	if err != nil {
		return fmt.Errorf("failed get data: %w", err)
	}
	card := &models.Card{
		Title:  title,
		Number: number,
		PIN:    pin,
		CVV:    cvv,
		Expiry: date,
	}

	bDate, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed marshal card: %w", err)
	}
	m.Title = title
	m.UpdateDT = k.store.UpdateDate()

	err = k.store.EditData(id, m, &bDate)
	if err != nil {
		k.log.Error("failed edit card", zap.Error(err))
		return fmt.Errorf("failed edit card %w", err)
	}

	return nil
}

func (k *keepClient) EventDeleteCard(id int64) error {
	return k.eventDeleteData(id)
}

func (k *keepClient) EventGetMetaDatas() (*[]models.FileMetaDataItem, error) {
	data, err := k.store.GetAll()
	if err != nil {
		k.log.Error("failed get data", zap.Error(err))
		return nil, fmt.Errorf("failed get data: %w", err)
	}

	return data, nil
}

func (k *keepClient) EventGetCard(id int64) (*models.Card, error) {
	_, data, err := k.store.GetData(id)
	if err != nil {
		k.log.Error("failed get card", zap.Error(err))
		return nil, fmt.Errorf("failed get card from store: %w", err)
	}

	card := &models.Card{}
	err = json.Unmarshal(*data, card)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal data: %w", err)
	}

	return card, nil
}

func (k *keepClient) EventGetText(id int64) (*models.Text, error) {
	_, data, err := k.store.GetData(id)
	if err != nil {
		k.log.Error("failed get card", zap.Error(err))
		return nil, fmt.Errorf("failed get card from store: %w", err)
	}

	text := &models.Text{}
	err = json.Unmarshal(*data, text)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal data: %w", err)
	}

	return text, nil
}

func (k *keepClient) EventEditText(id int64, title, text string) error {
	txt := &models.Text{
		Title: title,
		Text:  text,
	}
	bData, err := json.Marshal(txt)
	if err != nil {
		return fmt.Errorf("failed marshal text to byte: %w", err)
	}

	m, _, err := k.store.GetData(id)
	if err != nil {
		return fmt.Errorf("failed get data from store id=`%v`: %w", id, err)
	}
	m.Title = title
	m.UpdateDT = k.store.UpdateDate()

	err = k.store.EditData(id, m, &bData)
	if err != nil {
		k.log.Error("failed edit text", zap.Error(err))
		return fmt.Errorf("failed edit text: %w", err)
	}

	return nil
}

func (k *keepClient) EventDeleteText(id int64) error {
	return k.eventDeleteData(id)
}

func (k *keepClient) eventDeleteData(id int64) error {
	err := k.store.DelData(id)
	if err != nil {
		k.log.Error("failed delete data", zap.Error(err), zap.Int64("id", id))
		return fmt.Errorf("failed delete data %w", err)
	}

	return nil
}

func newRequest(k *keepClient) keepRequest {
	return func(method, url string, data *[]byte) (*http.Response, error) {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Transport: tr,
		}

		if data == nil {
			data = &[]byte{}
		}

		body := bytes.NewBuffer(*data)
		var req *http.Request
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, fmt.Errorf("failed create http client: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+k.token)
		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf(formatStringError, errMessageFailedRequest, err)
		}

		return res, nil
	}
}

func (k *keepClient) eventGetExternalMetaDatas() (*[]models.MetaDataItem, error) {
	r, err := k.newRequest(http.MethodGet, k.apiURL+"/api/v0/user/data", nil)
	if err != nil {
		return nil, fmt.Errorf(formatStringError, errMessageFailedRequest, err)
	}

	res, err := io.ReadAll(r.Body)
	if err != nil {
		k.log.Error(errMessageFailedReadBody, zap.Error(err))
		return nil, fmt.Errorf(formatStringError, errMessageFailedReadBody, err)
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			k.log.Error(errMessageFailedCloseBody, zap.Error(err))
		}
	}()

	if r.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("failed authenticate user")
	}

	data := tHalderGetDatasResponse{}
	err = json.Unmarshal(res, &data)
	if err != nil {
		k.log.Error(errMessageFailedUnmarshal, zap.Error(err))
		return nil, fmt.Errorf(formatStringError, errMessageFailedUnmarshal, err)
	}

	result := []models.MetaDataItem{}
	for _, d := range data.Data {
		result = append(result, models.MetaDataItem{
			ID:        d.ID,
			Title:     d.Title,
			DataType:  d.DataType,
			UpdatedDT: d.UpdatedAt,
			IsDeleted: d.IsDeleted,
		})
	}

	return &result, nil
}

func (k *keepClient) eventGetExternalData(id uint) (*models.MetaDataItem, error) {
	r, err := k.newRequest(http.MethodGet, fmt.Sprintf("%s/api/v0/user/data/%v", k.apiURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf(formatStringError, errMessageFailedRequest, err)
	}

	res, err := io.ReadAll(r.Body)
	if err != nil {
		k.log.Error(errMessageFailedReadBody, zap.Error(err))
		return nil, fmt.Errorf(formatStringError, errMessageFailedReadBody, err)
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			k.log.Error(errMessageFailedCloseBody, zap.Error(err))
		}
	}()

	if r.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("failed authenticate user")
	}

	data := rest.THandlerGetDataResponse{}
	err = json.Unmarshal(res, &data)
	if err != nil {
		k.log.Error(errMessageFailedUnmarshal, zap.Error(err))
		return nil, fmt.Errorf(formatStringError, errMessageFailedUnmarshal, err)
	}

	result := &models.MetaDataItem{
		ID:        data.Data.ID,
		Title:     data.Data.Title,
		DataType:  data.Data.DataType,
		Data:      &data.Data.Data,
		UpdatedDT: data.Data.UpdateDT,
		IsDeleted: data.Data.IsDeleted,
	}

	return result, nil
}

func (k *keepClient) eventUpdExternalData(id uint,
	title string, data *[]byte, dataType models.DataType, updateDT int64) error {
	req := rest.THandlerUpdDataRequest{
		Title:    title,
		DataType: dataType,
		Data:     *data,
		UpdateDT: updateDT,
	}
	bBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed marshal data: %w", err)
	}
	url := fmt.Sprintf("%s/api/v0/user/data/%v", k.apiURL, id)
	r, err := k.newRequest(http.MethodPut, url, &bBody)
	if err != nil {
		return fmt.Errorf(formatStringError, errMessageFailedRequest, err)
	}

	if r.StatusCode != http.StatusOK {
		k.log.Error("api", zap.String("url", url), zap.String("error", "api return not OK"))
		return errors.New("api return not OK")
	}

	res, err := io.ReadAll(r.Body)
	if err != nil {
		k.log.Error(errMessageFailedReadBody, zap.Error(err))
		return fmt.Errorf(formatStringError, errMessageFailedReadBody, err)
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			k.log.Error(errMessageFailedCloseBody, zap.Error(err))
		}
	}()

	response := rest.THandlerUpdDataResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return fmt.Errorf(formatStringError, errMessageFailedUnmarshal, err)
	}

	return nil
}

func (k *keepClient) eventAddExternalData(title string, data *[]byte, dataType models.DataType, updateDT int64) (
	*models.MetaDataItem, error,
) {
	req := rest.THandlerNewDataRequest{
		Title:    title,
		DataType: dataType,
		Data:     *data,
		UpdateDT: updateDT,
	}
	bBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed marshal data: %w", err)
	}
	r, err := k.newRequest(http.MethodPost, k.apiURL+"/api/v0/user/data", &bBody)
	if err != nil {
		return nil, fmt.Errorf(formatStringError, errMessageFailedRequest, err)
	}

	res, err := io.ReadAll(r.Body)
	if err != nil {
		k.log.Error(errMessageFailedReadBody, zap.Error(err))
		return nil, fmt.Errorf(formatStringError, errMessageFailedReadBody, err)
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			k.log.Error(errMessageFailedCloseBody, zap.Error(err))
		}
	}()

	response := rest.THandlerNewDataResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, fmt.Errorf(formatStringError, errMessageFailedUnmarshal, err)
	}

	return &models.MetaDataItem{
		ID:        response.Data.ID,
		Title:     response.Data.Title,
		Data:      data,
		DataType:  response.Data.DataType,
		UpdatedDT: response.Data.UpdateDT,
	}, nil
}

func (k *keepClient) eventDeleteExternalData(id uint) error {
	k.log.Debug("delete external data", zap.Uint("id", id))
	r, err := k.newRequest(http.MethodDelete, fmt.Sprintf("%s/api/v0/user/data/%v", k.apiURL, id), nil)
	if err != nil {
		return fmt.Errorf(formatStringError, errMessageFailedRequest, err)
	}

	res, err := io.ReadAll(r.Body)
	if err != nil {
		k.log.Error(errMessageFailedReadBody, zap.Error(err))
		return fmt.Errorf(formatStringError, errMessageFailedReadBody, err)
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			k.log.Error(errMessageFailedCloseBody, zap.Error(err))
		}
	}()

	response := rest.THandlerNewDataResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return fmt.Errorf(formatStringError, errMessageFailedUnmarshal, err)
	}

	return nil
}

func (k *keepClient) EventNewText(eID uint, title, text string) (*models.FileMetaDataItem, error) {
	txt := &models.Text{
		Title: title,
		Text:  text,
	}

	bTxt, err := json.Marshal(txt)
	if err != nil {
		return nil, fmt.Errorf("failed marshal text: %w", err)
	}

	m, err := k.store.NewData(eID, 0, title, models.TEXT, &bTxt)
	if err != nil {
		k.log.Error(errMessageFailedCreateCard, zap.Error(err))
		return nil, fmt.Errorf(formatStringError, errMessageFailedCreateCard, err)
	}

	return m, nil
}

func (k *keepClient) EventGetPassword(id int64) (*models.Password, error) {
	_, data, err := k.store.GetData(id)
	if err != nil {
		k.log.Error("failed get password", zap.Error(err))
		return nil, fmt.Errorf("failed get password from store: %w", err)
	}

	psw := &models.Password{}
	err = json.Unmarshal(*data, psw)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal data: %w", err)
	}

	return psw, nil
}

func (k *keepClient) EventNewPassword(eID uint, title, site, login, password string) (*models.FileMetaDataItem, error) {
	psw := &models.Password{
		Title:    title,
		Site:     site,
		Password: password,
		Login:    login,
	}

	bPsw, err := json.Marshal(psw)
	if err != nil {
		return nil, fmt.Errorf("failed marshal text: %w", err)
	}

	m, err := k.store.NewData(eID, 0, title, models.PASSWORD, &bPsw)
	if err != nil {
		k.log.Error(errMessageFailedCreateCard, zap.Error(err))
		return nil, fmt.Errorf(formatStringError, errMessageFailedCreateCard, err)
	}

	return m, nil
}

func (k *keepClient) EventEditPassword(id int64, title, site, login, password string) error {
	txt := &models.Password{
		Title:    title,
		Site:     site,
		Password: password,
		Login:    login,
	}
	bData, err := json.Marshal(txt)
	if err != nil {
		return fmt.Errorf("failed marshal password to byte: %w", err)
	}

	m, _, err := k.store.GetData(id)
	if err != nil {
		return fmt.Errorf("failed get data from store id=`%v`: %w", id, err)
	}
	m.Title = title
	m.UpdateDT = k.store.UpdateDate()

	err = k.store.EditData(id, m, &bData)
	if err != nil {
		k.log.Error("failed edit password", zap.Error(err))
		return fmt.Errorf("failed edit password: %w", err)
	}

	return nil
}

func (k *keepClient) EventDeletePassword(id int64) error {
	return k.eventDeleteData(id)
}

func (k *keepClient) EventNewFile(eID uint, title, path string) (*models.FileMetaDataItem, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed check status file: %w", err)
	}
	if stat.Size() > k.fileMaxSize {
		return nil, fmt.Errorf("file exceeds maximum %v bytes size", k.fileMaxSize)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("failed open file")
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed read file: %w", err)
	}

	m, err := k.store.NewData(eID, 0, title, models.BINARY, &data)
	if err != nil {
		k.log.Error(errMessageFailedCreateCard, zap.Error(err))
		return nil, fmt.Errorf(formatStringError, errMessageFailedCreateCard, err)
	}

	m.Filename = stat.Name()
	err = k.store.UpdMeta(m)
	if err != nil {
		return nil, fmt.Errorf("failed update meta data: %w", err)
	}

	return m, nil
}

func (k *keepClient) EventGetFile(id int64) (*models.Binary, error) {
	m, err := k.store.Get(id)
	if err != nil {
		k.log.Error("failed get meta data", zap.Error(err))
		return nil, fmt.Errorf("failed get meta data: %w", err)
	}

	file := &models.Binary{
		Title:    m.Title,
		Filename: m.Filename,
	}

	return file, nil
}

func (k *keepClient) EventEditFile(id int64, title, path string) error {
	m, err := k.store.Get(id)
	if err != nil {
		k.log.Error("failed get meta data", zap.Error(err))
		return fmt.Errorf("failed get meta data: %w", err)
	}

	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed check status file: %w", err)
	}

	f, err := os.Open(path)
	if err != nil {
		return errors.New("failed open file")
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed read file: %w", err)
	}

	err = k.store.EditData(id, m, &data)
	if err != nil {
		k.log.Error("failed edit password", zap.Error(err))
		return fmt.Errorf("failed edit password: %w", err)
	}

	m.Filename = stat.Name()
	m.UpdateDT = k.store.UpdateDate()
	err = k.store.UpdMeta(m)
	if err != nil {
		return fmt.Errorf("failed update meta data: %w", err)
	}

	return nil
}

func (k *keepClient) EventUploadFile(id int64, path string) error {
	if err := k.store.UploadFileToPath(id, path); err != nil {
		return fmt.Errorf("failed upload file: %w", err)
	}
	return nil
}

func (k *keepClient) EventDeleteFile(id int64) error {
	if err := k.eventDeleteData(id); err != nil {
		return fmt.Errorf("failed delete file: %w", err)
	}
	return nil
}
