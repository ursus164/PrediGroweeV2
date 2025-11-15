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
echo -e "${BLUE}║   Security Comparison: ${CURRENT_BRANCH}                 ║${NC}"
echo -e "${BLUE}║   - CIS Docker Benchmark                                  ║${NC}"
echo -e "${BLUE}║   - OWASP ZAP Baseline Scan                               ║${NC}"
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
# TEST 1: CIS Docker Benchmark
# ============================================================================
echo -e "\n${YELLOW}[1/3] CIS Docker Benchmark (PrediGrowee containers)${NC}"

CIS_DIR="$HOME/.docker-bench-security"

# Sprawdź czy docker-bench-security jest zainstalowany
if [ ! -d "$CIS_DIR" ]; then
    echo "Installing docker-bench-security..."
    git clone https://github.com/docker/docker-bench-security.git "$CIS_DIR"
else
    echo "docker-bench-security already installed"
    # Aktualizacja do najnowszej wersji
    cd "$CIS_DIR" && git pull origin main 2>/dev/null || true
    cd "$PROJECT_ROOT"
fi

# Get PrediGrowee container IDs
echo "Detecting PrediGrowee containers..."
CONTAINER_IDS=$(docker compose ps -q 2>/dev/null | tr '\n' ',' | sed 's/,$//')

if [ -z "$CONTAINER_IDS" ]; then
    echo -e "${YELLOW}Note: Containers not running, starting them...${NC}"
    docker compose up -d
    sleep 5
    CONTAINER_IDS=$(docker compose ps -q 2>/dev/null | tr '\n' ',' | sed 's/,$//')
fi

if [ -z "$CONTAINER_IDS" ]; then
    echo -e "${YELLOW}Warning: Could not detect containers, scanning Docker daemon only${NC}"
else
    echo "Found container IDs: $CONTAINER_IDS"
fi

echo "Running CIS Docker Benchmark..."
cd "$CIS_DIR"

# Run full CIS scan
sudo sh docker-bench-security.sh > "$BRANCH_DIR/cis-docker-benchmark.txt" 2>&1 || true

cd "$PROJECT_ROOT"

# Parsuj wyniki CIS
if [ -f "$BRANCH_DIR/cis-docker-benchmark.txt" ]; then
    PASS_COUNT=$(grep -c "\[PASS\]" "$BRANCH_DIR/cis-docker-benchmark.txt" 2>/dev/null || echo "0")
    WARN_COUNT=$(grep -c "\[WARN\]" "$BRANCH_DIR/cis-docker-benchmark.txt" 2>/dev/null || echo "0")
    INFO_COUNT=$(grep -c "\[INFO\]" "$BRANCH_DIR/cis-docker-benchmark.txt" 2>/dev/null || echo "0")
    NOTE_COUNT=$(grep -c "\[NOTE\]" "$BRANCH_DIR/cis-docker-benchmark.txt" 2>/dev/null || echo "0")

    echo ""
    echo -e "${GREEN}✓ CIS Docker Benchmark Results:${NC}"
    echo "  PASS: $PASS_COUNT"
    echo "  WARN: $WARN_COUNT"
    echo "  INFO: $INFO_COUNT"
    echo "  NOTE: $NOTE_COUNT"
    echo ""
    echo "Full report: $BRANCH_DIR/cis-docker-benchmark.txt"
fi

# ============================================================================
# TEST 2: OWASP ZAP Scan
# ============================================================================
echo -e "\n${YELLOW}[2/3] OWASP ZAP Baseline Scan${NC}"

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

    # Parsuj wyniki OWASP ZAP
    if [ -f "$BRANCH_DIR/owasp-zap.txt" ]; then
        echo ""
        echo -e "${GREEN}✓ OWASP ZAP Scan completed${NC}"
        echo "HTML Report: $BRANCH_DIR/owasp-zap-report.html"
        echo "JSON Report: $BRANCH_DIR/owasp-zap-report.json"
        echo "Full log: $BRANCH_DIR/owasp-zap.txt"
    fi
else
    echo -e "${YELLOW}Skipping OWASP ZAP scan (application not ready)${NC}"
fi

# ============================================================================
# TEST 3: Summary
# ============================================================================
echo -e "\n${YELLOW}[3/3] Summary${NC}"
echo ""
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Security Scan Summary - ${CURRENT_BRANCH}${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo ""

# CIS Docker Benchmark Summary
if [ -f "$BRANCH_DIR/cis-docker-benchmark.txt" ]; then
    echo -e "${GREEN}CIS Docker Benchmark:${NC}"
    echo "  ✓ PASS: $PASS_COUNT"
    echo "  ⚠ WARN: $WARN_COUNT"
    echo "  ℹ INFO: $INFO_COUNT"
    echo ""
fi

# OWASP ZAP Summary
if [ -f "$BRANCH_DIR/owasp-zap.txt" ]; then
    echo -e "${GREEN}OWASP ZAP:${NC}"
    if grep -q "FAIL-NEW: 0" "$BRANCH_DIR/owasp-zap.txt"; then
        echo "  ✓ No new alerts"
    else
        echo "  ⚠ Check report for details"
    fi
    echo ""
fi

echo -e "${GREEN}All reports saved to: $BRANCH_DIR${NC}"
echo ""
echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
echo ""

# Opcjonalnie: otwórz raport w przeglądarce
if command -v xdg-open &> /dev/null && [ -f "$BRANCH_DIR/owasp-zap-report.html" ]; then
    read -p "Open OWASP ZAP HTML report in browser? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        xdg-open "$BRANCH_DIR/owasp-zap-report.html" &
    fi
fi