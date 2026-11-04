package release

import (
	"net/http"
	"time"
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
					_ = os.Setenv(tt.tokenRef, tt.envValue)
					defer func() { _ = os.Unsetenv(tt.tokenRef) }()
				} else {
					_ = os.Unsetenv(tt.tokenRef)
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

func TestNewUploadClient_NoOverallTimeout(t *testing.T) {
	api := &http.Client{Transport: http.DefaultTransport.(*http.Transport).Clone(), Timeout: 30 * time.Second}

	up := newUploadClient(api)
	if up.Timeout != 0 {
		t.Errorf("upload client must not bound the whole exchange, Timeout = %v", up.Timeout)
	}
	transport, ok := up.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport = %T, want *http.Transport clone", up.Transport)
	}
	if transport.ResponseHeaderTimeout < time.Minute {
		t.Errorf("ResponseHeaderTimeout = %v, want at least 1m", transport.ResponseHeaderTimeout)
	}
	if transport == api.Transport {
		t.Error("the API transport must be cloned, not mutated")
	}
}

func TestNewUploadClient_CustomTransportKeepsClient(t *testing.T) {
	api := &http.Client{Transport: http.NewFileTransport(http.Dir(t.TempDir()))}
	if got := newUploadClient(api); got != api {
		t.Error("a non-*http.Transport client must be returned unchanged")
	}
}
