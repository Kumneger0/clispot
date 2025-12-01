package lyrics

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	lyricsServerBaseURL = "http://localhost:5222"
)

type Req struct {
	Title   string   `json:"title"`
	Artists []string `json:"artists"`
	Album   string   `json:"album"`
}

type Match struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	URL    string `json:"url"`
	Score  int    `json:"score"`
}

type Response struct {
	Match  *Match  `json:"match"`
	Lyrics *string `json:"lyrics"`
}

var LyricsErrors = []string{
	"We searched everywhere. Found nothing. Pain.",
	"If the lyrics exist, they’re hiding from you specifically.",
	"No lyrics. Just vibes.",
	"We tried. They said no.",
	"Lyrics is missing. I’m tired. Try again later.",
	"You'll have to guess the lyrics for this one",
}

func GetMusicLyrics(title string, artists []string, album string) (*Response, error) {
	reqBody := Req{
		Title:   title,
		Artists: artists,
		Album:   album,
	}
	jsonByte, err := json.Marshal(reqBody)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	body := bytes.NewReader(jsonByte)
	req, err := http.NewRequest("POST", lyricsServerBaseURL, body)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get the lyrics")
	}
	var lyricsResponse *Response
	err = json.NewDecoder(resp.Body).Decode(&lyricsResponse)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return lyricsResponse, nil
}

func IsLyricsServerRunning() (bool, error) {
	req, err := http.NewRequest("GET", lyricsServerBaseURL, nil)
	if err != nil {
		slog.Error(err.Error())
		return false, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return false, err
	}
	defer resp.Body.Close()
	isServerRunning := resp.StatusCode == http.StatusOK
	return isServerRunning, nil
}

func StartLyricsServer() (*os.Process, error) {
	cmd := exec.Command("clispot-lyrics")
	err := cmd.Start()
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	time.Sleep(time.Second * 2) // wait for the server to start
	return cmd.Process, nil
}
