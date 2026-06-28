#!/usr/bin/env pwsh
# stop.ps1 - Stop all Embalses services

$ErrorActionPreference = "Stop"

$repoRoot = "C:\Users\whala\git\embalses"
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

Write-Step "Stopping Embalses services..."

# --- Stop PowerShell background jobs ---
Write-Step "Stopping background jobs (API + frontend)..."
$jobs = Get-Job -ErrorAction SilentlyContinue
if ($jobs) {
    foreach ($job in $jobs) {
        Stop-Job $job -ErrorAction SilentlyContinue
        Remove-Job $job -ErrorAction SilentlyContinue
        Write-Ok "Stopped job: $($job.Name) (ID: $($job.Id))"
    }
} else {
    Write-Warn "No background jobs found"
}

# Also stop from PID file if exists
if (Test-Path $pidFile) {
    Remove-Item $pidFile -Force
    Write-Ok "Removed PID file"
}

# --- Stop Node processes (vite preview, npm) ---
Write-Step "Stopping Node.js processes..."
$nodeProcesses = Get-Process -Name "node" -ErrorAction SilentlyContinue
if ($nodeProcesses) {
    foreach ($proc in $nodeProcesses) {
        try {
            $proc.Kill()
            Write-Ok "Stopped node process (PID: $($proc.Id))"
        } catch {
            Write-Warn "Could not stop node process PID $($proc.Id): $($_.Exception.Message)"
        }
    }
} else {
    Write-Warn "No node processes found"
}

# --- Stop Go processes (API server) ---
Write-Step "Stopping Go processes..."
$goProcesses = Get-Process -Name "api" -ErrorAction SilentlyContinue
if ($goProcesses) {
    foreach ($proc in $goProcesses) {
        try {
            $proc.Kill()
            Write-Ok "Stopped api process (PID: $($proc.Id))"
        } catch {
            Write-Warn "Could not stop api process PID $($proc.Id): $($_.Exception.Message)"
        }
    }
}

# Also try killing go.exe if it was compiled
$goProcs = Get-Process -Name "go" -ErrorAction SilentlyContinue
if ($goProcs) {
    foreach ($proc in $goProcs) {
        try {
            $proc.Kill()
            Write-Ok "Stopped go process (PID: $($proc.Id))"
        } catch {
            Write-Warn "Could not stop go process PID $($proc.Id): $($_.Exception.Message)"
        }
    }
}

# --- Stop Docker containers ---
Write-Step "Stopping Docker containers..."
$hasDocker = $null -ne (Get-Command docker -ErrorAction SilentlyContinue)
if ($hasDocker) {
    Push-Location $repoRoot
    try {
        docker compose down 2>$null
        Write-Ok "Docker containers stopped"
    } catch {
        Write-Warn "Docker compose down failed or containers were already stopped"
    } finally {
        Pop-Location
    }
} else {
    Write-Warn "Docker not found, skipping container cleanup"
}

# --- Clean up any remaining port listeners ---
Write-Step "Checking for port conflicts..."
$portsToCheck = @(8080, 4174, 5432)
foreach ($port in $portsToCheck) {
    $listener = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($listener) {
        $proc = Get-Process -Id $listener.OwningProcess -ErrorAction SilentlyContinue
        if ($proc) {
            Write-Warn "Port $port still in use by $($proc.ProcessName) (PID: $($proc.Id))"
        }
    }
}

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "All Embalses services stopped." -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "`nTo restart everything, run:"
Write-Host "  $repoRoot\scripts\start.ps1"
