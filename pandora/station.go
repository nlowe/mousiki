package pandora

import (
	"fmt"
	"time"
)

type Station struct {
	ID                      string `json:"stationId"`
	StationFactoryPandoraId string `json:"sstationFactoryPandoraId"`
	PandoraId               string `json:"pandoraId"`
	ArtId                   string `json:"artId"`

	Name           string `json:"name"`
	CreatorWebName string `json:"creatorWebname"`
	Art            []Art  `json:"art"`

	CreatedAt  time.Time `json:"dateCreated"`
	LastPlayed time.Time `json:"lastPlayed"`
}

func (s Station) String() string {
	return fmt.Sprintf("[%s] %s", s.ID, s.Name)
}
