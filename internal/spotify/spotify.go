package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kumneger0/clispot/internal/types"
)

type UserSavedTracksListItem struct {
	Name string
}

func (userSavedTracksListItem UserSavedTracksListItem) Title() string {
	return userSavedTracksListItem.Name
}

func (userSavedTracksListItem UserSavedTracksListItem) FilterValue() string {
	return userSavedTracksListItem.Name
}

type UserTopItem string

const (
	artist UserTopItem = "artists"
	track  UserTopItem = "tracks"
)

const (
	//playlist base url
	playlistBase = "https://api.spotify.com/v1/me/playlists/"
	//tracks base url
	tracksBase = "https://api.spotify.com/v1/tracks/"
	//user profile base url
	userProfileBase     = "https://api.spotify.com/v1/me/"
	artistsURL          = "https://api.spotify.com/v1/artists/"
	followedArtistURL   = "https://api.spotify.com/v1/me/following?type=artist"
	userSavedTrackURL   = "https://api.spotify.com/v1/me/tracks"
	checkUserSavedTrack = "https://api.spotify.com/v1/me/tracks/contains?ids="
)

var (
	userTopItemsBaseURL = "https://api.spotify.com/v1/me/top/"
	searchBaseURL       = "https://api.spotify.com/v1/search"
)

type APIURLS interface {
	GetPlaylistBaseURL() string
	GetTrackBaseURL() string
	GetUserProfileBaseURL() string
	GetPlaylistItems(playlistID string) string
	GetArtistsURL() string
	GetFollowedArtistURL() string
	GetUserTopItems(itemType UserTopItem) string
	GetArtistsTopTrackURL(id string) string
	GetSearchURL(q string) string
	GetUserSavedTrackURL() string
	GetCheckUserSavedTrackURL() string
}

type apiURL struct{}

var APIURL APIURLS = apiURL{}

func (a apiURL) GetCheckUserSavedTrackURL() string {
	return checkUserSavedTrack
}

func (a apiURL) GetUserSavedTrackURL() string {
	return userSavedTrackURL
}

func (a apiURL) GetPlaylistBaseURL() string {
	return playlistBase
}
func (a apiURL) GetSearchURL(q string) string {
	searchType := "track,artist,playlist"
	limit := 30
	market := "US"
	offset := 0
	searchParams := url.Values{}
	searchParams.Add("q", q)
	searchParams.Add("type", searchType)
	searchParams.Add("limit", strconv.Itoa(limit))
	searchParams.Add("market", market)
	searchParams.Add("offset", strconv.Itoa(offset))
	return searchBaseURL + "?" + searchParams.Encode()
}

func (a apiURL) GetTrackBaseURL() string {
	return tracksBase
}

func (a apiURL) GetUserProfileBaseURL() string {
	return userProfileBase
}

func (a apiURL) GetArtistsURL() string {
	return artistsURL
}
func (a apiURL) GetUserTopItems(itemType UserTopItem) string {
	return userTopItemsBaseURL + string(itemType)
}
func (a apiURL) GetFollowedArtistURL() string {
	return followedArtistURL
}
func (a apiURL) GetArtistsTopTrackURL(id string) string {
	return "https://api.spotify.com/v1/artists/" + id + "/top-tracks"
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

func makeRequest(method string, urlToMakeRequestTo string, authorizationHeader string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, urlToMakeRequestTo, body)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	req.Header.Add("Authorization", authorizationHeader)
	return http.DefaultClient.Do(req)
}

func CheckUserSavedTrack(accessToken string, trackID string) ([]bool, error) {
	authorizationHeader := "Bearer " + accessToken

	resp, err := makeRequest("GET", APIURL.GetCheckUserSavedTrackURL()+trackID, authorizationHeader, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var contains []bool
	if err := json.NewDecoder(resp.Body).Decode(&contains); err != nil {
		return nil, err
	}

	return contains, nil
}

type SaveTrackForCurrentUserRequest struct {
	IDs []string `json:"ids"`
}

func SaveRemoveTrackForCurrentUser(accessToken string, trackIDs []string, isRemove bool) error {
	authorizationHeader := "Bearer " + accessToken

	reqBody := SaveTrackForCurrentUserRequest{
		IDs: trackIDs,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	body := bytes.NewReader(reqBodyBytes)

	var method string
	if isRemove {
		method = "DELETE"
	} else {
		method = "PUT"
	}

	resp, err := makeRequest(method, userSavedTrackURL, authorizationHeader, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func GetUserSavedTracks(accessToken string) (*types.UserSavedTracks, error) {
	authorizationHeader := "Bearer " + accessToken
	params := url.Values{}
	params.Add("limit", "30")
	resp, err := makeRequest("GET", APIURL.GetUserSavedTrackURL()+"?"+params.Encode(), authorizationHeader, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var userSavedTracks *types.UserSavedTracks
	if err := json.NewDecoder(resp.Body).Decode(&userSavedTracks); err != nil {
		return nil, err
	}

	return userSavedTracks, nil
}

func Search(accessToken string, query string) (*types.SearchResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", APIURL.GetSearchURL(query), authorizationHeader, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var searchResponse *types.SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, err
	}
	return searchResponse, nil
}

func GetArtistsTopTrackURL(accessToken string, artistID string) (*types.ArtistTopTracks, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", APIURL.GetArtistsTopTrackURL(artistID), authorizationHeader, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var artistTopTracks *types.ArtistTopTracks
	if err := json.NewDecoder(resp.Body).Decode(&artistTopTracks); err != nil {
		return nil, err
	}
	return artistTopTracks, nil
}

func GetUserTopItems(accessToken string) (*types.UserTopItemsResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", APIURL.GetUserTopItems(track), authorizationHeader, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var userFollowedArtist *types.UserTopItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&userFollowedArtist); err != nil {
		return nil, err
	}
	return userFollowedArtist, nil
}

func GetFollowedArtist(accessToken string) (*types.UserFollowedArtistResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", APIURL.GetFollowedArtistURL(), authorizationHeader, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var userFollowedArtist *types.UserFollowedArtistResponse
	if err := json.NewDecoder(resp.Body).Decode(&userFollowedArtist); err != nil {
		return nil, err
	}
	return userFollowedArtist, nil
}

func GetUserPlaylists(accessToken string) (*types.UserPlaylistsResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", playlistBase, authorizationHeader, nil)
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

func GetPlaylistItems(playlistID string, accessToken string) (*types.PlaylistItemsResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	playlistItemsURL := APIURL.GetPlaylistItems(playlistID)
	resp, err := makeRequest("GET", playlistItemsURL, authorizationHeader, nil)
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
	resp, err := makeRequest("GET", userProfileBase, authorizationHeader, nil)
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

func GetTrack(trackID string, accessToken string) (*types.Track, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", tracksBase, authorizationHeader, nil)
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
	var track types.Track
	if err := json.NewDecoder(resp.Body).Decode(&track); err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return &track, nil
}
