#!/usr/bin/env pwsh
# scripts/start.ps1 — Start the Embalses SQLite stack and Vite frontend on free ports.
# Usage: .\scripts\start.ps1 [apiBasePort] [frontendBasePort]
#
# Defaults: API 8082, frontend 5173. If a port is taken, the next free one is chosen.
# The script uses the existing data/embalses.db (368+ SNCZI reservoirs).

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$apiBin = Join-Path $repoRoot "bin\api-sqlite.exe"
$updaterBin = Join-Path $repoRoot "bin\updater.exe"
$dbFile = Join-Path $repoRoot "data\embalses.db"
$feDir = Join-Path $repoRoot "web"
$pidFile = Join-Path $PSScriptRoot ".pids"

$apiBasePort = if ($args[0]) { [int]$args[0] } else { 8082 }
$feBasePort = if ($args[1]) { [int]$args[1] } else { 5173 }

function Write-Step { param([string]$Message) Write-Host "`n>>> $Message" -ForegroundColor Cyan }
function Write-Ok   { param([string]$Message) Write-Host "    OK: $Message" -ForegroundColor Green }
function Write-Warn { param([string]$Message) Write-Host "    WARN: $Message" -ForegroundColor Yellow }
function Write-ErrorMsg { param([string]$Message) Write-Host "    ERROR: $Message" -ForegroundColor Red }

function Find-FreePort {
    param([int]$BasePort)
    $port = $BasePort
    while ($port -le 65535) {
        $listener = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue | Select-Object -First 1
        if (-not $listener) { return $port }
        $port++
    }
    throw "No free port found starting from $BasePort"
}

function Ensure-GoBinary {
    param([string]$BinPath, [string]$CmdPath)
    if (Test-Path $BinPath) { return }
    Write-Warn "Binary not found: $BinPath"
    if (Get-Command go -ErrorAction SilentlyContinue) {
        Write-Step "Building $CmdPath with local Go..."
        Push-Location $repoRoot
        try { & go build -o $BinPath $CmdPath } finally { Pop-Location }
    } else {
        Write-Step "Building $CmdPath with Docker..."
        docker run --rm -v "${repoRoot}:/app" -w /app golang:1.23-alpine go build -buildvcs=false -o $BinPath $CmdPath
    }
    Write-Ok "Built $BinPath"
}

# --- Cleanup old processes ---
Write-Step "Cleaning up stale processes..."
Get-Process -Name "api-sqlite" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Get-Process -Name "vite" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Get-Process -Name "node" -ErrorAction SilentlyContinue | Where-Object { $_.Path -like "*web*" -or $_.CommandLine -like "*vite*" } | Stop-Process -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 1
Write-Ok "Cleanup done"

# --- Ensure database ---
if (-not (Test-Path $dbFile)) {
    Write-Step "Database not found. Creating $dbFile..."
    Ensure-GoBinary $updaterBin "./cmd/updater"
    & $updaterBin -db $dbFile -geo-only -seed-readings
    Write-Ok "Database created"
} else {
    $size = (Get-Item $dbFile).Length / 1MB
    Write-Ok "Database ready: $dbFile ($([math]::Round($size,2)) MB)"
}

# --- Ensure API binary ---
Ensure-GoBinary $apiBin "./cmd/api-sqlite"

# --- Pick free ports ---
$apiPort = Find-FreePort $apiBasePort
$fePort = Find-FreePort $feBasePort
Write-Ok "API will use port: $apiPort"
Write-Ok "Frontend will use port: $fePort"

# --- Start API ---
Write-Step "Starting SQLite API on http://localhost:$apiPort..."
$env:DATABASE_URL = $dbFile
$env:API_ADDR = ":$apiPort"
$apiJob = Start-Job -ScriptBlock {
    param($bin, $db, $port)
    $env:DATABASE_URL = $db
    $env:API_ADDR = ":$port"
    & $bin
} -ArgumentList $apiBin, $dbFile, $apiPort
Start-Sleep -Seconds 2
if ($apiJob.State -eq "Failed") {
    Write-ErrorMsg "API failed to start"
    exit 1
}
Write-Ok "API started in background job (ID: $($apiJob.Id))"

# --- Frontend dependencies ---
Push-Location $feDir
try {
    if (-not (Test-Path "node_modules")) {
        Write-Step "Installing frontend dependencies..."
        npm install
        Write-Ok "Dependencies installed"
    }

    # --- Start frontend ---
    Write-Step "Starting Vite frontend on http://localhost:$fePort..."
    $frontendJob = Start-Job -ScriptBlock {
        param($dir, $port)
        Set-Location $dir
        $env:VITE_API_URL = "http://localhost:$port"
        npx vite --port $port
    } -ArgumentList $feDir, $fePort
    Start-Sleep -Seconds 3
    if ($frontendJob.State -eq "Failed") {
        Write-ErrorMsg "Frontend failed to start"
        Stop-Job $apiJob -ErrorAction SilentlyContinue
        exit 1
    }
    Write-Ok "Frontend started in background job (ID: $($frontendJob.Id))"
} finally {
    Pop-Location
}

# --- Save PIDs ---
@{
    apiJob = $apiJob.Id
    frontendJob = $frontendJob.Id
    apiPort = $apiPort
    fePort = $fePort
} | ConvertTo-Json | Out-File -FilePath $pidFile -Encoding UTF8

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "Embalses MVP is running!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "  Frontend:  http://localhost:$fePort"
Write-Host "  API:       http://localhost:$apiPort"
Write-Host "  DB:        $dbFile"
Write-Host "`nTo stop everything, run: .\scripts\stop.ps1"
Write-Host "Press Ctrl+C to exit this script (services keep running in background)"

# Keep the script alive so Ctrl+C can stop the jobs cleanly.
try {
    while ($true) {
        Start-Sleep -Seconds 1
        if ($apiJob.State -eq "Completed" -or $apiJob.State -eq "Failed" -or $apiJob.State -eq "Stopped") {
            Write-Warn "API job ended unexpectedly. Stopping frontend..."
            Stop-Job $frontendJob -ErrorAction SilentlyContinue
            break
        }
        if ($frontendJob.State -eq "Completed" -or $frontendJob.State -eq "Failed" -or $frontendJob.State -eq "Stopped") {
            Write-Warn "Frontend job ended unexpectedly. Stopping API..."
            Stop-Job $apiJob -ErrorAction SilentlyContinue
            break
        }
    }
} finally {
    if (Test-Path $pidFile) { Remove-Item $pidFile -Force }
    Stop-Job $apiJob -ErrorAction SilentlyContinue
    Stop-Job $frontendJob -ErrorAction SilentlyContinue
    Remove-Job $apiJob -ErrorAction SilentlyContinue
    Remove-Job $frontendJob -ErrorAction SilentlyContinue
}
