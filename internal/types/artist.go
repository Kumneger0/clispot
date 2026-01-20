package types // nolint:revive

type UserFollowedArtistResponse struct {
	Artists Artists `json:"artists"`
}

type Artists struct {
	Href    string   `json:"href"`
	Limit   int      `json:"limit"`
	Next    string   `json:"next"`
	Cursors Cursors  `json:"cursors"`
	Total   int      `json:"total"`
	Items   []Artist `json:"items"`
}

type Cursors struct {
	After  *string `json:"after"`
	Before *string `json:"before"`
}

type ArtistsResponse struct {
	Artists []Artist `json:"artists"`
}

type Artist struct {
	ExternalURLs   ExternalURLs    `json:"external_urls"`
	Href           string          `json:"href"`
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	URI            string          `json:"uri"`
	Followers      *Followers      `json:"followers"`
	Genres         []*string       `json:"genres"`
	Images         []*SpotifyImage `json:"images"`
	Popularity     *int            `json:"popularity"`
	IsItFromSearch bool            `json:"-"`
}

func (artist Artist) FilterValue() string {
	return artist.Name + " (artist)"
}
func (artist Artist) Title() string {
	return artist.Name + " (artist)"
}

type ExternalURLs struct {
	Spotify string `json:"spotify"`
}

type Followers struct {
	Href  string `json:"href"`
	Total int    `json:"total"`
}

type UserTopItemsResponse struct {
	Href     string   `json:"href"`
	Limit    int      `json:"limit"`
	Next     string   `json:"next"`
	Offset   int      `json:"offset"`
	Previous string   `json:"previous"`
	Total    int      `json:"total"`
	Items    []Artist `json:"items"`
}
