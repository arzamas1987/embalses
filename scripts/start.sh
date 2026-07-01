#!/usr/bin/env bash
# scripts/start.sh — Start the Embalses SQLite stack and Vite frontend on free ports.
# Usage: ./scripts/start.sh [api-base-port] [frontend-base-port]
#
# Defaults: API 8082, frontend 5173. If a port is taken, the next free one is chosen.
# The script uses the existing data/embalses.db (368+ SNCZI reservoirs). If the DB is
# missing, it builds/runs the updater to create it.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_BIN="${REPO_ROOT}/bin/api-sqlite"
UPDATER_BIN="${REPO_ROOT}/bin/updater"
DB_FILE="${REPO_ROOT}/data/embalses.db"
FE_DIR="${REPO_ROOT}/web"

API_BASE_PORT="${1:-8082}"
FE_BASE_PORT="${2:-5173}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log_step() { echo -e "${CYAN}==>${NC} $1"; }
log_ok()   { echo -e "${GREEN}  ✓${NC} $1"; }
log_warn() { echo -e "${YELLOW}  !${NC} $1"; }
log_error(){ echo -e "${RED}  ✗${NC} $1"; }

# Find the next free TCP port starting from $1.
find_free_port() {
    local base=$1
    local port=$base
    while true; do
        if ! lsof -Pi ":${port}" -sTCP:LISTEN -t >/dev/null 2>&1; then
            echo "$port"
            return
        fi
        port=$((port + 1))
        if [[ $port -gt 65535 ]]; then
            log_error "No free port found starting from $base"
            exit 1
        fi
    done
}

# Build a Go binary with Docker when Go is not installed locally.
ensure_go_binary() {
    local bin_path=$1
    local cmd_path=$2
    if [[ -x "$bin_path" ]]; then
        return
    fi

    log_warn "Binary not found: $bin_path"
    if command -v go >/dev/null 2>&1; then
        log_step "Building $cmd_path with local Go..."
        (cd "$REPO_ROOT" && go build -o "$bin_path" "$cmd_path")
    else
        log_step "Building $cmd_path with Docker..."
        docker run --rm -v "${REPO_ROOT}:/app" -w /app golang:1.23-alpine \
            go build -buildvcs=false -o "$bin_path" "$cmd_path"
    fi
    log_ok "Built $bin_path"
}

# Stop stale project processes so we do not conflict with old sessions.
cleanup_old_processes() {
    log_step "Cleaning up stale processes..."
    pkill -9 -f "api-sqlite" 2>/dev/null || true
    pkill -9 -f "vite" 2>/dev/null || true
    pkill -9 -f "npx.*vite" 2>/dev/null || true
    pkill -9 -f "node.*web" 2>/dev/null || true
    pkill -9 -f "go run.*api-sqlite" 2>/dev/null || true
    sleep 1
    log_ok "Cleanup done"
}

# ─── Main ─────────────────────────────────────────────────────────────────────
cleanup_old_processes

# Ensure the SQLite database exists. The checked-in DB already contains 368+
# SNCZI reservoirs; rebuild only when it is missing.
if [[ ! -f "$DB_FILE" ]]; then
    log_step "Database not found. Creating ${DB_FILE}..."
    ensure_go_binary "$UPDATER_BIN" "./cmd/updater"
    "$UPDATER_BIN" -db "$DB_FILE" -geo-only -seed-readings
    log_ok "Database created"
else
    log_ok "Database ready: $DB_FILE ($(du -h "$DB_FILE" | cut -f1))"
fi

# Ensure the SQLite API binary is available.
ensure_go_binary "$API_BIN" "./cmd/api-sqlite"

# Pick free ports so the script does not fail when the defaults are busy.
API_PORT=$(find_free_port "$API_BASE_PORT")
FE_PORT=$(find_free_port "$FE_BASE_PORT")
log_ok "API will use port: $API_PORT"
log_ok "Frontend will use port: $FE_PORT"

# Start API.
log_step "Starting SQLite API on http://localhost:${API_PORT}..."
API_ADDR=":${API_PORT}" DATABASE_URL="$DB_FILE" "$API_BIN" &
API_PID=$!
sleep 2
if ! kill -0 "$API_PID" 2>/dev/null; then
    log_error "API failed to start"
    exit 1
fi
log_ok "API running (PID $API_PID)"

# Install frontend dependencies if needed.
cd "$FE_DIR"
if [[ ! -d "node_modules" ]]; then
    log_step "Installing frontend dependencies..."
    npm install
    log_ok "Dependencies installed"
fi

# Start Vite frontend.
log_step "Starting Vite frontend on http://localhost:${FE_PORT}..."
npx vite --port "$FE_PORT" &
FE_PID=$!
sleep 3
if ! kill -0 "$FE_PID" 2>/dev/null; then
    log_error "Frontend failed to start"
    kill "$API_PID" 2>/dev/null || true
    exit 1
fi
log_ok "Frontend running (PID $FE_PID)"

# Summary.
echo ""
echo -e "${GREEN}=================================================${NC}"
echo -e "${GREEN}  Embalses MVP is running!${NC}"
echo -e "${GREEN}=================================================${NC}"
echo ""
echo "  API:      http://localhost:${API_PORT}"
echo "  Frontend: http://localhost:${FE_PORT}"
echo "  DB:       ${DB_FILE}"
echo ""
echo "  Press Ctrl+C to stop both servers"
echo ""

# Graceful shutdown.
cleanup() {
    echo ""
    log_step "Shutting down..."
    kill "$FE_PID" 2>/dev/null || true
    kill "$API_PID" 2>/dev/null || true
    sleep 1
    log_ok "Done."
    exit 0
}
trap cleanup INT TERM

wait "$FE_PID"
