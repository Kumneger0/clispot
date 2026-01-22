package types // nolint:revive

type NextPageURLType string

const (
	NextPageURLTypePlaylistTracks NextPageURLType = "playlistTracks"
	NextPageURLTypeUserSavedItems NextPageURLType = "userSavedItems"
)

type PaginationInfo struct {
	Next            string
	NextPageURLType NextPageURLType
	NextItemID      string
}

type UpdatePlaylistMsg struct {
	Playlist          []*PlaylistTrackObject
	Err               error
	PaginationInfo    *PaginationInfo
	ShouldAppendQueue bool
	ShouldAppend      bool
}

type UpdatePlayedSeconds struct {
	TrackID string
}

type MessageType string

const (
	NextTrack     MessageType = "nextTrack"
	PreviousTrack MessageType = "previousTrack"
	PlayPause     MessageType = "playPause"
)

type DBusMessage struct {
	MessageType
}

type SearchingMsg struct{}

type SpotifySearchResultMsg struct {
	Result *SearchResponse
	Err    error
}

type CheckUserSavedTrackResponseMsg struct {
	Saved bool
	Err   error
}

type LikeUnlikeTrackMsg struct {
	TrackID string
	Like    bool
	Err     error
}
