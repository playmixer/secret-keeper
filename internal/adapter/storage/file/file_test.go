package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/playmixer/secret-keeper/pkg/tools"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := tools.GetMD5Hash("user")
			s, err := Init(SetPath("./test"), SetLogger(zap.NewNop()))
			assert.NoError(t, err)
			err = s.Open("user")
			defer func() {
				_ = os.RemoveAll("./test/")
			}()
			assert.NoError(t, err)
			err = s.Close()
			assert.NoError(t, err)
			stat, err := os.Stat("./test/" + hash)
			assert.NoError(t, err)
			assert.True(t, stat.Size() > 0)
		})
	}
}
