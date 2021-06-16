package cmd

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/nlowe/mousiki/pandora"
	"github.com/nlowe/mousiki/pandora/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

var dumpFeedbackCmd = &cobra.Command{
	Use:   "dump-feedback",
	Short: "dump all feedback from pandora",
	Long:  "Dumps all feedback from all known stations",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		logrus.SetOutput(colorable.NewColorableStderr())
		p := api.NewClient()

		un := viper.GetString("username")
		pw := viper.GetString("password")

		if pw == "" {
			_, _ = fmt.Fprint(os.Stderr, "Password: ")
			raw, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
			pw = string(raw)

			if len(pw) < 8 {
				return fmt.Errorf("got bad password: %s (hex: %s)", pw, hex.EncodeToString(raw))
			}

			_, _ = fmt.Fprintln(os.Stderr)
		}

		if pw == "" {
			logrus.Fatal("No password provided")
		}

		if err := p.LegacyLogin(un, pw); err != nil {
			return err
		}

		w := csv.NewWriter(os.Stdout)
		defer w.Flush()

		if err := w.Write(pandora.FeedbackCSVHeaders); err != nil {
			return err
		}

		logrus.Info("Fetching Feedback...")
		i := 1
		for f := range p.ListFeedback() {
			if err := w.Write(f.MarshalCSV()); err != nil {
				return err
			}

			i++
			if i%10 == 0 {
				w.Flush()
			}
		}

		logrus.Infof("Fetched %d feedback entries", i-1)
		return nil
	},
}