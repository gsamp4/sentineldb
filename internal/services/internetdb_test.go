package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCallInternetDBUsesInternetDB(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/8.8.8.8" {
			t.Fatalf("expected path /8.8.8.8, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ip":"8.8.8.8","ports":[53,443],"hostnames":["dns.google"],"tags":["resolver"],"vulns":[],"cpes":[]}`))
	}))
	defer server.Close()

	originalClient := internetDBHTTPClient
	originalEndpoint := internetDBEndpoint
	internetDBHTTPClient = server.Client()
	internetDBEndpoint = server.URL
	t.Cleanup(func() {
		internetDBHTTPClient = originalClient
		internetDBEndpoint = originalEndpoint
	})

	result, err := CallInternetDB(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(result.Matches))
	}
	if result.Matches[0].IP != "8.8.8.8" || result.Matches[0].Port != 53 {
		t.Fatalf("unexpected first match: %+v", result.Matches[0])
	}
	if result.Matches[1].Port != 443 {
		t.Fatalf("unexpected second match: %+v", result.Matches[1])
	}
	if len(result.Matches[0].Hostnames) != 1 || result.Matches[0].Hostnames[0] != "dns.google" {
		t.Fatalf("unexpected hostnames: %+v", result.Matches[0].Hostnames)
	}
}

func TestCallInternetDBRejectsNonIPAsset(t *testing.T) {
	_, err := CallInternetDB(context.Background(), "example.com")
	if err == nil {
		t.Fatal("expected error for non-IP asset")
	}
}