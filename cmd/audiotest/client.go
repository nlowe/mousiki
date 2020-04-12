package audiotest

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	// TODO: Upstream our fixes to this package
	// TODO: Write our own AAC Decoder that works with github.com/faiface/beep to drop the cgo dependency?
	"github.com/notedit/gst"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var playbackRegex = regexp.MustCompile(`.*current=\(gint64\)(\d+), total=\(gint64\)(\d+),.*`)

var clientCmd = &cobra.Command{
	Use:     "client",
	Short:   "Test streaming music over HTTP",
	Long:    "Play a track over HTTP. Use with 'mousiki audiotest server'.",
	Example: "mousiki audiotest client http://localhost:5000/stream",
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		pipeline, err := gst.PipelineNew("mousiki")
		if err != nil {
			return err
		}

		src, err := gst.ElementFactoryMake("uridecodebin", "source")
		if err != nil {
			return err
		}
		src.SetObject("uri", args[0])

		progress, err := gst.ElementFactoryMake("progressreport", "progress")
		if err != nil {
			return err
		}

		progress.SetObject("update-freq", 1)
		progress.SetObject("silent", true)

		convert, err := gst.ElementFactoryMake("audioconvert", "convert")
		if err != nil {
			return err
		}

		sink, err := gst.ElementFactoryMake("autoaudiosink", "sink")
		if err != nil {
			return err
		}

		pipeline.AddMany(src, progress, convert, sink)
		progress.Link(convert)
		convert.Link(sink)

		src.SetPadAddedCallback(func(e *gst.Element, p *gst.Pad) {
			if strings.HasPrefix(p.GetCurrentCaps().ToString(), "audio") {
				logrus.Info("Source Audio Pad added")
				sinkpad := progress.GetStaticPad("sink")
				p.Link(sinkpad)
			}
		})

		pipeline.SetState(gst.StatePlaying)
		defer pipeline.SetState(gst.StateNull)

		bus := pipeline.GetBus()

	playback:
		for {
			message := bus.Pull(gst.MessageAny)
			switch message.GetType() {
			case gst.MessageEos:
				break playback
			case gst.MessageError:
				return fmt.Errorf("error in gst pipeline: %s", message.GetName())
			case gst.MessageDurationChanged:
				logrus.WithField("data", message.GetStructure().ToString()).Info("Duration Changed")
			case gst.MessageElement:
				data := message.GetStructure().ToString()

				if strings.HasPrefix(data, "progress") {
					hackyParser := playbackRegex.FindStringSubmatch(message.GetStructure().ToString())
					current, _ := strconv.Atoi(hackyParser[1])
					duration, _ := strconv.Atoi(hackyParser[2])

					logrus.WithField(
						"progress",
						fmt.Sprintf("%s/%s", time.Duration(current)*time.Second, time.Duration(duration)*time.Second),
					).Info("Made Progress!")
				}
			}
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(clientCmd)
}
