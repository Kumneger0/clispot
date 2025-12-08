package types // nolint:revive

type ArtistsTopTrackResponse struct {
	Tracks []Track `json:"tracks"`
}

type UserSavedTracks struct {
	Href     string       `json:"href"`
	Limit    int          `json:"limit"`
	Next     string       `json:"next"`
	Offset   int          `json:"offset"`
	Previous string       `json:"previous"`
	Total    int          `json:"total"`
	Items    []SavedTrack `json:"items"`
}

type SavedTrack struct {
	AddedAt string `json:"added_at"`
	Track   Track  `json:"track"`
}
