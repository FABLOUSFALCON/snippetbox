package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FABLOUSFALCON/snippetbox/internal/assert"
)

func TestPing(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	ping(rr, r)

	rs := rr.Result()

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
