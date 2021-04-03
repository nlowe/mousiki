package cmd

import (
	"fmt"
	"os"

	"github.com/nlowe/mousiki/pandora/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

var tokenTestCmd = &cobra.Command{
	Use:    "tokentest",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		p := api.NewClient()

		un := viper.GetString("username")
		pw := viper.GetString("password")

		if pw == "" {
			fmt.Print("Password: ")
			raw, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
			pw = string(raw)
			fmt.Println()
		}

		if pw == "" {
			logrus.Fatal("No password provided")
		}

		if err := p.LegacyLogin(un, pw); err != nil {
			return err
		}

		stations, err := p.GetStations()
		if err != nil {
			return err
		}

		for _, station := range stations {
			logrus.Infof("Found Station: %s", station)
		}

		return nil
	},
}
