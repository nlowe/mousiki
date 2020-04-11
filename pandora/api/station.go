package api

import "github.com/nlowe/mousiki/pandora"

type StationSortOrder string

const StationSortOrderLastPlayed StationSortOrder = "lastPlayedTime"

type StationRequest struct {
	PageSize   int `json:"pageSize"`
	StartIndex int `json:"startIndex"`
}

type StationResponse struct {
	TotalStations int               `json:"totalStations"`
	SortedBy      StationSortOrder  `json:"sortedBy"`
	Index         int               `json:"index"`
	Stations      []pandora.Station `json:"stations"`
}
