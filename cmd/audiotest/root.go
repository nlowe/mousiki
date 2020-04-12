package audiotest

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "audiotest",
	Short: "Test audio playback",
	Long:  "Commands for testing auido playback engines",
	Args:  cobra.NoArgs,
}
