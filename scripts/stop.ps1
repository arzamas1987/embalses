#!/usr/bin/env pwsh
# scripts/stop.ps1 — Stop all Embalses services started by scripts/start.ps1.

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$pidFile = Join-Path $PSScriptRoot ".pids"

function Write-Step { param([string]$Message) Write-Host "`n>>> $Message" -ForegroundColor Cyan }
function Write-Ok   { param([string]$Message) Write-Host "    OK: $Message" -ForegroundColor Green }
function Write-Warn { param([string]$Message) Write-Host "    WARN: $Message" -ForegroundColor Yellow }

Write-Step "Stopping Embalses services..."

# --- Stop tracked PowerShell background jobs ---
if (Test-Path $pidFile) {
    try {
        $pids = Get-Content $pidFile | ConvertFrom-Json
        if ($pids.apiJob) {
            Stop-Job -Id $pids.apiJob -ErrorAction SilentlyContinue
            Remove-Job -Id $pids.apiJob -ErrorAction SilentlyContinue
            Write-Ok "Stopped API background job"
        }
        if ($pids.frontendJob) {
            Stop-Job -Id $pids.frontendJob -ErrorAction SilentlyContinue
            Remove-Job -Id $pids.frontendJob -ErrorAction SilentlyContinue
            Write-Ok "Stopped frontend background job"
        }
    } catch {
        Write-Warn "Could not read PID file: $_"
    }
    Remove-Item $pidFile -Force
    Write-Ok "Removed PID file"
}

# --- Stop Node.js / Vite processes ---
Write-Step "Stopping Node.js / Vite processes..."
Get-Process -Name "node" -ErrorAction SilentlyContinue | Where-Object {
    $_.CommandLine -like "*vite*" -or $_.Path -like "*web*"
} | ForEach-Object {
    try {
        $_.Kill()
        Write-Ok "Stopped node process (PID: $($_.Id))"
    } catch {
        Write-Warn "Could not stop node process PID $($_.Id): $($_.Exception.Message)"
    }
}

# --- Stop Go API process ---
Write-Step "Stopping SQLite API processes..."
Get-Process -Name "api-sqlite" -ErrorAction SilentlyContinue | ForEach-Object {
    try {
        $_.Kill()
        Write-Ok "Stopped api-sqlite process (PID: $($_.Id))"
    } catch {
        Write-Warn "Could not stop api-sqlite process PID $($_.Id): $($_.Exception.Message)"
    }
}

# --- Check for remaining port listeners ---
Write-Step "Checking for remaining port listeners..."
$portsToCheck = @(8080, 8082, 5173, 5174)
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
Write-Host "To restart everything, run:"
Write-Host "  .\scripts\start.ps1"
