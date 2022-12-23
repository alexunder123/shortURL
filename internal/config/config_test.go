package config

import (
	"os"
	"reflect"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name string
		want *Param
	}{
		{
			name: "test1",
			want: &Param{
				Server:  "127.0.0.1:8080",
				URL:     "http://127.0.0.1:8080",
				Storage: "jsonDB.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("SERVER_ADDRESS", "127.0.0.1:8080")
			os.Setenv("BASE_URL", "http://127.0.0.1:8080")
			os.Setenv("FILE_STORAGE_PATH", "jsonDB.json")
			if got := NewEnv(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
