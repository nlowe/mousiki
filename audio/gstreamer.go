package audio

import (
	"fmt"
	"strings"
	"time"

	"github.com/notedit/gst"
	"github.com/sirupsen/logrus"
)

type gstreamerPlayer struct {
	pipeline *gst.Pipeline
	src      *gst.Element
	volume   *gst.Element

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

	convert, err := gst.ElementFactoryMake("audioconvert", "convert")
	if err != nil {
		return nil, err
	}

	result.volume, err = gst.ElementFactoryMake("volume", "volume")
	if err != nil {
		return nil, err
	}

	resample, err := gst.ElementFactoryMake("audioresample", "resample")
	if err != nil {
		return nil, err
	}

	sink, err := gst.ElementFactoryMake("autoaudiosink", "sink")
	if err != nil {
		return nil, err
	}

	result.pipeline.AddMany(convert, result.volume, resample, sink)
	convert.Link(result.volume)
	result.volume.Link(resample)
	resample.Link(sink)

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
			sinkpad := g.pipeline.GetByName("convert").GetStaticPad("sink")
			p.Link(sinkpad)
		}
	})

	g.pipeline.Add(g.src)

	// TODO: Stop this goroutine when we close the player
	go func() {
		for range time.Tick(time.Second) {
			var result PlaybackProgress

			result.Progress, err = g.src.QueryPosition()
			if err != nil {
				g.log.WithError(err).Error("Failed to provide playback progress")
				continue
			}

			result.Duration, err = g.src.QueryDuration()
			if err != nil {
				g.log.WithError(err).Error("Failed to provide playback progress")
				continue
			}

			select {
			case g.progress <- result:
			default:
				g.log.Warn("progress channel blocked upstream")
			}
		}
	}()

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
	percentGain := RelativeDBToPercent(volumeAdjustment)
	g.log.WithFields(logrus.Fields{
		"url":           url,
		"gain":          volumeAdjustment,
		"volumePercent": percentGain * 100,
	}).Debug("Updating Stream")
	g.playing = false
	g.pipeline.SetState(gst.StateNull)

	g.volume.SetObject("volume", percentGain)

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
