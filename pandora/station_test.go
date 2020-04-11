package pandora

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStation_String(t *testing.T) {
	sut := &Station{
		ID:   "DummyID",
		Name: "Test Station",
	}

	require.Equal(t, "[DummyID] Test Station", sut.String())
}
