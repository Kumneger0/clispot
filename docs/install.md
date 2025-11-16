# Building from Source

## Prerequisites

Before building clispot from source, ensure you have the following prerequisites installed:

- [Go](https://go.dev/dl/) version 1.25 or higher
- Git

## Installation Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/kumneger0/clispot.git
   cd clispot
   ```

2. Create a Spotify Application:
   - Go to the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/applications).
   - Log in with your Spotify account.
   - Click on "Create an app" and fill in the required details.
   - Once the app is created, navigate to "Settings".
   - Under "Redirect URIs", add `http://127.0.0.1:9292`. This is the redirect URI used by clispot for authentication.
   - Save the changes.
   - Note down your `Client ID` and `Client Secret`.

3. Build the project with your Spotify credentials:
   ```bash
   go build -ldflags="-X main.spotifyClientID=<your-client-id> -X main.spotifyClientSecret=<your-client-secret>" main.go
   ```

If you encounter any issues during installation:  
For additional help, please [open an issue](https://github.com/kumneger0/clispot/issues/new) on GitHub.
