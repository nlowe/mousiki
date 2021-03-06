package audiotest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/andrewstuart/limio"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "Stream a test file",
	Long:    "Stream a file to test audio playback",
	Example: "mousiki audiotest server test.aac",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		throttleLimit := viper.GetInt("kbps")
		log := logrus.WithFields(logrus.Fields{
			"file":   args[0],
			"prefix": "server",
		})
		info, stat := os.Stat(args[0])

		if os.IsNotExist(stat) {
			log.Fatal("Could not locate file to stream")
		}

		f, err := os.Open(args[0])
		if err != nil {
			return err
		}

		stream, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		s := &http.Server{
			Addr: "0.0.0.0:5000",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Info("Serving Client")

				w.Header().Set("Content-Type", viper.GetString("content-type"))
				w.Header().Set("Content-Length", strconv.Itoa(int(info.Size())))
				w.WriteHeader(http.StatusOK)

				if r.Method != http.MethodHead {
					throttle := limio.NewReader(bytes.NewReader(stream))
					throttle.SimpleLimit(throttleLimit*limio.KB, time.Second)

					_, _ = io.Copy(w, throttle)
					log.Info("Streaming Complete")
				}
			}),
		}

		go func() {
			log.WithFields(logrus.Fields{
				"endpoint": "http://localhost:5000/stream",
				"throttle": fmt.Sprintf("%dKB/s", throttleLimit),
			}).Info("Streaming")
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.WithError(err).Fatal("Error serving requests")
			}
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		<-c

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Info("Shutting Down")
		return s.Shutdown(ctx)
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)

	flags := serverCmd.PersistentFlags()
	flags.String("content-type", "audio/mp4", "The content type to set when streaming")
	flags.Int("kbps", 1024, "Throttle downloads to this rate, in KB/s (default: 1MB/s or 1024KB/s)")
	_ = viper.BindPFlags(flags)
}
