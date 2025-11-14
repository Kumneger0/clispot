package spotify

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kumneger0/clispot/internal/types"
)

var (
	authBaseURL = "https://accounts.spotify.com/authorize"
	scope       = "user-read-private user-read-email playlist-read-private user-library-read user-top-read user-follow-read"
	tokenRL     = "https://accounts.spotify.com/api/token"
	redirectURL = "http://127.0.0.1:9292/callback"
)

type Secret struct {
	ClientID     string
	ClientSecret string
}

func Authenticate(spotifyClientID, spotifyClientSecret string) (*types.UserTokenInfo, error) {
	var state = generateRandomString(16)
	logAuthenticationURL(state, spotifyClientID)

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":9292",
		Handler: mux,
	}

	var userToken *types.UserTokenInfo

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		error := r.URL.Query().Get("error")
		if error != "" {
			slog.Error(error)
			fmt.Fprintf(w, "Error: %s\n", error)
			return
		}
		if state != r.URL.Query().Get("state") {
			slog.Error("State mismatch")
			fmt.Fprintf(w, "State mismatch")
			return
		}
		code := r.URL.Query().Get("code")

		formData := url.Values{}
		formData.Set("code", code)
		formData.Set("redirect_uri", redirectURL)
		formData.Set("grant_type", "authorization_code")

		token, err := getToken(formData.Encode(), spotifyClientID, spotifyClientSecret)
		if err != nil {
			slog.Error(err.Error())
			fmt.Fprintf(w, "Error: %s\n", err)
			return
		}
		userToken = token
		fmt.Fprintf(w, "go back to your terminal")
		go func() {
			server.Close()
		}()
	})

	var wg sync.WaitGroup
	wg.Go(func() {
		fmt.Println("🚀 Server started at http://localhost:9292")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("❌ Server error: %v\n", err)
		}
	})
	wg.Wait()
	return userToken, nil
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result string
	for range length {
		result += string(charset[rand.Intn(len(charset))])
	}
	return result
}

func logAuthenticationURL(state string, spotifyClientID string) {
	params := url.Values{}
	params.Set("client_id", spotifyClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", redirectURL)
	params.Set("scope", scope)
	params.Set("state", state)
	fullURLToRedirect := fmt.Sprintf("%s?%s", authBaseURL, params.Encode())
	fmt.Println("opening the link using default browser")
	_ = openFileInDefaultApp(fullURLToRedirect)
	fmt.Println("click the following link to authenticate ")
	fmt.Println(fullURLToRedirect)
}

func openFileInDefaultApp(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		//fuck you windows 🖕 i hate you
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

func getToken(encodedFormData string, spotifyClientID, spotifyClientSecret string) (*types.UserTokenInfo, error) {
	authString := spotifyClientID + ":" + spotifyClientSecret
	base64Auth := base64.StdEncoding.EncodeToString([]byte(authString))

	req, err := http.NewRequest("POST", tokenRL, strings.NewReader(encodedFormData))
	if err != nil {
		return &types.UserTokenInfo{}, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+base64Auth)

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return &types.UserTokenInfo{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return &types.UserTokenInfo{}, fmt.Errorf("Unexpected status code: %d\n", res.StatusCode)
	}

	jsonBytes, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error(err.Error())
		return &types.UserTokenInfo{}, err
	}

	var tokenResponse types.UserTokenInfo

	if err := json.Unmarshal(jsonBytes, &tokenResponse); err != nil {
		slog.Error(err.Error())
		return &types.UserTokenInfo{}, err
	}

	if tokenResponse.AccessToken == "" {
		slog.Error("access token not found in response")
		return &types.UserTokenInfo{}, errors.New("access token not found in response")
	}
	err = saveUserCredentials(tokenResponse)
	if err != nil {
		slog.Error(err.Error())
		return &types.UserTokenInfo{}, err
	}
	return &tokenResponse, nil
}

func RefreshToken(refreshToken string, spotifyClientID, spotifyClientSecret string) (*types.UserTokenInfo, error) {
	formData := url.Values{}
	formData.Set("grant_type", "refresh_token")
	formData.Set("refresh_token", refreshToken)

	token, err := getToken(formData.Encode(), spotifyClientID, spotifyClientSecret)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	if token.RefreshToken == "" {
		token.RefreshToken = refreshToken
	}

	err = saveUserCredentials(*token)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return token, nil
}

func saveUserCredentials(userCredentials types.UserTokenInfo) error {
	currentUser, err := user.Current()
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("error getting current user: %w", err)
	}
	homeDir := currentUser.HomeDir

	dirPath := filepath.Join(homeDir, ".clispot")
	filePath := filepath.Join(dirPath, "token.json")

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			slog.Error(err.Error())
			return fmt.Errorf("error creating directory %s: %w", dirPath, err)
		}
	} else if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("error checking directory %s: %w", dirPath, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("error creating file %s: %w", filePath, err)
	}

	userCredentials.ExpiresAt = time.Now().Add(time.Duration(userCredentials.ExpiresIn) * time.Second).Unix()

	defer file.Close()
	jsonBytes, err := json.Marshal(userCredentials)
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("marshal error: %w", err)
	}

	_, err = file.Write(jsonBytes)
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}

func ReadUserCredentials() (*types.UserTokenInfo, error) {
	currentUser, err := user.Current()
	if err != nil {
		return &types.UserTokenInfo{}, err
	}
	homeDir := currentUser.HomeDir
	dirPath := filepath.Join(homeDir, ".clispot")
	clispotCredentialPath := filepath.Join(dirPath, "token.json")

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		slog.Error(err.Error())
		return &types.UserTokenInfo{}, errors.New("user credentials not found")
	}
	if _, err := os.Stat(clispotCredentialPath); err != nil {
		slog.Error(err.Error())
		return &types.UserTokenInfo{}, errors.New("user credentials not found")
	}

	file, err := os.Open(clispotCredentialPath)
	if err != nil {
		slog.Error(err.Error())
		return &types.UserTokenInfo{}, err
	}
	defer file.Close()

	jsonBytes, err := io.ReadAll(file)
	if err != nil {
		slog.Error(err.Error())
		return &types.UserTokenInfo{}, err
	}

	var tokenResponse types.UserTokenInfo
	if err := json.Unmarshal(jsonBytes, &tokenResponse); err != nil {
		slog.Error(err.Error())
		return &types.UserTokenInfo{}, err
	}

	return &tokenResponse, nil
}
