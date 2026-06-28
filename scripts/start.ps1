#!/usr/bin/env pwsh
# start.ps1 - Start all Embalses services
# Run from PowerShell (does NOT need Administrator)

$ErrorActionPreference = "Stop"

$repoRoot = "C:\Users\whala\git\embalses"
$webDir = "$repoRoot\web"
$logDir = "$repoRoot\logs"

# Ensure log directory exists
if (-not (Test-Path $logDir)) {
    New-Item -ItemType Directory -Path $logDir -Force | Out-Null
}

# PID file to track started processes
$pidFile = "$repoRoot\scripts\.pids"

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

# --- Check prerequisites ---
Write-Step "Checking prerequisites..."

$hasGo = $null -ne (Get-Command go -ErrorAction SilentlyContinue)
$hasNode = $null -ne (Get-Command node -ErrorAction SilentlyContinue)
$hasNpm = $null -ne (Get-Command npm -ErrorAction SilentlyContinue)
$hasDocker = $null -ne (Get-Command docker -ErrorAction SilentlyContinue)

if (-not $hasGo) {
    Write-Warn "Go not found. Please run scripts/setup.ps1 first."
    exit 1
}
if (-not $hasNode -or -not $hasNpm) {
    Write-Warn "Node.js/npm not found. Please run scripts/setup.ps1 first."
    exit 1
}
if (-not $hasDocker) {
    Write-Warn "Docker not found. Please install Docker Desktop first."
    exit 1
}

Write-Ok "All prerequisites found"

# --- Start Docker Desktop if not running ---
Write-Step "Starting Docker Desktop..."
$dockerProcess = Get-Process "Docker Desktop" -ErrorAction SilentlyContinue
if (-not $dockerProcess) {
    $dockerExe = "C:\Program Files\Docker\Docker\Docker Desktop.exe"
    if (Test-Path $dockerExe) {
        Start-Process $dockerExe
        Write-Ok "Docker Desktop starting..."
        Write-Host "    Waiting for Docker to be ready..."
        $tries = 0
        while ($tries -lt 30) {
            Start-Sleep -Seconds 2
            try {
                $null = docker ps 2>$null
                Write-Ok "Docker is ready"
                break
            } catch {
                $tries++
            }
        }
        if ($tries -ge 30) {
            Write-Warn "Docker may not be fully ready yet. Continuing anyway..."
        }
    } else {
        Write-Warn "Docker Desktop executable not found. Make sure Docker Desktop is installed."
    }
} else {
    Write-Ok "Docker Desktop already running"
}

# --- Start PostgreSQL ---
Write-Step "Starting PostgreSQL..."
Push-Location $repoRoot
try {
    docker compose up -d db
    Write-Ok "PostgreSQL container started"
    Write-Host "    Waiting for PostgreSQL to be ready..."
    Start-Sleep -Seconds 3
} finally {
    Pop-Location
}

# --- Set environment ---
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/embalses?sslmode=disable"

# --- Check if database needs migrations ---
Write-Step "Running database migrations..."
Push-Location $repoRoot
try {
    # Try to find migrate binary or download it
    $migratePath = "$env:USERPROFILE\go\bin\migrate.exe"
    if (-not (Test-Path $migratePath)) {
        Write-Host "    Installing golang-migrate..."
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    }
    
    if (Test-Path $migratePath) {
        & $migratePath -path migrations -database "$env:DATABASE_URL" up
        Write-Ok "Migrations applied"
    } else {
        Write-Warn "Could not find migrate. Please install it: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    }
} finally {
    Pop-Location
}

# --- Seed data ---
Write-Step "Seeding data (6 months of readings)..."
Push-Location $repoRoot
try {
    go run ./cmd/seed 2>&1 | Tee-Object -FilePath "$logDir\seed.log"
    Write-Ok "Data seeded"
} catch {
    Write-Warn "Seed may have already run or had errors. Check logs: $logDir\seed.log"
}
finally {
    Pop-Location
}

# --- Start API server ---
Write-Step "Starting API server on port 8080..."
Push-Location $repoRoot
try {
    $apiJob = Start-Job -ScriptBlock {
        param($root)
        $env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/embalses?sslmode=disable"
        Set-Location $root
        go run ./cmd/api 2>&1
    } -ArgumentList $repoRoot

    Write-Ok "API server started in background job (ID: $($apiJob.Id))"
    Start-Sleep -Seconds 2
} finally {
    Pop-Location
}

# --- Build frontend ---
Write-Step "Building frontend..."
Push-Location $webDir
try {
    npm run build 2>&1 | Tee-Object -FilePath "$logDir\build.log"
    Write-Ok "Frontend built successfully"
} finally {
    Pop-Location
}

# --- Start frontend preview ---
Write-Step "Starting frontend on http://localhost:4174..."
Push-Location $webDir
try {
    $frontendJob = Start-Job -ScriptBlock {
        param($dir)
        Set-Location $dir
        npm run preview -- --port 4174 2>&1
    } -ArgumentList $webDir

    Write-Ok "Frontend started in background job (ID: $($frontendJob.Id))"
    Start-Sleep -Seconds 3
} finally {
    Pop-Location
}

# --- Save PIDs for cleanup ---
$pids = @{
    apiJob = $apiJob.Id
    frontendJob = $frontendJob.Id
}
$pids | ConvertTo-Json | Out-File -FilePath $pidFile -Encoding UTF8

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "Embalses is running!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "  Frontend:  http://localhost:4174"
Write-Host "  API:       http://localhost:8080"
Write-Host "  Health:    http://localhost:8080/healthz"
Write-Host "  Logs:      $logDir"
Write-Host "`nTo stop everything, run: $repoRoot\scripts\stop.ps1"
Write-Host "`nPress Ctrl+C to exit this script (services keep running in background)"
Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
