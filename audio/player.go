package audio

import "io"

// Player represents an audio playback engine that can play arbitrary audio URLs
type Player interface {
	io.Closer

	// UpdateStream sets the target of the playback stream. If the stream
	// is playing, it is automatically restarted with the new media source
	UpdateStream(url string, volumeAdjustment float64)
	// Play starts the playback stream
	Play()
	// Pause pauses the playback stream
	Pause()

	// IsPlaying is true if the player is currently playing a track
	IsPlaying() bool

	// ProgressChan reports playback progress on at least a 1hz interval
	ProgressChan() <-chan PlaybackProgress

	// DoneChan reports when the player has finished playback of the current
	// stream target by returning a nil error. If the player encounters a
	// recoverable error, it will send a non-nil error on this channel. If it
	// encounters an unrecoverable error, it will send the error and then close
	// this channel
	DoneChan() <-chan error
}
