package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FABLOUSFALCON/snippetbox/internal/assert"
)

func TestCommonHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("OK")); err != nil {
			return
		}
	})

	commonHeaders(next).ServeHTTP(rr, r)

	rs := rr.Result()

	expectedValue := "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com"
	assert.Equal(t, rs.Header.Get("Content-Security-Policy"), expectedValue)

	expectedValue = "origin-when-cross-origin"
	assert.Equal(t, rs.Header.Get("Referrer-Policy"), expectedValue)
	// Check that the middleware has correctly set the X-Content-Type-Options
	// header on the response.
	expectedValue = "nosniff"
	assert.Equal(t, rs.Header.Get("X-Content-Type-Options"), expectedValue)
	// Check that the middleware has correctly set the X-Frame-Options header
	// on the response.
	expectedValue = "deny"
	assert.Equal(t, rs.Header.Get("X-Frame-Options"), expectedValue)
	// Check that the middleware has correctly set the X-XSS-Protection header
	// on the response
	expectedValue = "0"
	assert.Equal(t, rs.Header.Get("X-XSS-Protection"), expectedValue)
	// Check that the middleware has correctly set the Server header on the
	// response.
	expectedValue = "Go"
	assert.Equal(t, rs.Header.Get("Server"), expectedValue)
	// Check that the middleware has correctly called the next handler in line
	// and the response status code and body are as expected.
	assert.Equal(t, rs.StatusCode, http.StatusOK)

	defer func() {
		if err := rs.Body.Close(); err != nil {
			return
		}
	}()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	body = bytes.TrimSpace(body)

	assert.Equal(t, string(body), "OK")
}
