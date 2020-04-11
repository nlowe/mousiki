package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nlowe/mousiki/pandora"
	"github.com/nlowe/mousiki/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClientTest(t *testing.T, m *http.ServeMux, authToken string) (*client, *httptest.Server, string) {
	c := NewClient()
	csrfToken := uuid.Must(uuid.NewRandom()).String()
	csrfCookie := &http.Cookie{
		Name:   csrfCookieName,
		Value:  csrfToken,
		Path:   "/",
		Domain: ".pandora.com",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
		assert.Equal(t, r.Header.Get("X-CsrfToken"), csrfToken, "CSRF Token Set")
		testutil.AssertCookie(t, r, csrfCookieName, csrfToken)

		if !strings.HasSuffix(r.URL.Path, "/v1/auth/login") {
			assert.Equal(t, r.Header.Get("X-AuthToken"), authToken, "Auth Token Set")
		}

		m.ServeHTTP(w, r)
	})
	mux.HandleFunc("/_csrf", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, csrfCookie)
		w.WriteHeader(http.StatusOK)
	})

	sv := httptest.NewServer(mux)

	c.log = testutil.NopLogger()
	c.apiURL = fmt.Sprintf("%s/api", strings.TrimSuffix(sv.URL, "/"))
	c.csrfURL = fmt.Sprintf("%s/_csrf", strings.TrimSuffix(sv.URL, "/"))
	c.api = sv.Client()

	return c, sv, csrfToken
}

func expectLogin(t *testing.T, m *http.ServeMux, authToken string) {
	m.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		v := LoginRequest{}
		testutil.UnmarshalRequest(t, r, &v)

		assert.Equal(t, v.Username, "un", "Username")
		assert.Equal(t, v.Password, "pw", "Password")

		testutil.MarshalResponse(t, http.StatusOK, w, &LoginResponse{
			AuthToken: authToken,
			Username:  "un",
		})
	})
}

func TestClient_Login(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		authToken := uuid.Must(uuid.NewRandom()).String()
		m := http.NewServeMux()
		expectLogin(t, m, authToken)

		sut, server, _ := setupClientTest(t, m, authToken)
		defer server.Close()

		require.NoError(t, sut.Login("un", "pw"))
		assert.Equal(t, authToken, sut.authToken)
	})

	t.Run("HttpError", func(t *testing.T) {
		authToken := uuid.Must(uuid.NewRandom()).String()
		m := http.NewServeMux()
		m.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
			_, _ = w.Write([]byte("Foobar"))
		})

		sut, server, _ := setupClientTest(t, m, authToken)
		defer server.Close()

		require.EqualError(t, sut.Login("un", "pw"), "login: unexpected result 418 I'm a teapot:\nFoobar")
	})
}

func TestClient_GetStations(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		authToken := uuid.Must(uuid.NewRandom()).String()
		m := http.NewServeMux()
		expectLogin(t, m, authToken)
		m.HandleFunc("/api/v1/station/getStations", func(w http.ResponseWriter, r *http.Request) {
			v := StationRequest{}
			testutil.UnmarshalRequest(t, r, &v)

			assert.Equal(t, 250, v.PageSize)
			assert.Equal(t, 0, v.StartIndex)

			testutil.MarshalResponse(t, http.StatusOK, w, &StationResponse{
				TotalStations: 1,
				SortedBy:      StationSortOrderLastPlayed,
				Index:         0,
				Stations: []pandora.Station{
					{
						ID:                      uuid.Must(uuid.NewRandom()).String(),
						StationFactoryPandoraId: uuid.Must(uuid.NewRandom()).String(),
						PandoraId:               uuid.Must(uuid.NewRandom()).String(),
						ArtId:                   uuid.Must(uuid.NewRandom()).String(),
						Name:                    "Test Station",
						CreatorWebName:          "nlowe",
						Art:                     nil,
						CreatedAt:               time.Now().UTC().Add((-1 * 24 * 365) * time.Hour),
						LastPlayed:              time.Now().UTC(),
					},
				},
			})
		})

		sut, server, _ := setupClientTest(t, m, authToken)
		defer server.Close()

		require.NoError(t, sut.Login("un", "pw"))
		stations, err := sut.GetStations()

		require.NoError(t, err)
		require.Len(t, stations, 1)
		require.Equal(t, "Test Station", stations[0].Name)
	})

	t.Run("RequiresLogin", func(t *testing.T) {
		sut, server, _ := setupClientTest(t, http.NewServeMux(), uuid.Must(uuid.NewRandom()).String())
		defer server.Close()

		_, err := sut.GetStations()
		require.EqualError(t, err, "GetStations: post: not logged in")
	})
}

func TestClient_GetMoreTracks(t *testing.T) {
	stationId := uuid.Must(uuid.NewRandom()).String()

	t.Run("Valid", func(t *testing.T) {
		authToken := uuid.Must(uuid.NewRandom()).String()

		m := http.NewServeMux()
		expectLogin(t, m, authToken)
		m.HandleFunc("/api/v1/playlist/getFragment", func(w http.ResponseWriter, r *http.Request) {
			v := GetPlaylistFragmentRequest{}
			testutil.UnmarshalRequest(t, r, &v)

			require.Equal(t, stationId, v.StationID)

			testutil.MarshalResponse(t, http.StatusOK, w, &GetPlaylistFragmentResponse{
				Tracks: []pandora.Track{
					testutil.MakeTrack(),
					testutil.MakeTrack(),
					testutil.MakeTrack(),
					testutil.MakeTrack(),
				},
				IsBingeSkipping: false,
			})
		})

		sut, server, _ := setupClientTest(t, m, authToken)
		defer server.Close()

		require.NoError(t, sut.Login("un", "pw"))
		tracks, err := sut.GetMoreTracks(stationId)

		require.NoError(t, err)
		require.Len(t, tracks, 4)
	})

	t.Run("RequiresLogin", func(t *testing.T) {
		sut, server, _ := setupClientTest(t, http.NewServeMux(), uuid.Must(uuid.NewRandom()).String())
		defer server.Close()

		_, err := sut.GetMoreTracks(stationId)
		require.EqualError(t, err, "GetMoreTracks: post: not logged in")
	})
}
