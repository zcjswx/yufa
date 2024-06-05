package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_login(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/sign_in" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><head><meta name="csrf-token" content="test-csrf-token"></head></html>`))
		}
	}))
	defer server.Close()

	baseURI = server.URL // Redirect to our mock server
	client := server.Client()

	err := login(client)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
}
