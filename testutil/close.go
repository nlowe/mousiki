package testutil

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertCloses(t *testing.T, c io.Closer) func() {
	return func() {
		assert.NoError(t, c.Close())
	}
}
