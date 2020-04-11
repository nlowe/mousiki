package pandora

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/nlowe/mousiki/pandora/api"
	"github.com/sirupsen/logrus"
)

const (
	csrfCookieName = "csrftoken"
	pandoraBase    = "https://www.pandora.com"
)

var pandoraAPIBase = fmt.Sprintf("%s/api", pandoraBase)

// Client implements the Pandora REST API defined in https://6xq.net/pandora-apidoc/rest
type Client interface {
	Login(username, password string) error
}

type client struct {
	authToken string
	csrfToken *http.Cookie

	api *http.Client
	log logrus.FieldLogger
}

func NewClient() *client {
	return &client{
		api: cleanhttp.DefaultClient(),
		log: logrus.WithField("prefix", "client"),
	}
}

func (c *client) updateCSRF() error {
	resp, err := c.api.Head(pandoraBase)
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

	if c.authToken != "" {
		r.Header.Set("X-AuthToken", c.authToken)
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
			pandoraAPIBase,
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
	resp, err := c.post("/v1/auth/login", &api.LoginRequest{
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

	payload := api.LoginResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("login: read response: %w", err)
	}

	c.authToken = payload.AuthToken

	c.log.WithFields(logrus.Fields{"user": payload.Username, "webname": payload.WebName}).Info("Successfully Logged In")
	return nil
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
