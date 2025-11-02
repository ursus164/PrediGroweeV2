#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORT_DIR="$PROJECT_ROOT/security-comparison-report"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

cd "$PROJECT_ROOT"
CURRENT_BRANCH=$(git branch --show-current)

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   OWASP ZAP: ${CURRENT_BRANCH}                            ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Przygotowanie katalogu na raporty
mkdir -p "$REPORT_DIR"
BRANCH_DIR="$REPORT_DIR/${CURRENT_BRANCH}_$TIMESTAMP"
mkdir -p "$BRANCH_DIR"

echo -e "${GREEN}Report directory: $BRANCH_DIR${NC}\n"

# ============================================================================
# PRZYGOTOWANIE: Build i uruchomienie aplikacji
# ============================================================================
echo -e "${YELLOW}[Setup] Building and starting application${NC}"

echo "Building Docker images..."
docker-compose build 2>&1 | grep -E "(Building|Successfully)" || true

echo "Starting application stack..."
docker-compose up -d 2>&1 | grep -E "(Creating|Started)" || true

echo "Waiting for application to be ready..."
max_wait=120
elapsed=0
ready=false

while [ $elapsed -lt $max_wait ]; do
    http_code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080 2>/dev/null || echo "000")
    if [ "$http_code" != "000" ] && [ "$http_code" != "502" ] && [ "$http_code" != "503" ]; then
        ready=true
        echo "✓ Application ready after ${elapsed}s (HTTP $http_code)"
        break
    fi
    sleep 5
    elapsed=$((elapsed + 5))
    [ $((elapsed % 20)) -eq 0 ] && echo "  Waiting... ${elapsed}s/${max_wait}s (HTTP $http_code)"
done

if [ "$ready" != "true" ]; then
    echo -e "${RED}WARNING: Application not fully responding after ${max_wait}s${NC}"
    echo "Continuing with security tests on running containers..."
    docker-compose ps
fi

# ============================================================================
# TEST 2: OWASP ZAP Scan
# ============================================================================
echo -e "\n${YELLOW}[2/2] OWASP ZAP Baseline Scan${NC}"

if [ "$ready" = "true" ]; then
    echo "Starting OWASP ZAP scan on http://localhost:8080..."

    # Ustaw uprawnienia dla ZAP (uid=1000)
    chmod 777 "$BRANCH_DIR"

    # Uruchom OWASP ZAP
    docker run --rm --network host \
        -v "$BRANCH_DIR":/zap/wrk/:rw \
        -u zap \
        -t ghcr.io/zaproxy/zaproxy:stable \
        zap-baseline.py \
        -t http://localhost:8080 \
        -r owasp-zap-report.html \
        -w owasp-zap-report.md \
        -J owasp-zap-report.json \
        > "$BRANCH_DIR/owasp-zap.txt" 2>&1 || true

    chmod 755 "$BRANCH_DIR"
