package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/blowfish"
)

const (
	blowfishBlockSize = 8

	legacyAPIEndpoint = "https://tuner.pandora.com/services/json/"

	legacyPartnerUsername        = "android"
	legacyPartnerPassword        = "AC7IBG09A3DTSYM4R41UJWL07VLN8JI7"
	legacyPartnerDeviceID        = "android-generic"
	legacyPartnerEncryptPassword = `6#26FRL$ZWD`
	legacyPartnerDecryptPassword = `R=U!LH$O2B#`
	legacyPartnerAPIVersion      = "5"
)

func mustCipher(key string) *blowfish.Cipher {
	c, err := blowfish.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}

	return c
}

var (
	legacyEncryptCipher = mustCipher(legacyPartnerEncryptPassword)
	legacyDecryptCipher = mustCipher(legacyPartnerDecryptPassword)
)

type LegacyRequest struct {
	SyncTime int64 `json:"syncTime,omitempty"`
}

type LegacyResponse struct {
	Stat string `json:"stat"`
}

type LegacyPartnerLoginRequest struct {
	LegacyRequest

	Username    string `json:"username"`
	Password    string `json:"password"`
	DeviceModel string `json:"deviceModel"`
	Version     string `json:"version"`

	IncludeURLs                bool `json:"includeUrls"`
	ReturnDeviceType           bool `json:"returnDeviceType"`
	ReturnUpdatePromptVersions bool `json:"returnUpdatePromptVersions"`
}

type LegacyPartnerLoginResponseResult struct {
	EncryptedSyncTime string `json:"syncTime"`
	PartnerID         string `json:"partnerId"`
	PartnerAuthToken  string `json:"partnerAuthToken"`
}

type LegacyPartnerLoginResponse struct {
	LegacyResponse
	Result LegacyPartnerLoginResponseResult `json:"result"`
}

type LegacyUserLoginRequest struct {
	LegacyRequest

	LoginType        string `json:"loginType"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	PartnerAuthToken string `json:"partnerAuthToken"`
}

type LegacyUserLoginResponseResult struct {
	UserAuthToken string `json:"userAuthToken"`
}

type LegacyUserLoginResponse struct {
	LegacyResponse
	Result LegacyUserLoginResponseResult `json:"result"`
}

func (c *client) legacyPost(method, authToken, partnerId string, encrypt bool, payload interface{}) (*http.Response, error) {
	buff, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("legacyPost: marshal payload: %w", err)
	}

	var req *http.Request

	if encrypt {
		buff = blowfishPad(buff)
		encrypted := make([]byte, len(buff))
		legacyEncrypt(encrypted, buff)

		encoded := hex.EncodeToString(encrypted)
		req, err = http.NewRequest(http.MethodPost, legacyAPIEndpoint, strings.NewReader(encoded))
	} else {
		req, err = http.NewRequest(http.MethodPost, legacyAPIEndpoint, bytes.NewReader(buff))
	}

	if err != nil {
		return nil, fmt.Errorf("legacyPost: make request: %w", err)
	}

	q := req.URL.Query()
	q.Set("method", method)

	if authToken != "" {
		q.Set("auth_token", authToken)
	}

	if partnerId != "" {
		q.Set("partner_id", partnerId)
	}

	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-type", "text/plain")

	return c.api.Do(req)
}

func (c *client) legacyPartnerLogin() (LegacyPartnerLoginResponseResult, error) {
	c.log.WithFields(logrus.Fields{}).Trace("Attempting Partner Login")
	// Perform Partner Login
	resp, err := c.legacyPost("auth.partnerLogin", "", "", false, LegacyPartnerLoginRequest{
		Username:    legacyPartnerUsername,
		Password:    legacyPartnerPassword,
		DeviceModel: legacyPartnerDeviceID,
		Version:     legacyPartnerAPIVersion,
	})

	if err != nil {
		return LegacyPartnerLoginResponseResult{}, fmt.Errorf("partner login: %w", err)
	}

	defer mustClose(resp.Body)
	if err := checkHttpCode(resp); err != nil {
		return LegacyPartnerLoginResponseResult{}, fmt.Errorf("partner login: %w", err)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return LegacyPartnerLoginResponseResult{}, fmt.Errorf("partner login: read response: %w", err)
	}

	payload := LegacyPartnerLoginResponse{}
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&payload); err != nil {
		return LegacyPartnerLoginResponseResult{}, fmt.Errorf("partner login: decode response: %w", err)
	}

	if payload.Stat != "ok" {
		return LegacyPartnerLoginResponseResult{}, fmt.Errorf("partner login: request failed: %s", string(raw))
	}

	c.log.WithFields(logrus.Fields{
		"stat":              payload.Stat,
		"partnerAuthToken":  payload.Result.PartnerAuthToken,
		"partnerID":         payload.Result.PartnerID,
		"encryptedSyncTime": payload.Result.EncryptedSyncTime,
	}).Trace("Partner Login Complete")
	return payload.Result, nil
}

func (c *client) LegacyLogin(username, password string) error {
	c.log.WithField("username", username).Debug("Attempting legacy login")

	partner, err := c.legacyPartnerLogin()
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	rawEncryptedSyncTime, err := hex.DecodeString(partner.EncryptedSyncTime)
	if err != nil {
		return fmt.Errorf("login: decode sync time: %w", err)
	}

	rawDecryptedSyncTime := make([]byte, len(rawEncryptedSyncTime))
	legacyDecrypt(rawDecryptedSyncTime, rawEncryptedSyncTime)

	c.log.WithField("decryptedPartnerSyncTime", string(rawDecryptedSyncTime)).Trace("Decrypted Partner Sync Time")

	partnerSyncTime, err := strconv.ParseInt(string(rawDecryptedSyncTime[4:14]), 10, 32)
	if err != nil {
		return fmt.Errorf("login: decode decrypted sync time: %w", err)
	}

	// TODO: We're calling these back-to-back. Do we even need sync time?
	sync := time.Now().Add(time.Unix(partnerSyncTime, 0).Sub(time.Now())).UTC().Unix()
	c.log.WithFields(logrus.Fields{
		"partnerSyncTime": partnerSyncTime,
		"syncTime":        sync,
	}).Trace("Parsed partner sync time")

	// TODO: Pandora doesn't like this call, always returns code 0
	c.log.WithField("username", username).Trace("Performing User Login")
	resp, err := c.legacyPost("auth.userLogin", partner.PartnerAuthToken, partner.PartnerID, true, LegacyUserLoginRequest{
		LegacyRequest: LegacyRequest{
			SyncTime: sync,
		},
		LoginType:        "user",
		Username:         username,
		Password:         password,
		PartnerAuthToken: partner.PartnerAuthToken,
	})

	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	defer mustClose(resp.Body)
	if err := checkHttpCode(resp); err != nil {
		return fmt.Errorf("login: %w", err)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("login: read response: %w", err)
	}

	payload := LegacyUserLoginResponse{}
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&payload); err != nil {
		return fmt.Errorf("login: decode response: %w", err)
	}

	if payload.Stat != "ok" {
		return fmt.Errorf("login: request failed: %s", string(raw))
	}

	c.log.WithField("authToken", payload.Result.UserAuthToken).Trace("User Login Complete")
	c.authToken = payload.Result.UserAuthToken
	return nil
}

func blowfishPad(b []byte) []byte {
	return append(b, make([]byte, blowfishBlockSize-(len(b)%blowfishBlockSize))...)
}

func legacyEncrypt(dst []byte, src []byte) {
	for bs, be := 0, blowfishBlockSize; bs < len(src); bs, be = bs+blowfishBlockSize, be+blowfishBlockSize {
		legacyEncryptCipher.Encrypt(dst[bs:be], src[bs:be])
	}
}

func legacyDecrypt(dst []byte, src []byte) {
	for bs, be := 0, blowfishBlockSize; bs < len(src); bs, be = bs+blowfishBlockSize, be+blowfishBlockSize {
		legacyDecryptCipher.Decrypt(dst[bs:be], src[bs:be])
	}
}
