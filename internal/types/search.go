package types

type Restrictions struct {
	Reason string `json:"reason"`
}

type ResumePoint struct {
	FullyPlayed      bool `json:"fully_played"`
	ResumePositionMS int  `json:"resume_position_ms"`
}

type Author struct {
	Name string `json:"name"`
}

type Narrator struct {
	Name string `json:"name"`
}

type Paging[T any] struct {
	Href     string `json:"href"`
	Limit    int    `json:"limit"`
	Next     string `json:"next"`
	Offset   int    `json:"offset"`
	Previous string `json:"previous"`
	Total    int    `json:"total"`
	Items    []T    `json:"items"`
}

type PlaylistOwner struct {
	ExternalURLs ExternalURLs `json:"external_urls"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
	DisplayName  string       `json:"display_name"`
}

type PlaylistTracks struct {
	Href  string `json:"href"`
	Total int    `json:"total"`
}

type Show struct {
	AvailableMarkets   []string     `json:"available_markets"`
	Copyrights         []Copyright  `json:"copyrights"`
	Description        string       `json:"description"`
	HTMLDescription    string       `json:"html_description"`
	Explicit           bool         `json:"explicit"`
	ExternalURLs       ExternalURLs `json:"external_urls"`
	Href               string       `json:"href"`
	ID                 string       `json:"id"`
	Images             []Image      `json:"images"`
	IsExternallyHosted bool         `json:"is_externally_hosted"`
	Languages          []string     `json:"languages"`
	MediaType          string       `json:"media_type"`
	Name               string       `json:"name"`
	Publisher          string       `json:"publisher"`
	Type               string       `json:"type"`
	URI                string       `json:"uri"`
	TotalEpisodes      int          `json:"total_episodes"`
}

type Episode struct {
	AudioPreviewURL      string       `json:"audio_preview_url"`
	Description          string       `json:"description"`
	HTMLDescription      string       `json:"html_description"`
	DurationMS           int          `json:"duration_ms"`
	Explicit             bool         `json:"explicit"`
	ExternalURLs         ExternalURLs `json:"external_urls"`
	Href                 string       `json:"href"`
	ID                   string       `json:"id"`
	Images               []Image      `json:"images"`
	IsExternallyHosted   bool         `json:"is_externally_hosted"`
	IsPlayable           bool         `json:"is_playable"`
	Language             string       `json:"language"`
	Languages            []string     `json:"languages"`
	Name                 string       `json:"name"`
	ReleaseDate          string       `json:"release_date"`
	ReleaseDatePrecision string       `json:"release_date_precision"`
	ResumePoint          ResumePoint  `json:"resume_point"`
	Type                 string       `json:"type"`
	URI                  string       `json:"uri"`
	Restrictions         Restrictions `json:"restrictions"`
}

type Audiobook struct {
	Authors          []Author     `json:"authors"`
	AvailableMarkets []string     `json:"available_markets"`
	Copyrights       []Copyright  `json:"copyrights"`
	Description      string       `json:"description"`
	HTMLDescription  string       `json:"html_description"`
	Edition          string       `json:"edition"`
	Explicit         bool         `json:"explicit"`
	ExternalURLs     ExternalURLs `json:"external_urls"`
	Href             string       `json:"href"`
	ID               string       `json:"id"`
	Images           []Image      `json:"images"`
	Languages        []string     `json:"languages"`
	MediaType        string       `json:"media_type"`
	Name             string       `json:"name"`
	Narrators        []Narrator   `json:"narrators"`
	Publisher        string       `json:"publisher"`
	Type             string       `json:"type"`
	URI              string       `json:"uri"`
	TotalChapters    int          `json:"total_chapters"`
}

type SearchResponse struct {
	Tracks     Paging[Track]     `json:"tracks"`
	Artists    Paging[Artist]    `json:"artists"`
	Albums     Paging[Album]     `json:"albums"`
	Playlists  Paging[Playlist]  `json:"playlists"`
	Shows      Paging[Show]      `json:"shows"`
	Episodes   Paging[Episode]   `json:"episodes"`
	Audiobooks Paging[Audiobook] `json:"audiobooks"`
}
