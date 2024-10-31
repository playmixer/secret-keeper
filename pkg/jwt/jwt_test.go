package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJWT_Create(t *testing.T) {
	type args struct {
		secret []byte
		params map[string]string
	}
	tests := []struct {
		name       string
		args       args
		want       string
		wantVerify bool
		wantErr    bool
	}{
		{
			name: "ok",
			args: args{
				secret: []byte("secret"),
				params: map[string]string{
					"user_id": "1",
				},
			},
			want:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSJ9.TcXz_IwlmxO5nPd3m0Yo67WyYptabkqZW4R9HNwPmKE",
			wantVerify: true,
		},
		{
			name: "not veryfy",
			args: args{
				secret: nil,
				params: map[string]string{},
			},
			want:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.LwimMJA3puF3ioGeS-tfczR3370GXBZMIL-bdpu4hOU1",
			wantVerify: false,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testJWT := New(tt.args.secret)
			tkn, err := testJWT.Create(tt.args.params)
			assert.NoError(t, err)
			if tkn != tt.want && tt.wantVerify {
				t.Fatalf("Create() = %v, want %v", tkn, tt.want)
			}
			if !tt.wantVerify {
				tkn = ""
			}
			ok, err := testJWT.Verify(tkn)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			if ok != tt.wantVerify {
				t.Fatalf("Verify() = %v, want %v", ok, tt.wantVerify)
			}

			params, err := testJWT.GetParams(tkn)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			for k, v := range params {
				if va, ok := tt.args.params[k]; ok {
					if va != v {
						t.Fatalf("for key=%s, value want %s, actual %s", k, v, va)
					}
					continue
				}
				t.Fatalf("key %s not found", k)
			}
		})
	}
}
