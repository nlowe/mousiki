package audio

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/sirupsen/logrus"
	"go.uber.org/multierr"
)

const (
	targetSampleRate beep.SampleRate = 41000
	resampleQuality                  = 3
)

var ffmpegArgs = []string{
	"-y",                                 // Yes to All
	"-hide_banner", "-loglevel", "panic", // Be Quiet
	"-i", "pipe:0", // Input from stdin
	"-c:a", "pcm_s16le", // PCM Signed 16-bit Little Endian output
	"-f", "wav", // Output WAV
}

type beepFFmpegPlayer struct {
	ffmpeg string

	transcodedTrack     *os.File
	nowStreaming        beep.StreamSeekCloser
	streamingSampleRate beep.SampleRate
	ctrl                *beep.Ctrl

	progressTicker *time.Ticker
	progress       chan PlaybackProgress
	done           chan error

	log logrus.FieldLogger
}

// NewBeepFFmpegPipeline returns an audio.Player that transcodes tracks through FFmpeg
// via exec.Command to PCM and then plays audio via speaker.Play. Tracks must be fully
// transcoded first otherwise wav.Decode will refuse to play them. Because of this,
// UpdateStream will block until transcoding is complete.
func NewBeepFFmpegPipeline() (*beepFFmpegPlayer, error) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("could not locate ffmpeg on $PATH: %w", err)
	}

	if err := speaker.Init(targetSampleRate, targetSampleRate.N(100*time.Millisecond)); err != nil {
		return nil, fmt.Errorf("failed to init beep speaker: %w", err)
	}

	result := &beepFFmpegPlayer{
		ffmpeg: ffmpeg,

		ctrl: &beep.Ctrl{Paused: true},

		progressTicker: time.NewTicker(1 * time.Second),
		progress:       make(chan PlaybackProgress, 1),
		done:           make(chan error, 1),

		log: logrus.WithField("prefix", "ffmpeg"),
	}

	go func() {
		for range result.progressTicker.C {
			if result.nowStreaming != nil {
				result.progress <- result.calculateProgress()
			}
		}
	}()

	return result, nil
}

func (b *beepFFmpegPlayer) cleanup() (err error) {
	if b.transcodedTrack != nil {
		err = multierr.Combine(
			b.nowStreaming.Close(),
			b.transcodedTrack.Close(),
			os.Remove(b.transcodedTrack.Name()),
		)

		b.nowStreaming = nil
		b.transcodedTrack = nil
	}

	return err
}

func (b *beepFFmpegPlayer) Close() error {
	speaker.Lock()
	defer speaker.Unlock()

	speaker.Close()
	b.progressTicker.Stop()
	return b.cleanup()
}

func (b *beepFFmpegPlayer) UpdateStream(url string, volumeAdjustment float64) {
	// Stop playing anything currently playing
	speaker.Clear()
	b.ctrl.Paused = true

	// Clean up if we were previously playing something
	_ = b.cleanup()

	// Transcode to WAV
	var err error
	b.transcodedTrack, err = b.transcode(url)
	if err != nil {
		b.log.WithError(err).Errorf("Transcoding failed")
		b.done <- err
		return
	}

	// Decode
	var format beep.Format
	b.nowStreaming, format, err = wav.Decode(b.transcodedTrack)
	if err != nil {
		b.log.WithError(err).Errorf("Could not decode transcoded file")
		b.done <- err
		return
	}

	b.log.WithFields(logrus.Fields{
		"sampleRate": format.SampleRate,
		"channels":   format.NumChannels,
		"replayGain": volumeAdjustment,
	}).Debug("Decoded track")

	// Setup pipeline
	b.streamingSampleRate = format.SampleRate
	b.ctrl.Streamer = beep.Resample(resampleQuality, b.streamingSampleRate, targetSampleRate, &effects.Volume{
		Base:     10,
		Volume:   volumeAdjustment / 10,
		Streamer: b.nowStreaming,
	})

	// Reset progress
	b.progressTicker.Reset(1 * time.Second)
	b.progress <- b.calculateProgress()

	// Play!
	speaker.Play(beep.Seq(b.ctrl, beep.Callback(func() {
		b.done <- nil
	})))

	b.ctrl.Paused = false
}

func (b *beepFFmpegPlayer) Play() {
	b.log.WithFields(logrus.Fields{}).Trace("Asked to play")

	speaker.Lock()
	defer speaker.Unlock()

	b.ctrl.Paused = false
}

func (b *beepFFmpegPlayer) Pause() {
	b.log.WithFields(logrus.Fields{}).Trace("Asked to pause")

	speaker.Lock()
	defer speaker.Unlock()

	b.ctrl.Paused = true
}

func (b *beepFFmpegPlayer) IsPlaying() bool {
	speaker.Lock()
	defer speaker.Unlock()

	v := !b.ctrl.Paused
	return v
}

func (b *beepFFmpegPlayer) ProgressChan() <-chan PlaybackProgress {
	return b.progress
}

func (b *beepFFmpegPlayer) DoneChan() <-chan error {
	return b.done
}

func (b *beepFFmpegPlayer) transcode(url string) (*os.File, error) {
	b.log.WithField("track", url).Debug("Attempting to transcode track")

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("transcode: failed to fetch track: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	tmp, err := ioutil.TempFile(os.TempDir(), "mousiki")
	if err != nil {
		return nil, fmt.Errorf("transcode: failed to create temp file: %w", err)
	}
	_ = tmp.Close()

	b.log.WithField("file", tmp.Name()).Debug("Transcoding Track")

	cmd := exec.Command(b.ffmpeg, append(ffmpegArgs, tmp.Name())...)

	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("transcode: ffmpeg: failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("transcode: ffmpeg: transcoding failed")
	}

	n, err := io.Copy(stdin, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("transcode: ffmpeg: failed to transcode track: %w", err)
	}

	b.log.WithFields(logrus.Fields{
		"file": tmp.Name(),
		"len":  n,
	}).Debug("Transcoding complete")

	if err := stdin.Close(); err != nil {
		return nil, fmt.Errorf("transcode: ffmpeg: failed to close stdin: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("transcode: ffmpeg: unknown transcoding error: %w", err)
	}

	return os.Open(tmp.Name())
}

func (b *beepFFmpegPlayer) calculateProgress() PlaybackProgress {
	if b.nowStreaming == nil {
		return PlaybackProgress{}
	}

	return PlaybackProgress{
		Duration: b.streamingSampleRate.D(b.nowStreaming.Len()),
		Progress: b.streamingSampleRate.D(b.nowStreaming.Position()),
	}
}
