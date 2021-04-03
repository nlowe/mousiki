package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/spf13/viper"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/nlowe/mousiki/pandora"
	"github.com/sirupsen/logrus"
)

const (
	csrfCookieName = "csrftoken"
	pandoraBase    = "https://www.pandora.com"
	userAgent      = "Mozilla/5.0 (X11; Datanyze; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.92 Safari/537.36"
)

// Client implements the Pandora REST API defined in https://6xq.net/pandora-apidoc/rest
type Client interface {
	Login(username, password string) error
	GetStations() ([]pandora.Station, error)
	GetMoreTracks(stationId string) ([]pandora.Track, error)
	AddFeedback(trackToken string, isPositive bool) error
	AddTired(trackToken string) error
	GetNarrative(stationId, musicId string) (pandora.Narrative, error)
}

type client struct {
	authToken string
	csrfToken *http.Cookie

	apiURL  string
	csrfURL string

	api *http.Client
	log logrus.FieldLogger
}

func NewClient() *client {
	return &client{
		apiURL:  fmt.Sprintf("%s/api", pandoraBase),
		csrfURL: pandoraBase,

		api: cleanhttp.DefaultClient(),
		log: logrus.WithField("prefix", "client"),
	}
}

func (c *client) updateCSRF() error {
	resp, err := c.api.Head(c.csrfURL)
	if err != nil {
		return fmt.Errorf("update csrf: %w", err)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == csrfCookieName {
			c.csrfToken = cookie
			return nil
		}
	}

	return errors.New("CSRF Cookie not found")
}

func (c *client) prepare(r *http.Request) error {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("User-Agent", userAgent)

	if c.authToken != "" {
		r.Header.Set("X-AuthToken", c.authToken)
	} else if !strings.HasSuffix(r.URL.Path, "/v1/auth/login") {
		return errors.New("not logged in")
	}

	if c.csrfToken == nil {
		if err := c.updateCSRF(); err != nil {
			return fmt.Errorf("prepare request: %w", err)
		}
	}

	r.Header.Set("X-CsrfToken", c.csrfToken.Value)
	r.AddCookie(c.csrfToken)

	return nil
}

func (c *client) post(relPath string, payload interface{}) (*http.Response, error) {
	buff, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("post: marshal payload: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"%s/%s",
			c.apiURL,
			strings.TrimPrefix(relPath, "/"),
		),
		bytes.NewReader(buff),
	)

	if err != nil {
		return nil, err
	}

	if err := c.prepare(req); err != nil {
		return nil, fmt.Errorf("post: %w", err)
	}

	return c.api.Do(req)
}

func (c *client) Login(username, password string) error {
	c.log.WithField("username", username).Debug("Attempting to log in")
	resp, err := c.post("/v1/auth/login", &LoginRequest{
		KeepLoggedIn: true,
		Username:     username,
		Password:     password,
	})

	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	defer mustClose(resp.Body)
	if err := checkHttpCode(resp); err != nil {
		return fmt.Errorf("login: %w", err)
	}

	payload := LoginResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("login: read response: %w", err)
	}

	c.authToken = payload.AuthToken

	c.log.WithFields(logrus.Fields{"user": payload.Username, "webname": payload.WebName}).Info("Successfully Logged In")
	return nil
}

func (c *client) GetStations() ([]pandora.Station, error) {
	c.log.Debug("Fetching Stations")
	resp, err := c.post("/v1/station/getStations", &StationRequest{
		PageSize:   250,
		StartIndex: 0,
	})

	if err != nil {
		return nil, fmt.Errorf("GetStations: %w", err)
	}

	defer mustClose(resp.Body)
	if err := checkHttpCode(resp); err != nil {
		return nil, fmt.Errorf("GetStations: %w", err)
	}

	payload := StationResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("GetStations: read response: %w", err)
	}

	// TODO: Paging
	return payload.Stations, nil
}

func (c *client) GetMoreTracks(stationId string) ([]pandora.Track, error) {
	f := pandora.AudioFormat(viper.GetString("audio-format"))
	c.log.WithFields(logrus.Fields{
		"station":     stationId,
		"audioFormat": f,
	}).Debug("Fetching more tracks")

	// TODO: What audio formats can we request?
	// TODO: It doesn't seem to matter what format we request, pandora always gives us aacplus
	// TODO: Does StartingAtTrackId need to be set when continuing to play a station?
	resp, err := c.post("/v1/playlist/getFragment", &GetPlaylistFragmentRequest{
		StationID:             stationId,
		IsStationStart:        true,
		FragmentRequestReason: FragmentRequestReasonNormal,
		AudioFormat:           f,
	})

	if err != nil {
		return nil, fmt.Errorf("GetMoreTracks: %w", err)
	}

	defer mustClose(resp.Body)
	if err := checkHttpCode(resp); err != nil {
		return nil, fmt.Errorf("GetMoreTracks: %w", err)
	}

	payload := GetPlaylistFragmentResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("GetMoreTracks: read response: %w", err)
	}

	if payload.IsBingeSkipping {
		c.log.Warn("Pandora thinks you're skipping tracks too frequently")
	}

	return payload.Tracks, nil
}

func (c *client) AddFeedback(trackToken string, isPositive bool) error {
	c.log.WithFields(logrus.Fields{
		"track":      trackToken,
		"isPositive": isPositive,
	}).Debug("Adding Feedback")

	resp, err := c.post("/v1/station/addFeedback", &AddFeedbackRequest{
		TrackToken: trackToken,
		IsPositive: isPositive,
	})

	if err != nil {
		return fmt.Errorf("AddFeedback: %w", err)
	}

	defer mustClose(resp.Body)
	if err := checkHttpCode(resp); err != nil {
		return fmt.Errorf("AddFeedback: %w", err)
	}

	payload := AddFeedbackResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("AddFeedback: read response: %w", err)
	}

	c.log.WithFields(logrus.Fields{
		"trackToken": trackToken,
		"feedbackId": payload.ID,
		"isPositive": payload.IsPositive,
	}).Debug("Feedback Added")

	return nil
}

func (c *client) AddTired(trackToken string) error {
	c.log.WithFields(logrus.Fields{
		"track": trackToken,
	}).Debug("Adding Tired Song")

	resp, err := c.post("/v1/listener/addTiredSong", &AddTiredRequest{
		TrackToken: trackToken,
	})

	if err != nil {
		return fmt.Errorf("AddTired: %w", err)
	}

	defer mustClose(resp.Body)
	if err := checkHttpCode(resp); err != nil {
		return fmt.Errorf("AddTired: %w", err)
	}

	payload := AddTiredResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("AddTired: read response: %w", err)
	}

	return nil
}

func (c *client) GetNarrative(stationId, musicId string) (pandora.Narrative, error) {
	c.log.WithFields(logrus.Fields{
		"station": stationId,
		"track":   musicId,
	}).Debug("Getting Narrative")

	resp, err := c.post("/v1/playlist/narrative", &NarrativeRequest{
		MusicID:   musicId,
		StationID: stationId,
	})

	if err != nil {
		return pandora.Narrative{}, fmt.Errorf("GetNarrative: %w", err)
	}

	defer mustClose(resp.Body)
	if err := checkHttpCode(resp); err != nil {
		return pandora.Narrative{}, fmt.Errorf("GetNarrative: read response: %w", err)
	}

	var payload pandora.Narrative
	return payload, json.NewDecoder(resp.Body).Decode(&payload)
}

func checkHttpCode(r *http.Response) error {
	if r.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(r.Body)
		return fmt.Errorf("unexpected result %s:\n%s", r.Status, string(body))
	}

	return nil
}

func mustClose(c io.Closer) {
	if err := c.Close(); err != nil {
		panic(err)
	}
}
