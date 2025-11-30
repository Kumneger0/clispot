# clispot User Guide

**clispot** is a terminal-based Spotify client that lets you browse and discover music from your Spotify library and play it through YouTube

## Table of Contents

- [Features](#features)
  - [Current Features](#current-features)
  - [Technical Details](#technical-details)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [From GitHub Releases](#from-github-releases)
  - [Building from Source](#building-from-source)
- [Configuration](#configuration)
  - [Authentication](#authentication)
- [Usage](#usage)
  - [Starting the Application](#starting-the-application)
  - [Interface Layout](#interface-layout)
  - [Navigation](#navigation)
  - [Actions](#actions)
  - [Keyboard Shortcuts Reference](#keyboard-shortcuts-reference)
  - [Workflow Examples](#workflow-examples)
- [Troubleshooting](#troubleshooting)
  - [Common Issues](#common-issues)
- [Technical Details](#technical-details-1)
  - [How It Works](#how-it-works)
  - [File Locations](#file-locations)
  - [MPRIS Integration](#mpris-integration)
- [Contributing](#contributing)
- [License](#license)

## Features

### Current Features

* **Library Browsing**
  * Browse your saved playlists
  * View followed artists and their top tracks
  * Access featured playlists from Spotify
  * View playlist contents and artist discographies

* **Search**
  * Search Spotify's catalog for tracks, artists, and playlists
  * View search results in organized columns
  * Navigate between search result categories

* **Playback**
  * Play music through YouTube audio streams
  * Queue management for continuous playback
  * Visual progress indicator for currently playing track
  * Automatic track progression

* **Integration**
  * MPRIS2 support for media key control (play/pause, next, previous)
  * Works with system media key shortcuts

* **Caching**
  * Caches YouTube audio streams for played tracks so next time when you play same song it plays from cache.

* **Lyrics Display**
  * View lyrics for the currently playing music. To enable this feature, install the `clispot-lyrics` tool globally from [https://github.com/Kumneger0/clispot-lyrics](https://github.com/Kumneger0/clispot-lyrics). Once installed, clispot will automatically provide an option to open lyrics for the selected music.

### Technical Details

* Built with Go and the Bubble Tea TUI framework
* Uses Spotify Web API for music discovery
* Uses `yt-dlp` and `ffmpeg` for audio playback
* Stores authentication tokens securely in `~/.clispot/token.json`

## Installation

### Prerequisites

**Required:**
* Go 1.25 or higher (if building from source)
* [`yt-dlp`](https://github.com/yt-dlp/yt-dlp) - YouTube downloader for audio playback
* [`ffmpeg`](https://ffmpeg.org/) - Audio conversion

### From GitHub Releases

Download the latest release from the [GitHub releases page](https://github.com/kumneger0/clispot/releases). **Note:** Pre-built binaries are currently only available for Linux AMD64 architectures. Extract the archive for your architecture and place the `clispot` executable in a directory in your PATH (e.g., `/usr/local/bin`).

For other operating systems, please build from source (see [Building from Source](install.md)).

### Authentication

Before you can authenticate, you need to set up your Spotify API credentials. For detailed instructions on how to do this, please see the [Building from Source](install.md).

On first run, clispot will:
1. Start a local web server on port 9292
2. Open your default browser to Spotify's authorization page
3. Ask you to authorize the application
4. Save the authentication token to `~/.clispot/token.json`

The token is automatically refreshed when needed. You only need to authenticate once (unless you revoke it).

## Usage

### Starting the Application

Simply run:
```bash
clispot
```

The application will check for authentication and prompt you if needed.

**Command-line Options:**
* `-d, --debug-dir <path>` - Specify a directory where debug logs will be saved. The logs `ytstderr.log` and `ffstderr.log` will be created in this directory. If not specified, logs are saved in `~/.clispot/logs` directory.
* `--disable-cache` - Disable caching of YouTube audio streams. When this flag is used, audio streams will not be saved to disk. Defaults to `false` (caching is enabled by default).
* `--cookies-from-browser <browser_name[:profile]>` - Pass cookies from the specified browser to `yt-dlp`. This is useful for accessing age-restricted content or content that requires login.
* `--cookies <filepath>` - Pass cookies from a specified file to `yt-dlp`. The file should be in Netscape cookie file format.


Example:
```bash
clispot -d ~/logs/clispot
```

Example with cache disabled:
```bash
clispot --disable-cache
```

Example using cookies from a browser (e.g., Firefox):
```bash
clispot --cookies-from-browser firefox
```

Example using cookies from a file:
```bash
clispot --cookies ~/my_cookies.txt
```

### Interface Layout

clispot uses a three-panel layout:


**Left Panel (Sidebar):** Your Spotify library
* Saved playlists (if no featured playlists available)
* Featured playlists (by default, if available)
* Followed artists

**Center Panel (Main Content):**
* Displays tracks when you select a playlist or artist
* Shows search results (tracks, artists, playlists in columns)
* Changes based on current view mode

**Right Panel (Queue):**
* Current music queue
* Tracks are added here when you play music
* Use `n` and `b` to navigate through the queue

**Bottom (Player Controls):**
* Shows currently playing track with progress bar
* Playback control buttons with keyboard shortcuts

### Navigation

**Moving Between Panels:**
* `Tab` - Cycle focus forward through panels (Sidebar → Main → Queue → repeat)
* `Shift+Tab` - Cycle focus backward

**Within Lists:**
* `j` or `↓` - Move down
* `k` or `↑` - Move up
* Arrow keys also work for navigation

**Focusing Search:**
* `Ctrl+K` - Focus the search bar

### Actions

**Search:**
1. Press `Ctrl+K` to focus the search bar
2. Type your query
3. Press `Enter` to search
4. Results appear in three columns: Tracks, Artists, Playlists
5. Use `Tab` to navigate between result columns

**Liking and Unliking Songs:**
* `l` - Like or unlike a song. Liked songs can be found in the "Liked Songs" playlist in the sidebar.

**Managing the Queue:**
* `a` - Add a song to the queue.
* `d` - Remove a song from the queue.

**Playing Music:**
* `Enter` (when focused on):
  * **Sidebar:** Load playlist/artist tracks into main view
  * **Main View:** Play selected track and add current view to queue
  * **Queue:** Play selected track
  * **Search Results:**
    * **Track:** Play the track directly
    * **Artist:** Load artist's top tracks
    * **Playlist:** Load playlist tracks

**Playback Controls:**
* `Space` - Toggle play/pause
* `n` - Play next track in queue(the player section needs to be focused here)
* `b` - Play previous track in queue(the player section needs to be focused here)

**Exiting:**
* `q` or `Ctrl+C` - Quit the application

### Keyboard Shortcuts Reference

| Key | Action |
|-----|--------|
| `Tab` | Cycle focus between panels |
| `Ctrl+K` | Focus search bar |
| `j` / `↓` | Move down in list |
| `k` / `↑` | Move up in list |
| `l` | Like/Unlike a song |
| `a` | Add song to queue |
| `d` | Remove song from queue |
| `Enter` | Context-dependent action (play/load) |
| `Space` | Play/pause current track |
| `n` | Next track |
| `b` | Previous track |
| `q` / `Ctrl+C` | Quit |
| `Ctrl+L` | Open Lyrics |

### Workflow Examples

**Playing a Playlist:**
1. Navigate to the sidebar (starts focused here)
2. Use `j`/`k` to select a playlist
3. Press `Enter` to load tracks
4. Select a track in the main view
5. Press `Enter` to start playback

**Searching and Playing:**
1. Press `Ctrl+K` to focus search
2. Type your query (e.g., "Radiohead")
3. Press `Enter`
4. Use `Tab` to navigate to the Tracks column
5. Select a track and press `Enter` to play

**Browsing an Artist:**
1. In the sidebar, select an artist (if in your followed artists)
2. Press `Enter` to load their top tracks
3. Select a track and press `Enter` to play
4. Or search for an artist, select from search results, and press `Enter`

## Troubleshooting

### Common Issues

**Music doesn't play**
* Check that `yt-dlp` and `ffmpeg` are installed and in your PATH
* Verify your internet connection
* Check debug log files: `ytstderr.log` and `ffstderr.log`
  * If you specified a debug directory with `-d`, logs are in that directory
  * Otherwise, logs are in `~/.clispot/logs` directory

**Authentication fails**
* Ensure port 9292 is not blocked by a firewall
* Check that your browser can open automatically (or manually open the URL shown)

## Technical Details

### How It Works

1. **Music Discovery:** clispot uses the Spotify Web API to browse your library, search, and get track metadata
2. **Playback:** When you play a track, clispot:
   - Searches YouTube using `yt-dlp` with the track name and artist
   - Downloads the audio stream
   - Converts it using `ffmpeg` to PCM format
   - Plays it using the Oto audio library

### File Locations

* **Config:** `~/.clispot/token.json` - Stores Spotify authentication token
* **Debug logs:** 
  * `ytstderr.log` - YouTube downloader error logs
  * `ffstderr.log` - FFmpeg conversion error logs
  * Log location depends on the `-d` flag:
    * If `-d <path>` is specified: logs are saved in that directory
    * If not specified: logs are saved in `~/.clispot/logs` directory

These logs can be useful for troubleshooting playback issues and reporting bugs.

### Caching Behavior

When a track is played, clispot attempts to cache its audio stream.
*   If you skip a song during its *first-time playback* (i.e., while it's being downloaded and cached), the partially downloaded cache file will be removed.
*   If a song has already been *fully downloaded and cached*, and you skip it during a subsequent playback (when it's playing from the cache), the existing cache file will **not** be removed.

This ensures that only incomplete or interrupted cache files are cleaned up, while fully cached tracks remain available for future playback.

### MPRIS Integration

clispot implements MPRIS2, allowing integration with:
* Desktop environments that support media keys
* Media controllers and widgets
* Keyboard media keys (if configured to use MPRIS)

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

If you find bugs or have feature requests, please [open an issue](https://github.com/kumneger0/clispot/issues/new) on GitHub.

## License

This project is licensed under the MIT License. See the [LICENSE](../LICENSE) file for details.
