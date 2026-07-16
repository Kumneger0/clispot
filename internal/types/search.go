package types // nolint:revive

type Paging[T any] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}

type SearchResponse struct {
	Tracks    Paging[Track]    `json:"tracks"`
	Artists   Paging[Artist]   `json:"artists"`
	Albums    Paging[Album]    `json:"albums"`
	Playlists Paging[Playlist] `json:"playlists"`
}
