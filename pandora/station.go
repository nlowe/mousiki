package pandora

import "time"

type StationArt struct {
	Url  string `json:"url"`
	Size int    `json:"size"`
}

type Station struct {
	ID                      string `json:"stationId"`
	StationFactoryPandoraId string `json:"sstationFactoryPandoraId"`
	PandoraId               string `json:"pandoraId"`
	ArtId                   string `json:"artId"`

	Name           string       `json:"name"`
	CreatorWebName string       `json:"creatorWebname"`
	Art            []StationArt `json:"art"`

	CreatedAt  time.Time `json:"dateCreated"`
	LastPlayed time.Time `json:"lastPlayed"`
}
