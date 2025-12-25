# Contributing to clispot

Thank you for your interest in contributing to the Clispot CLI!

## Development Setup

1. **Fork and Clone**:
   - Fork the [clispot](https://github.com/kumneger0/clispot) repository.
   - Clone your fork locally:
     ```bash
     git clone https://github.com/YOUR_USERNAME/clispot.git
     cd clispot
     ```

2. **Prerequisites**:
   - Go 1.25 or higher.
   - `yt-dlp` and `ffmpeg` installed in your PATH.
   - **Spotify Developer Credentials**: You must have a Spotify Client ID and Client Secret exported as environment variables (`SPOTIFY_CLIENT_ID` and `SPOTIFY_CLIENT_SECRET`).

3. **Clone and Run**:
   ```bash
   go mod download
   go run main.go
   ```

## Pull Request Guidelines

- Follow standard Go formatting (`go fmt`).
- Provide a clear description of the changes in your PR.
- If you are adding a new feature, please update the documentation in the [clispot_docs](https://github.com/Kumneger0/clispot_docs) repository if applicable.

## Reporting Issues

Use the GitHub Issues tab to report bugs or suggest features. Please provide environment details (OS, Go version) when reporting bugs.
