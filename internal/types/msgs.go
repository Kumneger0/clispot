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
