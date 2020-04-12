package audio

import (
	"fmt"
	"time"
)

type PlaybackProgress struct {
	Progress time.Duration
	Duration time.Duration
}

func (p PlaybackProgress) String() string {
	return fmt.Sprintf("%s/%s", formatDuration(p.Progress), formatDuration(p.Duration))
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	return fmt.Sprintf("%d:%02d", m, s)
}
