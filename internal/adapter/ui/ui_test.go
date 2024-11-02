package ui

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/playmixer/secret-keeper/internal/adapter/storage/file"
	"github.com/playmixer/secret-keeper/internal/core/uiapi"
)

func Test_terminal_startPage(t *testing.T) {
	isTrueDraw := true
	type args struct {
		btnSignIn func()
		btnReg    func()
		isDraw    *bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ok",
			args: args{
				btnSignIn: func() {},
				btnReg:    func() {},
				isDraw:    &isTrueDraw,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			lgr := zap.NewNop()

			store, err := file.Init(file.SetLogger(lgr))
			assert.NoError(t, err)
			api, err := uiapi.New(
				ctx,
				store,
				lgr,
			)
			assert.NoError(t, err)

			client, err := New(ctx, api, lgr, SetVersion("test", "test", "test"))
			assert.NoError(t, err)
			go func() {
				client.startPage(tt.args.btnSignIn, tt.args.btnReg, tt.args.isDraw)
			}()
			client.Close()
		})
	}
}

func createUI(t *testing.T) *terminal {
	t.Helper()

	ctx := context.Background()
	lgr := zap.NewNop()

	store, err := file.Init(file.SetLogger(lgr))
	assert.NoError(t, err)
	api, err := uiapi.New(
		ctx,
		store,
		lgr,
	)
	assert.NoError(t, err)

	client, err := New(ctx, api, lgr)
	assert.NoError(t, err)
	return client
}

func Test_terminal_authPage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.authPage()
		})
	}
}

func Test_terminal_registratingPage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.registratingPage()
		})
	}
}

func Test_terminal_mainPage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.mainPage()
		})
	}
}

func Test_terminal_errorPage(t *testing.T) {
	type args struct {
		message string
		okBtn   func()
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ok",
			args: args{
				message: "test",
				okBtn:   func() {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.errorPage(tt.args.message, tt.args.okBtn)
		})
	}
}

func Test_terminal_newCardPage(t *testing.T) {
	tests := []struct {
		name string
		tr   *terminal
	}{
		{
			name: "ok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.newCardPage()
		})
	}
}

func Test_terminal_editCardPage(t *testing.T) {
	type args struct {
		id int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ok",
			args: args{
				id: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.editCardPage(tt.args.id)
		})
	}
}

func Test_terminal_newTextPage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.newTextPage()
		})
	}
}

func Test_terminal_editTextPage(t *testing.T) {
	type args struct {
		id int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ok",
			args: args{
				id: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.editTextPage(tt.args.id)
		})
	}
}

func Test_terminal_newPasswordPage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.newPasswordPage()
		})
	}
}

func Test_terminal_modal(t *testing.T) {
	type args struct {
		message string
		btns    map[string]func()
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ok",
			args: args{
				message: "test",
				btns:    map[string]func(){},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.modal(tt.args.message, tt.args.btns)
		})
	}
}

func Test_terminal_editPasswordPage(t *testing.T) {
	type args struct {
		id int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ok",
			args: args{
				id: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.editPasswordPage(tt.args.id)
		})
	}
}

func Test_terminal_newFilePage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.newFilePage()
		})
	}
}

func Test_terminal_editFilePage(t *testing.T) {
	type args struct {
		id int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ok",
			args: args{
				id: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.editFilePage(tt.args.id)
		})
	}
}

func Test_terminal_uploadFilePage(t *testing.T) {
	type args struct {
		id int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ok",
			args: args{
				id: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createUI(t)
			client.uploadFilePage(tt.args.id)
		})
	}
}
