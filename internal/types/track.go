package types

type SpotifyTrack struct {
	Album struct {
		AlbumType        string   `json:"album_type"`
		TotalTracks      int      `json:"total_tracks"`
		AvailableMarkets []string `json:"available_markets"`
		ExternalURLs     struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			URL    string `json:"url"`
			Height int    `json:"height"`
			Width  int    `json:"width"`
		} `json:"images"`
		Name                 string `json:"name"`
		ReleaseDate          string `json:"release_date"`
		ReleaseDatePrecision string `json:"release_date_precision"`
		Restrictions         struct {
			Reason string `json:"reason"`
		} `json:"restrictions"`
		Type    string `json:"type"`
		URI     string `json:"uri"`
		Artists []struct {
			ExternalURLs struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
	} `json:"album"`

	Artists []struct {
		ExternalURLs struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"artists"`

	AvailableMarkets []string `json:"available_markets"`
	DiscNumber       int      `json:"disc_number"`
	DurationMS       int      `json:"duration_ms"`
	Explicit         bool     `json:"explicit"`

	ExternalIDs struct {
		ISRC string `json:"isrc"`
		EAN  string `json:"ean"`
		UPC  string `json:"upc"`
	} `json:"external_ids"`

	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`

	Href        string `json:"href"`
	ID          string `json:"id"`
	IsPlayable  bool   `json:"is_playable"`
	Name        string `json:"name"`
	Popularity  int    `json:"popularity"`
	PreviewURL  string `json:"preview_url"`
	TrackNumber int    `json:"track_number"`
	Type        string `json:"type"`
	URI         string `json:"uri"`
	IsLocal     bool   `json:"is_local"`
}
