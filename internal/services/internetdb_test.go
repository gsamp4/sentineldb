package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInternetDBLookupByIP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/8.8.8.8" {
			t.Fatalf("expected /8.8.8.8, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ip":"8.8.8.8","ports":[53,443],"hostnames":["dns.google"],"tags":["resolver"],"vulns":[],"cpes":[]}`))
	}))
	defer server.Close()

	service := NewInternetDBService(server.URL, server.Client())
	result, err := service.LookupByIP(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.IP != "8.8.8.8" {
		t.Fatalf("expected IP 8.8.8.8, got %s", result.IP)
	}
	if len(result.Ports) != 2 || result.Ports[0] != 53 || result.Ports[1] != 443 {
		t.Fatalf("unexpected ports: %+v", result.Ports)
	}
	if len(result.Hostnames) != 1 || result.Hostnames[0] != "dns.google" {
		t.Fatalf("unexpected hostnames: %+v", result.Hostnames)
	}
}

func TestInternetDBLookupByIPRejectsInvalidIP(t *testing.T) {
	service := NewInternetDBService("", nil)

	_, err := service.LookupByIP(context.Background(), "example.com")
	if err == nil {
		t.Fatal("expected error for invalid IP")
	}
}