package domain

import (
	"testing"
)

func TestValidateAsset(t *testing.T) {
    tests := []struct {
        name      string
        assetType string
        value     string
        wantErr   bool
    }{
        {"valid ip",   "ip", "192.168.15.1",    false},
        {"invalid ip", "ip", "999.999.999.999", true},
        {"ip as text", "ip", "banana",          true},

        {"valid email",          "email", "user@example.com", false},
        {"email without @",      "email", "userexample.com",  true},
        {"email without domain", "email", "user@",            true},

        {"valid domain",   "domain", "example.com",  false},
        {"invalid domain", "domain", "ex ample.com", true},

        {"unknown type", "url",  "https://example.com", true},
        {"empty type",   "",     "192.168.15.1",        true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateAsset(tt.assetType, tt.value)

            if tt.wantErr && err == nil {
                t.Errorf("expected error, got nil")
            }
            if !tt.wantErr && err != nil {
                t.Errorf("expected no error, got %v", err)
            }
        })
    }
}

