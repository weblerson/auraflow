package util

import (
	"os"
	"testing"
)

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultVal   string
		envValue     string
		setEnv       bool
		expectedVal  string
	}{
		{
			name:        "variável existe",
			key:         "TEST_VAR_EXISTS",
			defaultVal:  "default",
			envValue:    "from-env",
			setEnv:      true,
			expectedVal: "from-env",
		},
		{
			name:        "variável não existe retorna default",
			key:         "TEST_VAR_NOT_EXISTS",
			defaultVal:  "default-value",
			envValue:    "",
			setEnv:      false,
			expectedVal: "default-value",
		},
		{
			name:        "variável vazia retorna default",
			key:         "TEST_VAR_EMPTY",
			defaultVal:  "fallback",
			envValue:    "",
			setEnv:      true,
			expectedVal: "fallback",
		},
		{
			name:        "default vazio e variável não existe",
			key:         "TEST_VAR_NO_DEFAULT",
			defaultVal:  "",
			envValue:    "",
			setEnv:      false,
			expectedVal: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			got := GetEnvOrDefault(tt.key, tt.defaultVal)
			if got != tt.expectedVal {
				t.Errorf("GetEnvOrDefault() = %v, want %v", got, tt.expectedVal)
			}
		})
	}
}
