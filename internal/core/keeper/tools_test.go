package keeper

import (
	"testing"
)

func Test_hashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "ok",
			password: "user",
			wantErr:  false,
		},
		{
			name:     "logn password",
			password: "1222222222222222222222222222222222222222222222222222222222222222222222222",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("hashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("hashPassword() = %v, empty", got)
			}
		})
	}
}

func Test_checkPasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "ok",
			password: "test",
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hashPassword(tt.password)
			if err != nil {
				t.Error("failed generate hash password")
			}
			if got := checkPasswordHash(tt.password, hash); got != tt.want {
				t.Errorf("checkPasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
