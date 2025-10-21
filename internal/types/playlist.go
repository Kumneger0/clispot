package types

type SpotifyPlaylist struct {
	Collaborative bool   `json:"collaborative"`
	Description   string `json:"description"`
	ExternalURLs  struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href   string `json:"href"`
	ID     string `json:"id"`
	Images []struct {
		URL    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	} `json:"images"`
	Name  string `json:"name"`
	Owner struct {
		ExternalURLs struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href        string `json:"href"`
		ID          string `json:"id"`
		Type        string `json:"type"`
		URI         string `json:"uri"`
		DisplayName string `json:"display_name"`
	} `json:"owner"`
	Public     bool   `json:"public"`
	SnapshotID string `json:"snapshot_id"`
	Tracks     struct {
		Href  string `json:"href"`
		Total int    `json:"total"`
	} `json:"tracks"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}

func (spotifyPlaylist SpotifyPlaylist) FilterValue() string {
	return spotifyPlaylist.Name
}
func (spotifyPlaylist SpotifyPlaylist) Title() string {
	return spotifyPlaylist.Name
}

type PlaylistPage struct {
	Href     string            `json:"href"`
	Limit    int               `json:"limit"`
	Next     string            `json:"next"`
	Offset   int               `json:"offset"`
	Previous string            `json:"previous"`
	Total    int               `json:"total"`
	Items    []SpotifyPlaylist `json:"items"`
}

type FeaturedPlaylistsResponse struct {
	Message   string       `json:"message"`
	Playlists PlaylistPage `json:"playlists"`
}

type UserPlaylistsResponse = PlaylistPage

type SpotifyImage struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type PlaylistItemsResponse struct {
	Href     string                `json:"href"`
	Limit    int                   `json:"limit"`
	Next     string                `json:"next"`
	Offset   int                   `json:"offset"`
	Previous string                `json:"previous"`
	Total    int                   `json:"total"`
	Items    []PlaylistTrackObject `json:"items"`
}

type PlaylistTrackObject struct {
	AddedAt string      `json:"added_at"`
	AddedBy SpotifyUser `json:"added_by"`
	IsLocal bool        `json:"is_local"`
	Track   TrackObject `json:"track"`
}

func (playlist PlaylistTrackObject) FilterValue() string {
	return playlist.Track.Name
}

func (playlist PlaylistTrackObject) Title() string {
	return playlist.Track.Name
}

type TrackObject struct {
	Album            AlbumObject    `json:"album"`
	Artists          []ArtistObject `json:"artists"`
	AvailableMarkets []string       `json:"available_markets"`
	DiscNumber       int            `json:"disc_number"`
	DurationMs       int            `json:"duration_ms"`
	Explicit         bool           `json:"explicit"`
	ExternalIDs      struct {
		ISRC string `json:"isrc"`
		EAN  string `json:"ean"`
		UPC  string `json:"upc"`
	} `json:"external_ids"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href         string         `json:"href"`
	ID           string         `json:"id"`
	IsPlayable   bool           `json:"is_playable"`
	LinkedFrom   map[string]any `json:"linked_from"` // sometimes empty, keep flexible
	Restrictions struct {
		Reason string `json:"reason"`
	} `json:"restrictions"`
	Name        string `json:"name"`
	Popularity  int    `json:"popularity"`
	PreviewURL  string `json:"preview_url"`
	TrackNumber int    `json:"track_number"`
	Type        string `json:"type"`
	URI         string `json:"uri"`
	IsLocal     bool   `json:"is_local"`
}

type AlbumObject struct {
	AlbumType        string   `json:"album_type"`
	TotalTracks      int      `json:"total_tracks"`
	AvailableMarkets []string `json:"available_markets"`
	ExternalURLs     struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href                 string         `json:"href"`
	ID                   string         `json:"id"`
	Images               []SpotifyImage `json:"images"`
	Name                 string         `json:"name"`
	ReleaseDate          string         `json:"release_date"`
	ReleaseDatePrecision string         `json:"release_date_precision"`
	Restrictions         struct {
		Reason string `json:"reason"`
	} `json:"restrictions"`
	Type    string         `json:"type"`
	URI     string         `json:"uri"`
	Artists []ArtistObject `json:"artists"`
}

type ArtistObject struct {
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href string `json:"href"`
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}
