package types // nolint:revive

type UserFollowedArtistResponse struct {
	Artists Artists `json:"artists"`
}

type Artists struct {
	Total int      `json:"total"`
	Items []Artist `json:"items"`
}

type ArtistsResponse struct {
	Artists []Artist `json:"artists"`
}

type Artist struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Images         []Image `json:"images"`
	IsItFromSearch bool    `json:"-"`
}

func (artist Artist) FilterValue() string {
	return artist.Name + " (artist)"
}
func (artist Artist) Title() string {
	return artist.Name + " (artist)"
}

type UserTopItemsResponse struct {
	Total int      `json:"total"`
	Items []Artist `json:"items"`
}
