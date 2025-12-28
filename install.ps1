$ErrorActionPreference = 'Stop'
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

# Variables
$repo = "kumneger0/clispot"
$clispotDir = "$env:LOCALAPPDATA\clispot"
$binDir = "$clispotDir\bin"
$exe = "$binDir\clispot.exe"

# Functions
function Write-Success {
    param($Message)
    Write-Host " > $Message" -ForegroundColor 'Green'
}

function Write-Info {
    param($Message)
    Write-Host " > $Message" -ForegroundColor 'Cyan'
}

function Test-Admin {
    $currentUser = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    $currentUser.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Checks
if (Test-Admin) {
    Write-Warning "The script is running as administrator. It is recommended to install clispot as a regular user."
    $choices = [System.Management.Automation.Host.ChoiceDescription[]] @(
        (New-Object System.Management.Automation.Host.ChoiceDescription '&Yes', 'Abort installation.'),
        (New-Object System.Management.Automation.Host.ChoiceDescription '&No', 'Resume installation.')
    )
    $choice = $Host.UI.PromptForChoice('Warning', 'Do you want to abort the installation process?', $choices, 0)
    if ($choice -eq 0) {
        Write-Host 'Installation aborted.' -ForegroundColor 'Yellow'
        exit
    }
}

# Determine Architecture
if ($env:PROCESSOR_ARCHITECTURE -eq 'AMD64') {
    $target = "Windows_x86_64"
}
elseif ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') {
    # Note: Check if ARM64 is actually supported in releases, usually it is x86_64 for Windows
    $target = "Windows_arm64"
}
else {
    Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
    exit
}

# Fetch Version
if ($v) {
    $version = $v.Replace('v', '')
}
else {
    Write-Info "Fetching latest version info..."
    $latestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
    $version = $latestRelease.tag_name.Replace('v', '')
}

Write-Info "Installing clispot v$version for $target..."

# Download
$archivePath = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "clispot.zip")
$downloadUrl = "https://github.com/$repo/releases/download/v$version/clispot_$($target).zip"

Write-Info "Downloading from $downloadUrl..."
Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -UseBasicParsing

# Install
if (-not (Test-Path $binDir)) {
    New-Item -Path $binDir -ItemType Directory -Force | Out-Null
}

Write-Info "Extracting to $binDir..."
Expand-Archive -Path $archivePath -DestinationPath $binDir -Force

# Cleanup
Remove-Item -Path $archivePath -Force -ErrorAction SilentlyContinue

# PATH Update
Write-Info "Adding clispot to PATH..."
$userPath = [Environment]::GetEnvironmentVariable('PATH', [EnvironmentVariableTarget]::User)
if ($userPath -notlike "*$binDir*") {
    $newPath = "$userPath;$binDir"
    [Environment]::SetEnvironmentVariable('PATH', $newPath, [EnvironmentVariableTarget]::User)
    $env:PATH = "$env:PATH;$binDir"
    Write-Success "clispot added to User PATH."
}
else {
    Write-Info "clispot is already in PATH."
}

Write-Success "clispot v$version was successfully installed!"
Write-Host "Restart your terminal to start using 'clispot'."
