#!/usr/bin/env pwsh
# setup.ps1 - Install all dependencies for Embalses
# Run as Administrator

$ErrorActionPreference = "Stop"

function Test-Command {
    param([string]$Command)
    try { $null = Get-Command $Command -ErrorAction Stop; return $true } catch { return $false }
}

function Write-Step {
    param([string]$Message)
    Write-Host "`n>>> $Message" -ForegroundColor Cyan
}

function Write-Ok {
    param([string]$Message)
    Write-Host "    OK: $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "    WARN: $Message" -ForegroundColor Yellow
}

$installDir = "C:\Tools"
$envFile = "$env:USERPROFILE\.embalses-env.ps1"

# Ensure install directory exists
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# --- Install Go ---
Write-Step "Checking Go..."
if (Test-Command "go") {
    $goVersion = (go version) -replace "go version ", ""
    Write-Ok "Go already installed: $goVersion"
} else {
    Write-Warn "Go not found. Downloading Go 1.23..."
    $goUrl = "https://go.dev/dl/go1.23.6.windows-amd64.zip"
    $goZip = "$env:TEMP\go.zip"
    $goDir = "$installDir\go"

    Invoke-WebRequest -Uri $goUrl -OutFile $goZip -UseBasicParsing
    Expand-Archive -Path $goZip -DestinationPath $goDir -Force
    Remove-Item $goZip

    # Add to PATH via environment variable
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$goDir\bin*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$goDir\bin", "User")
        Write-Ok "Added Go to PATH"
    }
    $env:Path += ";$goDir\bin"
    Write-Ok "Go installed at $goDir"
}

# --- Install Node.js ---
Write-Step "Checking Node.js..."
if (Test-Command "node") {
    $nodeVersion = (node -v)
    Write-Ok "Node.js already installed: $nodeVersion"
} else {
    Write-Warn "Node.js not found. Downloading Node.js 22 LTS..."
    $nodeUrl = "https://nodejs.org/dist/v22.14.0/node-v22.14.0-win-x64.zip"
    $nodeZip = "$env:TEMP\node.zip"
    $nodeDir = "$installDir\nodejs"

    Invoke-WebRequest -Uri $nodeUrl -OutFile $nodeZip -UseBasicParsing
    Expand-Archive -Path $nodeZip -DestinationPath $installDir -Force
    Remove-Item $nodeZip

    # Rename extracted folder to consistent name
    $extracted = Get-ChildItem "$installDir" -Filter "node-v*" -Directory | Select-Object -First 1
    if ($extracted) {
        Rename-Item $extracted.FullName $nodeDir -Force
    }

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$nodeDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$nodeDir", "User")
        Write-Ok "Added Node.js to PATH"
    }
    $env:Path += ";$nodeDir"
    Write-Ok "Node.js installed at $nodeDir"
}

# Verify npm
if (Test-Command "npm") {
    $npmVersion = (npm -v)
    Write-Ok "npm available: $npmVersion"
} else {
    Write-Warn "npm not found after Node.js install. Try restarting PowerShell."
}

# --- Install Docker Desktop ---
Write-Step "Checking Docker..."
if (Test-Command "docker") {
    $dockerVersion = (docker --version)
    Write-Ok "Docker already installed: $dockerVersion"
} else {
    Write-Warn "Docker not found. Downloading Docker Desktop..."
    Write-Host "    NOTE: Docker Desktop requires a manual install. Downloading installer..."
    $dockerUrl = "https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe"
    $dockerInstaller = "$env:TEMP\DockerDesktopInstaller.exe"
    Invoke-WebRequest -Uri $dockerUrl -OutFile $dockerInstaller -UseBasicParsing
    Write-Host "    Docker Desktop installer downloaded to: $dockerInstaller"
    Write-Host "    Please run it manually and follow the installation wizard."
    Write-Host "    After installation, restart your computer and re-run this script."
}

# --- Install frontend dependencies ---
Write-Step "Installing frontend dependencies..."
$webDir = "C:\Users\whala\git\embalses\web"
if (Test-Path $webDir) {
    Push-Location $webDir
    try {
        npm install
        Write-Ok "Frontend dependencies installed"
    } finally {
        Pop-Location
    }
} else {
    Write-Warn "Web directory not found at $webDir"
}

# --- Write environment helper script ---
Write-Step "Creating environment helper script..."
$envContent = @"
# Embalses environment setup
# Source this file in PowerShell: . `$env:USERPROFILE\.embalses-env.ps1

`$goDir = "$installDir\go\bin"
`$nodeDir = "$installDir\nodejs"

if (Test-Path `$goDir) { `$env:Path += ";`$goDir" }
if (Test-Path `$nodeDir) { `$env:Path += ";`$nodeDir" }

`$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/embalses?sslmode=disable"
`$env:VITE_API_KEY = "test-key-123"
"@

$envContent | Out-File -FilePath $envFile -Encoding UTF8
Write-Ok "Environment helper created at $envFile"

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "`nNext steps:"
Write-Host "  1. Restart PowerShell (to reload PATH)"
Write-Host "  2. Run: . $envFile"
Write-Host "  3. Then run: C:\Users\whala\git\embalses\scripts\start.ps1"
Write-Host "`nIf Docker was not installed, install Docker Desktop from the downloaded installer first."
