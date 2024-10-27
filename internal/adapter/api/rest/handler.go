package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/playmixer/secret-keeper/internal/adapter/keeperr"
	"github.com/playmixer/secret-keeper/internal/core/keeper"
	"github.com/playmixer/secret-keeper/pkg/jwt"
)

var (
	errFailedGetData = "failed get datas"
)

// @Summary	Register user
// @Schemes
// @Description	registration user
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			registration	body		tHandlerRegistrationRequest		true	"registration"
// @Success		201				{object}	tHandlerRegistrationResponse	"пользователь успешно зарегистрирован"
// @failure		400				{object}	tResultErrorResponse			"неверный формат запроса"
// @failure		409				{object}	tResultErrorResponse			"логин уже занят"
// @failure		500				"внутренняя ошибка сервера"
// @Router			/auth/registration [post]
func (s *Server) handlerRegistration(c *gin.Context) {
	bBody, statusCode := s.readBody(c)
	if statusCode > 0 {
		c.Writer.WriteHeader(statusCode)
		return
	}

	jBody := tHandlerRegistrationRequest{}
	err := json.Unmarshal(bBody, &jBody)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.keeper.Registration(c.Request.Context(), jBody.Login, jBody.Password)
	if err != nil {
		if errors.Is(err, keeper.ErrLoginNotValid) {
			c.JSON(http.StatusBadRequest, tResultErrorResponse{
				Status: false,
				Error:  "Invalid login",
			})
			return
		}
		if errors.Is(err, keeper.ErrPasswordNotValid) {
			c.JSON(http.StatusBadRequest, tResultErrorResponse{
				Status: false,
				Error:  "Invalid password",
			})
			return
		}
		if errors.Is(err, keeperr.ErrLoginNotUnique) {
			c.JSON(http.StatusConflict, tResultErrorResponse{
				Status: false,
				Error:  "Login not unique",
			})
			return
		}
		c.Writer.WriteHeader(http.StatusInternalServerError)
		s.log.Error("failed create user", zap.Error(err))
		return
	}

	c.JSON(http.StatusCreated, tHandlerRegistrationResponse{
		tResultResponse: tResultResponse{
			Status:  true,
			Message: "Registration successful",
		},
	})
}

// @Summary	Login user
// @Schemes
// @Description	login user
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			auth	body		tHandlerLoginRequest	true	"auth"
// @Success		200		{object}	tHandlerLoginResponse	"пользователь успешно аутентифицирован"
// @failure		400		{object}	tResultErrorResponse	"неверный формат запроса"
// @failure		401		{object}	tResultErrorResponse	"логин или пароль не верный"
// @failure		500		"внутренняя ошибка сервера"
// @Router			/auth/login [post]
func (s *Server) handlerLogin(c *gin.Context) {
	bBody, statusCode := s.readBody(c)
	if statusCode > 0 {
		c.Writer.WriteHeader(statusCode)
		return
	}

	jBody := tHandlerLoginRequest{}
	err := json.Unmarshal(bBody, &jBody)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := s.keeper.Login(c.Request.Context(), jBody.Login, jBody.Password)
	if err != nil {
		if errors.Is(err, keeperr.ErrLoginOrPasswordNotCorrect) ||
			errors.Is(err, keeperr.ErrNotFound) {
			c.JSON(http.StatusUnauthorized, tResultErrorResponse{
				Status: false,
				Error:  "Login or password not correct",
			})
			return
		}
		s.log.Error("failed login user", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	jwtManager := jwt.New(s.secretKey)
	accessToken, err := jwtManager.Create(map[string]string{
		"user_id": strconv.Itoa(int(user.ID)),
	})
	if err != nil {
		s.log.Error("failed create access token", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, tHandlerLoginResponse{
		tResultResponse: tResultResponse{
			Status:  true,
			Message: "User authenticated",
		},
		AccessToken: accessToken,
	})
}

// @Summary	Login user
// @Schemes
// @Description	login user
// @Tags			user
// @Param			Authorization	header	string	true	"authorization"
// @Accept			json
// @Produce		json
// @Success		200	{object}	tHandlerLoginResponse	"пользователь успешно аутентифицирован"
// @failure		500	"внутренняя ошибка сервера"
// @Router			/user/info [get]
func (s *Server) handlerUserInfo(c *gin.Context) {
	userID, err := s.authUserID(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
	})
}

// @Summary	Get Datas
// @Schemes
// @Description	получить данные пользователя
// @Tags			user
// @Param			Authorization	header	string	true	"authorization"
// @Accept			json
// @Produce		json
// @Success		200	{object}	tHalderGetDatasResponse	"данные получены"
// @failure		401	"ошибка авторизации"
// @failure		500	"внутренняя ошибка сервера"
// @Router			/user/data [get]
func (s *Server) handlerGetDatas(c *gin.Context) {
	userID, err := s.authUserID(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	data, err := s.keeper.GetMetaDatasByUserID(c.Request.Context(), userID)
	if err != nil {
		s.log.Error(errFailedGetData, zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}

	res := []tHandlerGetData{}
	_data := *data
	for i := range _data {
		res = append(res, tHandlerGetData{
			ID:        _data[i].ID,
			Title:     _data[i].Title,
			DataType:  _data[i].DataType,
			IsDeleted: _data[i].IsDeleted,
			UpdatedAt: _data[i].UpdateDT,
		})
	}

	c.JSON(http.StatusOK, tHalderGetDatasResponse{
		tResultResponse: tResultResponse{
			Status: true,
		},
		Data: res,
	})
}

// @Summary	Get Data
// @Schemes
// @Description	получить данные
// @Tags			user
// @Param			Authorization	header	string	true	"authorization"
// @Param			id				path	string	true	"id"
// @Accept			json
// @Produce		json
// @Success		200	{object}	THandlerGetDataResponse	"данные получены"
// @failure		204	{object}	tResultErrorResponse	"нет данных"
// @failure		400	{object}	tResultErrorResponse	"ошибка запроса"
// @failure		401	"ошибка авторизации"
// @failure		500	"внутренняя ошибка сервера"
// @Router			/user/data/{id} [get]
func (s *Server) handlerGetData(c *gin.Context) {
	_, err := s.authUserID(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	idS, _ := c.Params.Get("id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		c.JSON(http.StatusBadRequest, tResultErrorResponse{
			Status: false,
			Error:  fmt.Sprintf("Data id `%v` is not correct", id),
		})
		return
	}

	data, err := s.keeper.GetSecret(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNoContent, tResultErrorResponse{
				Status: false,
				Error:  "not found content",
			})
			return
		}
		s.log.Error(errFailedGetData, zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, THandlerGetDataResponse{
		tResultResponse: tResultResponse{
			Status: true,
		},
		Data: tGetData{
			ID:       data.ID,
			Title:    data.Title,
			Data:     data.Data,
			DataType: data.DataType,
			UpdateDT: data.UpdateDT,
		},
	})
}

// @Summary	Delete Data
// @Schemes
// @Description	удаляем данные
// @Tags			user
// @Param			Authorization	header	string	true	"authorization"
// @Param			id				path	string	true	"id"
// @Accept			json
// @Produce		json
// @Success		200	{object}	THandlerDelDataResponse	"данные удалены"
// @failure		204	{object}	tResultErrorResponse	"нет данных"
// @failure		400	{object}	tResultErrorResponse	"ошибка запроса"
// @failure		401	"ошибка авторизации"
// @failure		500	"внутренняя ошибка сервера"
// @Router			/user/data/{id} [delete]
func (s *Server) handlerDelData(c *gin.Context) {
	_, err := s.authUserID(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	idS, _ := c.Params.Get("id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		c.JSON(http.StatusBadRequest, tResultErrorResponse{
			Status: false,
			Error:  fmt.Sprintf("Data id `%v` is not correct", id),
		})
		return
	}

	err = s.keeper.DelSecret(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNoContent, tResultErrorResponse{
				Status: false,
				Error:  "not found content",
			})
			return
		}
		s.log.Error("failed delete data", zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, THandlerDelDataResponse{
		tResultResponse: tResultResponse{
			Status: true,
		},
	})
}

// @Summary	Create Data
// @Schemes
// @Description	создать секрет
// @Tags			user
// @Param			Authorization	header	string					true	"authorization"
// @Param			data			body	THandlerNewDataRequest	true	"данные"
// @Accept			json
// @Produce		json
// @Success		201	{object}	THandlerNewDataResponse	"данные добавлены"
// @failure		400	{object}	tResultErrorResponse	"ошибка запроса"
// @failure		401	"ошибка авторизации"
// @failure		500	"внутренняя ошибка сервера"
// @Router			/user/data [post]
func (s *Server) handlerNewData(c *gin.Context) {
	userID, err := s.authUserID(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	reqData := THandlerNewDataRequest{}
	err = c.ShouldBindJSON(&reqData)
	if err != nil {
		s.log.Error("failed bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, tResultErrorResponse{
			Status: false,
			Error:  "failed bind json",
		})
		return
	}

	data, err := s.keeper.NewSecret(c.Request.Context(),
		&reqData.Data, reqData.Title, reqData.DataType, reqData.UpdateDT, userID)
	if err != nil {
		s.log.Error(errFailedGetData, zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, THandlerNewDataResponse{
		tResultResponse: tResultResponse{
			Status: true,
		},
		Data: tNewData{
			ID:       data.ID,
			Title:    data.Title,
			DataType: data.DataType,
			UpdateDT: data.CreatedAt.UTC().Unix(),
		},
	})
}

// @Summary	Update Data
// @Schemes
// @Description	обновить секрет
// @Tags			user
// @Param			Authorization	header	string					true	"authorization"
// @Param			card			body	THandlerUpdDataRequest	true	"authorization"
// @Param			id				path	string					true	"data id"
// @Accept			json
// @Produce		json
// @Success		200	{object}	THandlerUpdDataResponse	"данные получены"
// @failure		400	{object}	tResultErrorResponse	"ошибка запроса"
// @failure		401	"ошибка авторизации"
// @failure		500	"внутренняя ошибка сервера"
// @Router			/user/data/{id} [put]
func (s *Server) handlerUpdData(c *gin.Context) {
	userID, err := s.authUserID(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	idS, _ := c.Params.Get("id")
	id, err := strconv.Atoi(idS)
	if err != nil {
		c.JSON(http.StatusBadRequest, tResultErrorResponse{
			Status: false,
			Error:  fmt.Sprintf("Data id `%v` is not correct", id),
		})
		return
	}

	rData := THandlerUpdDataRequest{}
	err = c.ShouldBindJSON(&rData)
	if err != nil {
		s.log.Error("failed bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, tResultErrorResponse{
			Status: false,
			Error:  "failed bind json",
		})
		return
	}

	_, err = s.keeper.UpdSecret(c.Request.Context(),
		uint(id), &rData.Data, rData.Title, rData.DataType, rData.UpdateDT, userID)
	if err != nil {
		s.log.Error(errFailedGetData, zap.Error(err))
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, THandlerUpdDataResponse{
		tResultResponse: tResultResponse{
			Status: true,
		},
	})
}
