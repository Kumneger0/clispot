package headless

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kumneger0/clispot/internal/spotify"
	"github.com/kumneger0/clispot/internal/types"
	"github.com/kumneger0/clispot/internal/ui"
	"github.com/kumneger0/clispot/internal/youtube"
)

type UserLibrary struct {
	// LikedSongs         []types.SavedTrack `json:"likedSongs"`
	Playlist           []types.Playlist `json:"playlist"`
	UserFollowedArtist []types.Artist   `json:"artist"`
}

type TracksType string

const (
	PlaylistType   TracksType = "playlist"
	FollowedArtist TracksType = "followed_artist"
)

type TracksResponse struct {
	Tracks []*types.PlaylistTrackObject `json:"tracks"`
}

type PlayRequestBodyType struct {
	TrackID   string   `json:"trackID"`
	TrackName string   `json:"name"`
	Artists   []string `json:"artists"`
	AlbumName string   `json:"album"`
}

func StartServer(m *ui.Model) {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8282",
		Handler: mux,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "server is  up")
	})

	mux.HandleFunc("/library", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userToken := m.GetUserToken()
		if userToken == nil {
			slog.Error("failed to get userToken")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		userPlaylist, err := spotify.GetUserPlaylists(userToken.AccessToken)
		if err != nil {
			slog.Error("GetUserPlaylists failed: " + err.Error())
			http.Error(w, `{"error":"failed to fetch user playlists"}`, http.StatusInternalServerError)
			return
		}
		if userPlaylist == nil {
			slog.Error("GetUserPlaylists returned nil playlist")
			http.Error(w, `{"error":"no playlist data returned"}`, http.StatusInternalServerError)
			return
		}

		followedArtists, err := spotify.GetFollowedArtist(userToken.AccessToken)
		if err != nil {
			slog.Error("GetFollowedArtist failed: " + err.Error())
			http.Error(w, `{"error":"failed to fetch followed artists"}`, http.StatusInternalServerError)
			return
		}
		if followedArtists == nil {
			slog.Error("GetFollowedArtist returned nil")
			http.Error(w, `{"error":"no artist data returned"}`, http.StatusInternalServerError)
			return
		}

		userLibrary := &UserLibrary{
			Playlist:           userPlaylist.Items,
			UserFollowedArtist: followedArtists.Artists.Items,
		}

		data, err := json.Marshal(userLibrary)
		if err != nil {
			slog.Error("json.Marshal failed: " + err.Error())
			http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			slog.Error(err.Error())
		}
	})

	mux.HandleFunc("/tracks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, `{"error":"missing required query param: id"}`, http.StatusBadRequest)
			return
		}

		queryType := r.URL.Query().Get("type")
		if queryType == "" {
			http.Error(w, `{"error":"missing required query param: type"}`, http.StatusBadRequest)
			return
		}

		userToken := m.GetUserToken()
		if userToken == nil {
			slog.Error("failed to get userToken")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		if TracksType(queryType) == FollowedArtist {
			artistSongs, err := spotify.GetArtistsTopTrackURL(userToken.AccessToken, id)
			if err != nil {
				slog.Error(err.Error())
				http.Error(w, `{"error":"failed to fetch artist tracks"}`, http.StatusInternalServerError)
				return
			}

			var tracks []*types.PlaylistTrackObject
			for _, track := range artistSongs.Tracks {
				tracks = append(tracks, &types.PlaylistTrackObject{
					AddedAt: "",
					AddedBy: nil,
					IsLocal: false,
					Track:   track,
				})
			}

			resp := &TracksResponse{Tracks: tracks}
			data, err := json.Marshal(resp)
			if err != nil {
				slog.Error(err.Error())
				http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, err = w.Write(data)
			if err != nil {
				slog.Error(err.Error())
			}
			return
		}

		if TracksType(queryType) == PlaylistType {
			playlistItems, err := spotify.GetPlaylistItems(id, userToken.AccessToken)
			if err != nil {
				slog.Error(err.Error())
				http.Error(w, `{"error":"failed to fetch playlist items"}`, http.StatusInternalServerError)
				return
			}

			resp := &TracksResponse{Tracks: playlistItems.Items}
			data, err := json.Marshal(resp)
			if err != nil {
				slog.Error(err.Error())
				http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, err = w.Write(data)
			if err != nil {
				slog.Error(err.Error())
			}
			return
		}

		http.Error(w, `{"error":"unknown type"}`, http.StatusBadRequest)
	})

	mux.HandleFunc("/player", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var action string
		if m.PlayerProcess != nil && m.PlayerProcess.OtoPlayer.IsPlaying() {
			action = "paused"
		} else {
			action = "play"
		}
		m.HandleMusicPausePlay()
		resp := map[string]any{
			"status": "ok",
			"action": action,
		}
		data, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(data)
		if err != nil {
			slog.Error(err.Error())
		}
	})

	mux.HandleFunc("POST /player/play", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		if r.Body == nil {
			http.Error(w, `{"error":"request body required"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var reqBody PlayRequestBodyType
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			slog.Error("failed to decode body: " + err.Error())
			http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
			return
		}

		if reqBody.TrackID == "" {
			http.Error(w, `{"error":"trackID is required"}`, http.StatusBadRequest)
			return
		}
		if reqBody.TrackName == "" {
			http.Error(w, `{"error":"trackName is required"}`, http.StatusBadRequest)
			return
		}
		if len(reqBody.Artists) == 0 {
			http.Error(w, `{"error":"artists is required"}`, http.StatusBadRequest)
			return
		}

		if m.PlayerProcess != nil && m.PlayerProcess.OtoPlayer.IsPlaying() {
			err := m.PlayerProcess.Close(true)
			slog.Error(err.Error())
		}

		process, err := youtube.SearchAndDownloadMusic(
			reqBody.TrackName,
			reqBody.AlbumName,
			reqBody.Artists,
			reqBody.TrackID,
			m.PlayerProcess == nil,
		)

		if err != nil {
			slog.Error("SearchAndDownloadMusic failed: " + err.Error())
			http.Error(w, `{"error":"failed to play track"}`, http.StatusInternalServerError)
			return
		}

		m.PlayerProcess = process

		resp := map[string]any{
			"status":  "ok",
			"message": "track is now playing",
			"track": map[string]any{
				"id":      reqBody.TrackID,
				"name":    reqBody.TrackName,
				"album":   reqBody.AlbumName,
				"artists": reqBody.Artists,
			},
		}

		data, err := json.Marshal(resp)
		if err != nil {
			slog.Error("failed to encode response: " + err.Error())
			http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			slog.Error(err.Error())
		}
	})

	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ctx := r.Context()

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "event: connected\ndata: ok\n\n")
		flusher.Flush()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				if m.PlayerProcess == nil || m.PlayerProcess.ByteCounterReader == nil {
					fmt.Fprintf(w, "data: 0\n\n")
				} else {
					seconds := m.PlayerProcess.ByteCounterReader.CurrentSeconds()
					msg, _ := json.Marshal(map[string]float64{"seconds": seconds})
					fmt.Fprintf(w, "data: %s\n\n", msg)
				}
				flusher.Flush()
			}
		}
	})

	fmt.Println("🚀 Server started at http://localhost:8282")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error(err.Error())
		fmt.Printf("❌ Server error: %v\n", err)
	}
}
