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
	track UserTopItem = "tracks"
)

const (
	playlistBase           = "https://api.spotify.com/v1/me/playlists/"
	tracksBase             = "https://api.spotify.com/v1/tracks/"
	userProfileBase        = "https://api.spotify.com/v1/me/"
	artistsURL             = "https://api.spotify.com/v1/artists/"
	followedArtistURL      = "https://api.spotify.com/v1/me/following?type=artist"
	userSavedTrackURL      = "https://api.spotify.com/v1/me/tracks"
	checkUserSavedTrack    = "https://api.spotify.com/v1/me/tracks/contains?ids="
	userSavedAlbumsBaseURL = "https://api.spotify.com/v1/me/albums"
	albumTracksURL         = "https://api.spotify.com/v1/albums"
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
	GetUserSavedAlbumsBaseURL() string
	GetAlbumTracksURL(albumID string) string
	GetUserPlaylistsBaseURL() string
	GetTrackURL(trackID string) string
}

type apiURL struct{}

func NewAPIURL() APIURLS {
	return apiURL{}
}

func (a apiURL) GetTrackURL(trackID string) string {
	return tracksBase + trackID
}

func (a apiURL) GetUserPlaylistsBaseURL() string {
	return playlistBase
}

func (a apiURL) GetUserSavedAlbumsBaseURL() string {
	return userSavedAlbumsBaseURL
}

func (a apiURL) GetAlbumTracksURL(albumID string) string {
	return albumTracksURL + "/" + albumID + "/tracks"
}

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
	searchType := "track,artist,playlist,album"
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

type APIClient interface {
	GetAlbumTracks(accessToken string, albumID string) (*types.AlbumTracksResponse, error)
	GetTrack(accessToken string, trackID string) (*types.Track, error)
	GetUserPlaylists(accessToken string) (*types.UserPlaylistsResponse, error)
	GetUserSavedAlbums(accessToken string) (*types.SavedAlbumsResponse, error)
	CheckUserSavedTrack(accessToken string, trackID string) ([]bool, error)
	GetUserProfile(accessToken string) (*types.SpotifyUserProfile, error)
	GetArtistsTopTrack(accessToken string, artistID string) (*types.ArtistsTopTrackResponse, error)
	GetSearchResults(accessToken string, query string) (*types.SearchResponse, error)
	GetUserTopItems(accessToken string) (*types.UserTopItemsResponse, error)
	GetUserSavedTracks(accessToken string) (*types.UserSavedTracks, error)
	SaveRemoveTrackForCurrentUser(accessToken string, trackIDs []string, isRemove bool) error
	GetFollowedArtist(accessToken string) (*types.UserFollowedArtistResponse, error)
}

type APIClientImpl struct {
	apiURL APIURLS
}

func NewAPIClient(apiURL APIURLS) *APIClientImpl {
	return &APIClientImpl{
		apiURL: apiURL,
	}
}

func (a *APIClientImpl) GetAlbumTracks(accessToken string, albumID string) (*types.AlbumTracksResponse, error) {
	authorizationHeader := "Bearer " + accessToken

	resp, err := makeRequest("GET", a.apiURL.GetAlbumTracksURL(albumID), authorizationHeader, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var albumTracks types.AlbumTracksResponse
	if err := json.NewDecoder(resp.Body).Decode(&albumTracks); err != nil {
		return nil, err
	}

	return &albumTracks, nil
}

func (a *APIClientImpl) GetUserSavedAlbums(accessToken string) (*types.SavedAlbumsResponse, error) {
	authorizationHeader := "Bearer " + accessToken

	resp, err := makeRequest("GET", a.apiURL.GetUserSavedAlbumsBaseURL(), authorizationHeader, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var savedAlbums types.SavedAlbumsResponse
	if err := json.NewDecoder(resp.Body).Decode(&savedAlbums); err != nil {
		return nil, err
	}

	return &savedAlbums, nil
}

func (a *APIClientImpl) CheckUserSavedTrack(accessToken string, trackID string) ([]bool, error) {
	authorizationHeader := "Bearer " + accessToken

	resp, err := makeRequest("GET", a.apiURL.GetCheckUserSavedTrackURL()+trackID, authorizationHeader, nil)
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

func (a *APIClientImpl) SaveRemoveTrackForCurrentUser(accessToken string, trackIDs []string, isRemove bool) error {
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

	resp, err := makeRequest(method, a.apiURL.GetUserSavedTrackURL(), authorizationHeader, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (a *APIClientImpl) GetUserSavedTracks(accessToken string) (*types.UserSavedTracks, error) {
	authorizationHeader := "Bearer " + accessToken
	params := url.Values{}
	params.Add("limit", "30")
	resp, err := makeRequest("GET", a.apiURL.GetUserSavedTrackURL()+"?"+params.Encode(), authorizationHeader, nil)
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

func (a *APIClientImpl) Search(accessToken string, query string) (*types.SearchResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", a.apiURL.GetSearchURL(query), authorizationHeader, nil)
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

func (a *APIClientImpl) GetArtistsTopTrackURL(accessToken string, artistID string) (*types.ArtistsTopTrackResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", a.apiURL.GetArtistsTopTrackURL(artistID), authorizationHeader, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var artistTopTracks *types.ArtistsTopTrackResponse
	if err := json.NewDecoder(resp.Body).Decode(&artistTopTracks); err != nil {
		return nil, err
	}
	return artistTopTracks, nil
}

func (a *APIClientImpl) GetUserTopItems(accessToken string) (*types.UserTopItemsResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", a.apiURL.GetUserTopItems(track), authorizationHeader, nil)
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

func (a *APIClientImpl) GetFollowedArtist(accessToken string) (*types.UserFollowedArtistResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", a.apiURL.GetFollowedArtistURL(), authorizationHeader, nil)
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

func (a *APIClientImpl) GetUserPlaylists(accessToken string) (*types.UserPlaylistsResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", a.apiURL.GetUserPlaylistsBaseURL(), authorizationHeader, nil)
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

func (a *APIClientImpl) GetPlaylistItems(playlistID string, accessToken string) (*types.PlaylistItemsResponse, error) {
	authorizationHeader := "Bearer " + accessToken
	playlistItemsURL := a.apiURL.GetPlaylistItems(playlistID)
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

func (a *APIClientImpl) GetUserProfile(accessToken string) (*types.SpotifyUserProfile, error) {
	authorizationHeader := "Bearer " + accessToken
	resp, err := makeRequest("GET", a.apiURL.GetUserProfileBaseURL(), authorizationHeader, nil)
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
	var userProfile types.SpotifyUserProfile
	if err := json.NewDecoder(resp.Body).Decode(&userProfile); err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return &userProfile, nil
}

func (a *APIClientImpl) GetTrack(trackID string, accessToken string) (*types.Track, error) {
	authorizationHeader := "Bearer " + accessToken
	trackURL := a.apiURL.GetTrackURL(trackID)
	resp, err := makeRequest("GET", trackURL, authorizationHeader, nil)
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
