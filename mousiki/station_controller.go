package mousiki

import (
	"context"
	"sync"

	"github.com/nlowe/mousiki/audio"
	"github.com/nlowe/mousiki/pandora"
	"github.com/nlowe/mousiki/pandora/api"
	"github.com/sirupsen/logrus"
)

const NoStationSelected = "__mousiki_no_station"

var noStationSelected = pandora.Station{
	ID:   NoStationSelected,
	Name: "No Station Selected",
}

type narrativeCache struct {
	station   string
	track     string
	narrative pandora.Narrative
}

type StationController struct {
	stationLock sync.Mutex
	station     pandora.Station
	pandora     api.Client
	player      audio.Player

	playing *pandora.Track
	queue   []pandora.Track

	skip           chan struct{}
	notifications  chan *pandora.Track
	stationChanged chan pandora.Station

	narrativeCache narrativeCache

	log logrus.FieldLogger
}

func (n narrativeCache) matches(t *pandora.Track) bool {
	return n.station == t.StationId && n.track == t.MusicId
}

func NewStationController(c api.Client, p audio.Player) *StationController {
	return &StationController{
		pandora: c,
		player:  p,
		station: noStationSelected,

		notifications:  make(chan *pandora.Track, 1),
		stationChanged: make(chan pandora.Station, 1),

		log: logrus.WithFields(logrus.Fields{
			"prefix":  "stationController",
			"station": noStationSelected,
		}),
	}
}

func (s *StationController) Play(ctx context.Context) {
	if s.station.ID == NoStationSelected {
		s.log.Error("No Station Selected, nothing to play")
		return
	}

	s.skip = make(chan struct{}, 1)

	for {
		// TODO: Configure prefetch limit?
		s.stationLock.Lock()
		if len(s.queue) <= 1 {
			s.log.Info("Fetching more tracks")
			tracks, err := s.pandora.GetMoreTracks(s.station.ID)
			if err != nil {
				// TODO: More graceful error handling
				s.log.WithError(err).Fatal("Failed to fetch more tracks")
			}

			s.queue = append(s.queue, tracks...)
		}

		s.playing, s.queue = &s.queue[0], s.queue[1:]

		s.log.WithField("track", s.playing.String()).Info("Playing new track")
		select {
		case s.notifications <- s.playing:
		}
		s.player.UpdateStream(s.playing.AudioUrl, s.playing.FileGain)
		s.stationLock.Unlock()

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
	if s.skip != nil {
		s.player.Pause()
		s.skip <- struct{}{}
	}
}

func (s *StationController) NowPlaying() *pandora.Track {
	return s.playing
}

// TODO: There are endpoints listed for removing feedback, but they're not documented
func (s *StationController) ProvideFeedback(f pandora.TrackRating) error {
	s.stationLock.Lock()
	defer s.stationLock.Unlock()

	log := s.log.WithField("track", s.playing)

	if s.playing.Rating == f {
		log.Warn("Not adding duplicate feedback")
		return nil
	}

	if f == pandora.TrackRatingTired {
		log.Info("Temporarily timing-out song")
		err := s.pandora.AddTired(s.playing.TrackToken)

		if err == nil {
			// TODO: The UI does not currently differentiate between banned and tired songs
			s.playing.Rating = pandora.TrackRatingBan
			select {
			case s.skip <- struct{}{}:
			}
		}

		return err
	} else {
		positive := true
		if f == pandora.TrackRatingBan {
			log.Info("Banning song")
			positive = false
		} else {
			log.Info("Loving song")
		}

		err := s.pandora.AddFeedback(s.playing.TrackToken, positive)
		if err == nil {
			s.playing.Rating = f
			if !positive {
				select {
				case s.skip <- struct{}{}:
				}
			}
		}

		return err
	}
}

func (s *StationController) UpNext() []pandora.Track {
	result := make([]pandora.Track, len(s.queue))
	copy(result, s.queue)

	return result
}

func (s *StationController) ListStations() ([]pandora.Station, error) {
	return s.pandora.GetStations()
}

func (s *StationController) SwitchStations(station pandora.Station) {
	s.stationLock.Lock()
	defer s.stationLock.Unlock()

	if s.station.ID == station.ID {
		s.log.Info("Requested station is already playing")
		return
	}

	s.log.WithField("newStation", station).Info("Switching Stations")

	// Change the station and clear the queue to force the next control loop
	// to fetch tracks from the new station
	s.station = station
	s.queue = []pandora.Track{}

	// Try to skip immediately in case we're currently playing a track
	s.Skip()

	// Reset the logger to pick up the new station name
	s.log = logrus.WithFields(logrus.Fields{
		"prefix":  "stationController",
		"station": station,
	})

	s.stationChanged <- station
}

func (s *StationController) ExplainCurrentTrack() (pandora.Narrative, error) {
	if s.narrativeCache.matches(s.playing) {
		s.log.Debug("Returning Cached Narrative")
		return s.narrativeCache.narrative, nil
	}

	s.log.Debug("Fetching Narrative")
	result, err := s.pandora.GetNarrative(s.playing.StationId, s.playing.MusicId)
	if err == nil {
		s.narrativeCache = narrativeCache{
			station:   s.playing.StationId,
			track:     s.playing.MusicId,
			narrative: result,
		}
	}

	return result, err
}

func (s *StationController) StationChanged() <-chan pandora.Station {
	return s.stationChanged
}

func (s *StationController) CurrentStation() pandora.Station {
	return s.station
}

func (s *StationController) NotificationChan() <-chan *pandora.Track {
	return s.notifications
}
