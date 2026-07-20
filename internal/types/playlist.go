package types // nolint:revive

type Playlist struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Images         []Image `json:"images"`
	Count          int     `json:"count"`
	Author         string  `json:"author"`
	IsItFromSearch bool    `json:"-"`
}

func (p Playlist) FilterValue() string {
	return p.Name + " (playlist)"
}
func (p Playlist) Title() string {
	return p.Name + " (playlist)"
}

type PlaylistPage struct {
	Total int        `json:"total"`
	Items []Playlist `json:"items"`
}

type FeaturedPlaylistsResponse struct {
	Message   string       `json:"message"`
	Playlists PlaylistPage `json:"playlists"`
}

type UserPlaylistsResponse = PlaylistPage

type PlaylistItemsResponse struct {
	Total int                    `json:"total"`
	Items []*PlaylistTrackObject `json:"items"`
}

type PlaylistTrackObject struct {
	Track          Track `json:"track"`
	IsItFromQueue  bool  `json:"isItFromQueue"`
	IsItFromSearch bool  `json:"-"`
}

func (playlist PlaylistTrackObject) FilterValue() string {
	return playlist.Track.Name
}

func (playlist PlaylistTrackObject) Title() string {
	return playlist.Track.Name
}
