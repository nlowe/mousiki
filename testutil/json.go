package testutil

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func UnmarshalRequest(t *testing.T, r *http.Request, v interface{}) {
	defer AssertCloses(t, r.Body)
	assert.NoError(t, json.NewDecoder(r.Body).Decode(v))
}

func MarshalResponse(t *testing.T, status int, w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	assert.NoError(t, json.NewEncoder(w).Encode(v))
}
