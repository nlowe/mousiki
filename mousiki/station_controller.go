package mousiki

import (
	"context"

	"github.com/nlowe/mousiki/audio"
	"github.com/nlowe/mousiki/pandora"
	"github.com/nlowe/mousiki/pandora/api"
	"github.com/sirupsen/logrus"
)

type StationController struct {
	station pandora.Station
	pandora api.Client
	player  audio.Player

	playing pandora.Track
	queue   []pandora.Track

	skip          chan struct{}
	notifications chan pandora.Track

	log logrus.FieldLogger
}

func NewStationController(s pandora.Station, c api.Client, p audio.Player) *StationController {
	return &StationController{
		station: s,
		pandora: c,
		player:  p,

		skip:          make(chan struct{}, 1),
		notifications: make(chan pandora.Track, 1),

		log: logrus.WithFields(logrus.Fields{
			"prefix":  "stationController",
			"station": s.String(),
		}),
	}
}

func (s *StationController) Play(ctx context.Context) {
	for {
		// TODO: Configure prefetch limit?
		if len(s.queue) == 0 {
			s.log.Info("Fetching more tracks")
			tracks, err := s.pandora.GetMoreTracks(s.station.ID)
			if err != nil {
				// TODO: More graceful error handling
				s.log.WithError(err).Fatal("Failed to fetch more tracks")
			}

			// TODO: Notify when we fetch new tracks?
			for _, t := range tracks {
				s.log.WithField("track", t).Info("Up Next")
			}
			s.queue = append(s.queue, tracks...)
		}

		s.playing, s.queue = s.queue[0], s.queue[1:]

		s.log.WithField("track", s.playing.String()).Info("Playing new track")
		select {
		case s.notifications <- s.playing:
		}
		s.player.UpdateStream(s.playing.AudioUrl, s.playing.FileGain)

		select {
		case <-s.skip:
			s.log.Info("Skipping to next track")
		case err := <-s.player.DoneChan():
			if err != nil {
				// TODO: Bubble up error?
				s.log.WithError(err).Error("Error during playback")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *StationController) Skip() {
	s.player.Pause()
	s.skip <- struct{}{}
}

func (s *StationController) NowPlaying() pandora.Track {
	return s.playing
}

func (s *StationController) UpNext() []pandora.Track {
	result := make([]pandora.Track, len(s.queue))
	copy(result, s.queue)

	return result
}

func (s *StationController) NotificationChan() <-chan pandora.Track {
	return s.notifications
}
