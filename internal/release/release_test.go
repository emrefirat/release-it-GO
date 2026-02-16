package release

import (
	"os"
	"testing"
)

func TestGetToken(t *testing.T) {
	tests := []struct {
		name       string
		tokenRef   string
		envValue   string
		skipChecks bool
		wantToken  string
		wantErr    bool
	}{
		{
			name:      "token from env",
			tokenRef:  "TEST_TOKEN_REF",
			envValue:  "my-secret-token",
			wantToken: "my-secret-token",
		},
		{
			name:     "missing token without skip",
			tokenRef: "TEST_TOKEN_MISSING",
			envValue: "",
			wantErr:  true,
		},
		{
			name:       "missing token with skip",
			tokenRef:   "TEST_TOKEN_SKIP",
			envValue:   "",
			skipChecks: true,
			wantToken:  "",
		},
		{
			name:    "empty token ref without skip",
			wantErr: true,
		},
		{
			name:       "empty token ref with skip",
			skipChecks: true,
			wantToken:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tokenRef != "" {
				if tt.envValue != "" {
					os.Setenv(tt.tokenRef, tt.envValue)
					defer os.Unsetenv(tt.tokenRef)
				} else {
					os.Unsetenv(tt.tokenRef)
				}
			}

			token, err := getToken(tt.tokenRef, tt.skipChecks)
			if (err != nil) != tt.wantErr {
				t.Errorf("getToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if token != tt.wantToken {
				t.Errorf("getToken() = %q, want %q", token, tt.wantToken)
			}
		})
	}
}
