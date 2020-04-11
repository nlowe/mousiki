package testutil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertCookie(t *testing.T, r *http.Request, name, value string) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == name {
			assert.Equal(t, value, cookie.Value)
			return
		}
	}

	t.Errorf("Cookie %s=%s not found in request", name, value)
}
