package types // nolint:revive

type SavedAlbumsResponse struct {
	Total int          `json:"total"`
	Items []SavedAlbum `json:"items"`
}

type SavedAlbum struct {
	Album Album `json:"album"`
}

type Album struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Artists  []Artist `json:"artists"`
	Images   []Image  `json:"images"`
	Year     string   `json:"year"`
	Type     string   `json:"type"` // "Album", "Single", "EP"
	Explicit bool     `json:"explicit"`
}

type AlbumTracks struct {
	Total int     `json:"total"`
	Items []Track `json:"items"`
}

type Track struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Artists    []Artist `json:"artists"`
	Album      Album    `json:"album"`
	DurationMS int      `json:"duration_ms"`
	Explicit   bool     `json:"explicit"`
	IsLocal    bool     `json:"is_local"`
	URL        string   `json:"url"`
}

type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type AlbumTracksResponse struct {
	Total int     `json:"total"`
	Items []Track `json:"items"`
}
