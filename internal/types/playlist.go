package types // nolint:revive

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
	return spotifyPlaylist.Name + " (playlist)"
}
func (spotifyPlaylist SpotifyPlaylist) Title() string {
	return spotifyPlaylist.Name + " (playlist)"
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
	Href     string                 `json:"href"`
	Limit    int                    `json:"limit"`
	Next     string                 `json:"next"`
	Offset   int                    `json:"offset"`
	Previous string                 `json:"previous"`
	Total    int                    `json:"total"`
	Items    []*PlaylistTrackObject `json:"items"`
}

type PlaylistTrackObject struct {
	AddedAt string       `json:"added_at"`
	AddedBy *SpotifyUser `json:"added_by"`
	IsLocal bool         `json:"is_local"`
	Track   Track        `json:"track"`
}

func (playlist PlaylistTrackObject) FilterValue() string {
	return playlist.Track.Name
}

func (playlist PlaylistTrackObject) Title() string {
	return playlist.Track.Name
}
