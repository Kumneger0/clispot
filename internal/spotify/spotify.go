package spotify

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kumneger0/clispot/internal/types"
)

const (
	//playlist base url
	playlistBase = "https://api.spotify.com/v1/me/playlists/"
	//featured playlist endpoint
	featuredPlaylistBase = "https://api.spotify.com/v1/browse/featured-playlists"
	//tracks base url
	tracksBase = "https://api.spotify.com/v1/tracks/"
	//user profile base url
	userProfileBase = "https://api.spotify.com/v1/me/"
)

type APIURLS interface {
	GetPlaylistBaseURL() string
	GetFeaturedPlayListURL() string
	GetTrackBaseURL() string
	GetUserProfileBaseURL() string
	GetPlaylistItems(playlistID string) string
}

type apiURL struct{}

var APIURL APIURLS = apiURL{}

func (a apiURL) GetPlaylistBaseURL() string {
	return playlistBase
}
func (a apiURL) GetFeaturedPlayListURL() string {
	return featuredPlaylistBase
}
func (a apiURL) GetTrackBaseURL() string {
	return tracksBase
}
func (a apiURL) GetUserProfileBaseURL() string {
	return userProfileBase
}

func (a apiURL) GetPlaylistItems(playlistID string) string {
	base := "https://api.spotify.com/v1/playlists/"
	return base + playlistID + "/tracks"
}

type Decoder struct {
	Playlist         types.UserPlaylistsResponse
	FeaturedPlaylist types.FeaturedPlaylistsResponse
	UserProfile      types.SpotifyUserProfile
	Track            types.SpotifyUser
}

func makeRequest(method string, url string, authorizationHeader string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	req.Header.Add("Authorization", authorizationHeader)
	return http.DefaultClient.Do(req)
}

func GetUserPlaylists(accessToken string) (*types.UserPlaylistsResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", playlistBase, authorizationHeader)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var playlists *types.UserPlaylistsResponse
	if err := json.NewDecoder(resp.Body).Decode(&playlists); err != nil {
		return nil, err
	}
	return playlists, nil
}

func GetFeaturedPlaylist(accessToken string) (*types.FeaturedPlaylistsResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", featuredPlaylistBase, authorizationHeader)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		slog.Error(errMsg)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var featuredPlaylist *types.FeaturedPlaylistsResponse
	if err := json.NewDecoder(resp.Body).Decode(&featuredPlaylist); err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return featuredPlaylist, nil
}

func GetPlaylistItems(playlistID string, accessToken string) (*types.PlaylistItemsResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	playlistItemsURL := APIURL.GetPlaylistItems(playlistID)
	resp, err := makeRequest("GET", playlistItemsURL, authorizationHeader)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		slog.Error(errMsg)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var featuredPlaylist *types.PlaylistItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&featuredPlaylist); err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return featuredPlaylist, nil
}

func GetUserProfile(accessToken string) (*types.SpotifyUserProfile, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", userProfileBase, authorizationHeader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		slog.Error(errMsg)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var userProfile types.SpotifyUserProfile
	if err := json.NewDecoder(resp.Body).Decode(&userProfile); err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return &userProfile, nil
}

func GetTrack(trackID string, accessToken string) (*types.SpotifyTrack, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", tracksBase, authorizationHeader)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		slog.Error(errMsg)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var track types.SpotifyTrack
	if err := json.NewDecoder(resp.Body).Decode(&track); err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return &track, nil
}
