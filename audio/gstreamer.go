package audio

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/notedit/gst"
	"github.com/sirupsen/logrus"
)

var playbackRegex = regexp.MustCompile(`.*current=\(gint64\)(\d+), total=\(gint64\)(\d+),.*`)

type gstreamerPlayer struct {
	pipeline *gst.Pipeline
	src      *gst.Element

	playing  bool
	progress chan PlaybackProgress
	done     chan error

	log logrus.FieldLogger
}

func NewGstreamerPipeline() (*gstreamerPlayer, error) {
	result := &gstreamerPlayer{
		progress: make(chan PlaybackProgress, 1),
		done:     make(chan error, 1),
		log:      logrus.WithField("prefix", "gstreamer"),
	}

	var err error
	result.pipeline, err = gst.PipelineNew("mousiki")
	if err != nil {
		return nil, err
	}

	progress, err := gst.ElementFactoryMake("progressreport", "progress")
	if err != nil {
		return nil, err
	}

	progress.SetObject("update-freq", 1)
	progress.SetObject("silent", true)

	convert, err := gst.ElementFactoryMake("audioconvert", "convert")
	if err != nil {
		return nil, err
	}

	sink, err := gst.ElementFactoryMake("autoaudiosink", "sink")
	if err != nil {
		return nil, err
	}

	result.pipeline.AddMany(progress, convert, sink)
	progress.Link(convert)
	convert.Link(sink)

	// TODO: Stop this goroutine when we close the player
	go func() {
		bus := result.pipeline.GetBus()
		for {
			msg := bus.Pull(gst.MessageEos | gst.MessageError | gst.MessageElement)

			switch msg.GetType() {
			case gst.MessageEos:
				result.done <- nil
			case gst.MessageError:
				result.done <- fmt.Errorf("error during playback: %s: %s", msg.GetName(), msg.GetStructure().ToString())
			case gst.MessageElement:
				data := msg.GetStructure().ToString()
				result.log.WithField("msg", data).Debug("Got message")

				if strings.HasPrefix(data, "progress") {
					hackyParser := playbackRegex.FindStringSubmatch(data)
					current, _ := strconv.Atoi(hackyParser[1])
					duration, _ := strconv.Atoi(hackyParser[2])

					// TODO: Is this always in seconds?
					result.progress <- PlaybackProgress{
						Progress: time.Duration(current) * time.Second,
						Duration: time.Duration(duration) * time.Second,
					}
				}
			}
		}
	}()

	return result, nil
}

func (g *gstreamerPlayer) setupInitialSource(url string) error {
	var err error
	g.src, err = gst.ElementFactoryMake("uridecodebin", "source")
	if err != nil {
		return err
	}
	g.src.SetObject("uri", url)

	g.src.SetPadAddedCallback(func(e *gst.Element, p *gst.Pad) {
		g.log.Debug("Source Pad Callback Called")
		if strings.HasPrefix(p.GetCurrentCaps().ToString(), "audio") {
			g.log.Debug("Source Audio Pad added")
			sinkpad := g.pipeline.GetByName("progress").GetStaticPad("sink")
			p.Link(sinkpad)
		}
	})

	g.pipeline.Add(g.src)

	return nil
}

func (g *gstreamerPlayer) Close() error {
	// TODO: Do we have to do anything else to close the pipeline properly?
	g.pipeline.SetState(gst.StateNull)
	g.playing = false
	close(g.done)
	close(g.progress)
	return nil
}

func (g *gstreamerPlayer) UpdateStream(url string, volumeAdjustment float64) {
	g.log.WithField("url", url).Info("Updating Stream")
	g.playing = false
	g.pipeline.SetState(gst.StateNull)
	if g.src == nil {
		if err := g.setupInitialSource(url); err != nil {
			g.done <- err
			close(g.done)
		}
	} else {
		g.src.SetObject("uri", url)
	}
	g.Play()
}

func (g *gstreamerPlayer) Play() {
	g.log.Info("Resuming playback")
	g.pipeline.SetState(gst.StatePlaying)
	g.playing = true
}

func (g *gstreamerPlayer) Pause() {
	g.log.Info("Pausing Playback")
	g.pipeline.SetState(gst.StatePaused)
	g.playing = false
}

func (g *gstreamerPlayer) IsPlaying() bool {
	return g.playing
}

func (g *gstreamerPlayer) ProgressChan() <-chan PlaybackProgress {
	return g.progress
}

func (g *gstreamerPlayer) DoneChan() <-chan error {
	return g.done
}
