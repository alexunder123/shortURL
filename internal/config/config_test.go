package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{
			name: "test1",
			want: &Config{
				ServerAddress:   "127.0.0.1:8080",
				BaseURL:         "http://127.0.0.1:8080",
				FileStoragePath: "json_storage.json",
				DatabaseDSN:     "PostgreSQL:localhost",
				SavePlace:       SaveSQL,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("SERVER_ADDRESS", "127.0.0.1:8080")
			os.Setenv("BASE_URL", "http://127.0.0.1:8080")
			os.Setenv("FILE_STORAGE_PATH", "json_storage.json")
			os.Setenv("DATABASE_DSN", "PostgreSQL:localhost")
			got, err := NewConfig()
			require.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
