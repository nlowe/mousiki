package api

type LoginRequest struct {
	ExistingAuthToken string `json:"existingAuthToken"`
	KeepLoggedIn      bool   `json:"keepLoggedIn"`
	Username          string `json:"username"`
	Password          string `json:"password"`
}

type LoginResponse struct {
	AuthToken     string `json:"authToken"`
	ListenerToken string `json:"listenerToken"`

	Username string `json:"username"`
	WebName  string `json:"webname"`

	HighQualityStreamingEnabled bool `json:"highQualityStreamingEnabled"`
}
