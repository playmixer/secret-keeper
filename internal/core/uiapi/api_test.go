package uiapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/playmixer/secret-keeper/internal/adapter/models"
	"github.com/playmixer/secret-keeper/internal/adapter/storage/file"
)

func Test_keepClient_eventGetExternalMetaDatas(t *testing.T) {
	tests := []struct {
		name     string
		want     *[]models.MetaDataItem
		fRequest func(method, url string, data *[]byte) (*http.Response, error)
		wantErr  bool
	}{
		{
			name: "ok",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				res := tHalderGetDatasResponse{
					tResultResponse: tResultResponse{Status: true},
					Data:            []tHandlerGetData{},
				}
				bRes, err := json.Marshal(res)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				_, err = w.Write(bRes)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				return w.Result(), nil
			},
			want:    &[]models.MetaDataItem{},
			wantErr: false,
		},
		{
			name: "empty response",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				_, err := w.Write([]byte{})
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				return w.Result(), nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "not auth",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				_, err := w.Write([]byte{})
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				w.WriteHeader(http.StatusUnauthorized)
				return w.Result(), nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "request error",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				_, err := w.Write([]byte{})
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				w.WriteHeader(http.StatusUnauthorized)
				return w.Result(), errors.New("any")
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := file.Init()
			assert.NoError(t, err)
			k, err := New(context.TODO(), s, zap.NewNop(), SetEnableWorker(false))
			assert.NoError(t, err)
			k.newRequest = tt.fRequest
			got, err := k.eventGetExternalMetaDatas()
			if (err != nil) != tt.wantErr {
				t.Errorf("keepClient.eventGetExternalMetaDatas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("keepClient.eventGetExternalMetaDatas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_keepClient_EventAuthorization(t *testing.T) {
	tests := []struct {
		name     string
		want     *tSignInResponse
		fRequest func(method, url string, data *[]byte) (*http.Response, error)
		wantErr  bool
	}{
		{
			name: "ok",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				res := tSignInResponse{
					tResultResponse: tResultResponse{Status: true},
				}
				bRes, err := json.Marshal(res)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				_, err = w.Write(bRes)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				return w.Result(), nil
			},
			want:    &tSignInResponse{},
			wantErr: false,
		},
		{
			name: "empty response",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				_, err := w.Write([]byte{})
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				return w.Result(), nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "not auth",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				_, err := w.Write([]byte{})
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				w.WriteHeader(http.StatusUnauthorized)
				return w.Result(), nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "request error",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				return nil, errors.New("any")
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := file.Init()
			assert.NoError(t, err)
			k, err := New(context.TODO(), s, zap.NewNop(), SetEnableWorker(false))
			assert.NoError(t, err)
			k.newRequest = tt.fRequest
			err = k.EventAuthorization("user", "password")
			if (err != nil) != tt.wantErr {
				t.Errorf("keepClient.EventAuthorization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_keepClient_eventGetExternalData(t *testing.T) {
	tests := []struct {
		name     string
		want     *models.MetaDataItem
		fRequest func(method, url string, data *[]byte) (*http.Response, error)
		wantErr  bool
	}{
		{
			name: "ok",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				res := tSignInResponse{
					tResultResponse: tResultResponse{Status: true},
				}
				bRes, err := json.Marshal(res)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				_, err = w.Write(bRes)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				return w.Result(), nil
			},
			want:    &models.MetaDataItem{},
			wantErr: false,
		},
		{
			name: "empty response",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				_, err := w.Write([]byte{})
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				return w.Result(), nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "not auth",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				_, err := w.Write([]byte{})
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				w.WriteHeader(http.StatusUnauthorized)
				return w.Result(), nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "request error",
			fRequest: func(method, url string, data *[]byte) (*http.Response, error) {
				return nil, errors.New("any")
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := file.Init()
			assert.NoError(t, err)
			k, err := New(context.TODO(), s, zap.NewNop(), SetEnableWorker(false))
			assert.NoError(t, err)
			k.newRequest = tt.fRequest
			got, err := k.eventGetExternalData(1)
			if (err != nil) != tt.wantErr {
				t.Errorf("keepClient.eventGetExternalData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got == nil {
				t.Errorf("keepClient.eventGetExternalData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_keepClient_EventLogout(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		req     func(method, url string, data *[]byte) (*http.Response, error)
	}{
		{
			name:    "ok",
			wantErr: false,
			req: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				res := tSignInResponse{
					tResultResponse: tResultResponse{Status: true},
					AccessToken:     "test",
				}
				bRes, err := json.Marshal(res)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				_, err = w.Write(bRes)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				return w.Result(), nil
			},
		},
		{
			name:    "not auth",
			wantErr: false,
			req: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				res := tSignInResponse{
					tResultResponse: tResultResponse{Status: false},
					AccessToken:     "",
				}
				bRes, err := json.Marshal(res)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				_, err = w.Write(bRes)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				w.WriteHeader(http.StatusUnauthorized)
				return w.Result(), nil
			},
		},
		{
			name:    "error",
			wantErr: false,
			req: func(method, url string, data *[]byte) (*http.Response, error) {
				w := httptest.NewRecorder()
				res := tSignInResponse{
					tResultResponse: tResultResponse{Status: true},
					AccessToken:     "test",
				}
				bRes, err := json.Marshal(res)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				_, err = w.Write(bRes)
				if err != nil {
					return nil, fmt.Errorf("any error: %w", err)
				}
				w.WriteHeader(http.StatusUnauthorized)
				return w.Result(), nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := file.Init()
			assert.NoError(t, err)
			k, err := New(context.TODO(), s, zap.NewNop(), SetEnableWorker(false))
			assert.NoError(t, err)
			k.newRequest = tt.req
			err = k.EventAuthorization("user", "password")
			assert.NoError(t, err)
			if err := k.EventLogout(); (err != nil) != tt.wantErr {
				t.Errorf("keepClient.EventLogout() error = %v, wantErr %v", err, tt.wantErr)
			}
			_ = os.RemoveAll("./data")
		})
	}
}
