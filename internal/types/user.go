package types // nolint:revive

type UserTokenInfo struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
}

// UserSavedTracksListItem is a simple sidebar item representing the user's liked songs.
type UserSavedTracksListItem struct {
	Name string
}

func (u UserSavedTracksListItem) FilterValue() string {
	return u.Name
}

func (u UserSavedTracksListItem) Title() string {
	return u.Name
}

type HomeSidebarItem struct {
	Name string
}

func (h HomeSidebarItem) FilterValue() string {
	return h.Name
}

func (h HomeSidebarItem) Title() string {
	return h.Name
}

type HomePageSectionItem struct {
	SectionTitle string
	Index        int
}

func (h HomePageSectionItem) FilterValue() string {
	return h.SectionTitle
}

func (h HomePageSectionItem) Title() string {
	return h.SectionTitle
}

type HomePageContentItem struct {
	ItemTitle   string
	PlaylistID  string
	Description string
}

func (h HomePageContentItem) FilterValue() string {
	return h.ItemTitle
}

func (h HomePageContentItem) Title() string {
	return h.ItemTitle
}
