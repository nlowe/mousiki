package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/nlowe/mousiki/pandora"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"golang.org/x/crypto/ssh/terminal"
)

var rootCmd = &cobra.Command{
	Use:   "mousiki",
	Short: "A command-line pandora client",
	Long:  "A command-line pandora client based off of pianobar",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		p := pandora.NewClient()

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

		if err := p.Login(un, pw); err != nil {
			return err
		}

		return nil
	},
}

func MarkFlagRequired(cmd *cobra.Command, name string) {
	_ = cmd.MarkFlagRequired(name)
}

func init() {
	logrus.SetOutput(colorable.NewColorableStdout())
	logrus.SetFormatter(&prefixed.TextFormatter{
		ForceColors:     true,
		ForceFormatting: true,
		FullTimestamp:   true,
	})

	viper.SetConfigName("mousiki")
	viper.SetEnvPrefix("mousiki")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	cobra.OnInitialize(func() {
		verbosity, err := logrus.ParseLevel(viper.GetString("verbosity"))
		if err != nil {
			logrus.WithError(err).WithField("verbosity", verbosity).Fatal("Failed to set verbosity")
		}

		logrus.SetLevel(verbosity)
	})

	flags := rootCmd.PersistentFlags()

	flags.StringP("username", "u", "", "Pandora Username")
	MarkFlagRequired(rootCmd, "username")
	flags.StringP("password", "p", "", "Pandora Password")

	flags.StringP("verbosity", "v", "info", "Verbosity []")

	_ = viper.BindPFlags(flags)
}

func Exec() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
