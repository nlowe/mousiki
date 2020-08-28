package pandora

type Narrative struct {
	Intro       string   `json:"intro"`
	FocusTraits []string `json:"focusTraits"`
	Paragraph   string   `json:"paragraph"`
}
