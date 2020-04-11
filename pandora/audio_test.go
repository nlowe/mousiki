package pandora

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidAudioFormat(t *testing.T) {
	valid := func(f AudioFormat) func(*testing.T) {
		return func(t *testing.T) {
			require.True(t, IsValidAudioFormat(string(f)))
		}
	}

	t.Run(string(AudioFormatAACPlus), valid(AudioFormatAACPlus))
	t.Run(string(AudioFormatMP3), valid(AudioFormatMP3))
	require.False(t, IsValidAudioFormat("foo"))
}
