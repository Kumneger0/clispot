package spotify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kumneger0/clispot/internal/types"
	"github.com/stretchr/testify/assert"
)

type mockAPIURL struct {
	baseURL string
}

func (m mockAPIURL) GetPlaylistBaseURL() string { return m.baseURL + "/me/playlists/" }
func (m mockAPIURL) GetTrackBaseURL() string    { return m.baseURL + "/tracks/" }
func (m mockAPIURL) GetUserProfileBaseURL() string {
	return m.baseURL + "/me/"
}
func (m mockAPIURL) GetPlaylistItems(playlistID string) string {
	return m.baseURL + "/playlists/" + playlistID + "/tracks"
}
func (m mockAPIURL) GetArtistsURL() string        { return m.baseURL + "/artists/" }
func (m mockAPIURL) GetFollowedArtistURL() string { return m.baseURL + "/me/following?type=artist" }
func (m mockAPIURL) GetUserTopItems(itemType UserTopItem) string {
	return m.baseURL + "/me/top/" + string(itemType)
}
func (m mockAPIURL) GetArtistsTopTrackURL(id string) string {
	return m.baseURL + "/artists/" + id + "/top-tracks"
}
func (m mockAPIURL) GetSearchURL(q string) string      { return m.baseURL + "/search?q=" + q }
func (m mockAPIURL) GetUserSavedTrackURL() string      { return m.baseURL + "/me/tracks" }
func (m mockAPIURL) GetCheckUserSavedTrackURL() string { return m.baseURL + "/me/tracks/contains?ids=" }
func (m mockAPIURL) GetUserSavedAlbumsBaseURL() string { return m.baseURL + "/me/albums" }
func (m mockAPIURL) GetAlbumTracksURL(albumID string) string {
	return m.baseURL + "/albums/" + albumID + "/tracks"
}
func (m mockAPIURL) GetUserPlaylistsBaseURL() string { return m.baseURL + "/me/playlists/" }
func (m mockAPIURL) GetTrackURL(trackID string) string {
	return m.baseURL + "/tracks/" + trackID
}

func TestGetUserProfile(t *testing.T) {
	expectedProfile := &types.SpotifyUserProfile{
		ID:          "user123",
		DisplayName: "Test User",
		Country:     "US",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/me/", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedProfile)
		assert.NoError(t, err)
	}))
	defer server.Close()

	mockURLs := mockAPIURL{baseURL: server.URL}
	client := NewAPIClient(mockURLs, server.Client())

	profile, err := client.GetUserProfile("test-token")

	assert.NoError(t, err)
	assert.Equal(t, expectedProfile, profile)
}

func TestGetTrack(t *testing.T) {
	expectedTrack := &types.Track{
		Name: "Test Track",
		ID:   "track123",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/tracks/track123", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedTrack)
		assert.NoError(t, err)
	}))
	defer server.Close()

	mockURLs := mockAPIURL{baseURL: server.URL}
	client := NewAPIClient(mockURLs, server.Client())

	track, err := client.GetTrack("test-token", "track123")

	assert.NoError(t, err)
	assert.Equal(t, expectedTrack.Name, track.Name)
	assert.Equal(t, expectedTrack.ID, track.ID)
}

func TestGetUserSavedTracks(t *testing.T) {
	expectedTracks := &types.UserSavedTracks{
		Items: []types.SavedTrack{
			{
				Track: types.Track{
					Name: "Saved Track",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/me/tracks", r.URL.Path)
		assert.Equal(t, "30", r.URL.Query().Get("limit"))

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedTracks)
		assert.NoError(t, err)
	}))
	defer server.Close()

	mockURLs := mockAPIURL{baseURL: server.URL}
	client := NewAPIClient(mockURLs, server.Client())

	tracks, err := client.GetUserSavedTracks("test-token", nil)

	assert.NoError(t, err)
	assert.Equal(t, "Saved Track", tracks.Items[0].Track.Name)
}

func TestCheckUserSavedTrack(t *testing.T) {
	expectedResult := []bool{true}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/me/tracks/contains", r.URL.Path)
		assert.Equal(t, "track123", r.URL.Query().Get("ids"))

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedResult)
		assert.NoError(t, err)
	}))
	defer server.Close()

	mockURLs := mockAPIURL{baseURL: server.URL}
	client := NewAPIClient(mockURLs, server.Client())

	result, err := client.CheckUserSavedTrack("test-token", "track123")

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestGetAlbumTracks(t *testing.T) {
	expectedTracks := &types.AlbumTracksResponse{
		Total: 1,
		Items: []types.Track{{Name: "Album Track", ID: "track789"}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/albums/album123/tracks", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedTracks)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewAPIClient(mockAPIURL{baseURL: server.URL}, server.Client())
	tracks, err := client.GetAlbumTracks("test-token", "album123")

	assert.NoError(t, err)
	assert.Equal(t, "Album Track", tracks.Items[0].Name)
}

func TestGetUserSavedAlbums(t *testing.T) {
	expectedAlbums := &types.SavedAlbumsResponse{
		Items: []types.SavedAlbum{{Album: types.Album{Name: "Saved Album"}}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/me/albums", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedAlbums)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewAPIClient(mockAPIURL{baseURL: server.URL}, server.Client())
	albums, err := client.GetUserSavedAlbums("test-token")

	assert.NoError(t, err)
	assert.Equal(t, "Saved Album", albums.Items[0].Album.Name)
}

func TestGetUserPlaylists(t *testing.T) {
	expectedPlaylists := &types.UserPlaylistsResponse{
		Items: []types.Playlist{{Name: "My Playlist"}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/me/playlists/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedPlaylists)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewAPIClient(mockAPIURL{baseURL: server.URL}, server.Client())
	playlists, err := client.GetUserPlaylists("test-token")

	assert.NoError(t, err)
	assert.Equal(t, "My Playlist", playlists.Items[0].Name)
}

func TestGetArtistsTopTrack(t *testing.T) {
	expectedTracks := &types.ArtistsTopTrackResponse{
		Tracks: []types.Track{{Name: "Top track"}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/artists/artist123/top-tracks", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedTracks)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewAPIClient(mockAPIURL{baseURL: server.URL}, server.Client())
	tracks, err := client.GetArtistsTopTrack("test-token", "artist123")

	assert.NoError(t, err)
	assert.Equal(t, "Top track", tracks.Tracks[0].Name)
}

func TestGetSearchResults(t *testing.T) {
	expectedResults := &types.SearchResponse{
		Tracks: types.Paging[types.Track]{Items: []types.Track{{Name: "Found track"}}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/search", r.URL.Path)
		assert.Equal(t, "test-query", r.URL.Query().Get("q"))
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedResults)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewAPIClient(mockAPIURL{baseURL: server.URL}, server.Client())
	results, err := client.GetSearchResults("test-token", "test-query")

	assert.NoError(t, err)
	assert.Equal(t, "Found track", results.Tracks.Items[0].Name)
}

func TestSaveRemoveTrackForCurrentUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/me/tracks", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient(mockAPIURL{baseURL: server.URL}, server.Client())
	err := client.SaveRemoveTrackForCurrentUser("test-token", []string{"track1"}, false)

	assert.NoError(t, err)

	server.Close()
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client = NewAPIClient(mockAPIURL{baseURL: server.URL}, server.Client())
	err = client.SaveRemoveTrackForCurrentUser("test-token", []string{"track1"}, true)
	assert.NoError(t, err)
}

func TestGetFollowedArtist(t *testing.T) {
	expectedResponse := &types.UserFollowedArtistResponse{
		Artists: types.Artists{Items: []types.Artist{{Name: "Artist1"}}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/me/following", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedResponse)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewAPIClient(mockAPIURL{baseURL: server.URL}, server.Client())
	resp, err := client.GetFollowedArtist("test-token")

	assert.NoError(t, err)
	assert.Equal(t, "Artist1", resp.Artists.Items[0].Name)
}

func TestGetPlaylistItems(t *testing.T) {
	expectedResponse := &types.PlaylistItemsResponse{
		Items: []*types.PlaylistTrackObject{{Track: types.Track{Name: "Playlist Track"}}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/playlists/playlist123/tracks", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedResponse)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := NewAPIClient(mockAPIURL{baseURL: server.URL}, server.Client())
	resp, err := client.GetPlaylistItems("test-token", "playlist123", nil)

	assert.NoError(t, err)
	assert.Equal(t, "Playlist Track", resp.Items[0].Track.Name)
}
