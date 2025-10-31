package types // nolint:revive

// type PlayMusicMsg struct {
// 	MusicPath string
// }

type UpdatePlaylistMsg struct {
	Playlist []*PlaylistTrackObject
	Err      error
}

type UpdatePlayedSeconds struct {
	CurrentSeconds float64
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
