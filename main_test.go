package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersion(t *testing.T) {
	w := httptest.NewRecorder()
	version(w, nil)

	resp := w.Result()
	if have, want := resp.StatusCode, http.StatusOK; have != want {
		t.Errorf("Status code is wrong. Have: %d, want: %d.", have, want)
	}
}
