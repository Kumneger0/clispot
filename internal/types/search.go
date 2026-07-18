package types // nolint:revive

import "github.com/charmbracelet/bubbles/list"

type Paging[T any] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}

type SearchResponse struct {
	Items []list.Item
}
