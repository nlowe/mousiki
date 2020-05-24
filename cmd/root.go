package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/mattn/go-colorable"
	"github.com/nlowe/mousiki/audio"
	"github.com/nlowe/mousiki/cmd/audiotest"
	"github.com/nlowe/mousiki/mousiki"
	"github.com/nlowe/mousiki/mousiki/ui"
	"github.com/nlowe/mousiki/pandora"
	"github.com/nlowe/mousiki/pandora/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"golang.org/x/crypto/ssh/terminal"
)

const trackTemplate pb.ProgressBarTemplate = `{{string . "track"}} {{bar . "[" "#" ">" "_" "]"}} {{string . "suffix"}}`

var RootCmd = &cobra.Command{
	Use:   "mousiki",
	Short: "A command-line pandora client",
	Long:  "A command-line pandora client based off of pianobar",
	Args:  cobra.NoArgs,
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

		if err := p.Login(un, pw); err != nil {
			return err
		}

		player, err := audio.NewGstreamerPipeline()
		if err != nil {
			return err
		}

		defer func() {
			_ = player.Close()
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		controller := mousiki.NewStationController(p, player)

		app := ui.New(ctx, cancel, player, controller)
		return app.Run()
	},
}

func MarkFlagRequired(cmd *cobra.Command, name string) {
	_ = cmd.MarkFlagRequired(name)
}

func init() {
	RootCmd.AddCommand(audiotest.RootCmd)

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

	flags := RootCmd.PersistentFlags()

	flags.StringP("username", "u", "", "Pandora Username")
	MarkFlagRequired(RootCmd, "username")
	flags.StringP("password", "p", "", "Pandora Password")

	flags.StringP("audio-format", "a", string(pandora.AudioFormatAACPlus), "Audio Format to use [aacplus, mp3]")

	flags.StringP("verbosity", "v", "info", "Verbosity []")

	_ = viper.BindPFlags(flags)
}

func Exec() {
	if err := RootCmd.Execute(); err != nil {
		panic(err)
	}
}
