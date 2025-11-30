#!/bin/bash
# Quick CIS Docker Benchmark scan
# Usage: ./scripts/cis-scan.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORT_DIR="$PROJECT_ROOT/security-comparison-report"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

cd "$PROJECT_ROOT"
CURRENT_BRANCH=$(git branch --show-current)

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   CIS Docker Benchmark - Quick Scan                       ║${NC}"
echo -e "${BLUE}║   Branch: ${CURRENT_BRANCH}                               ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Prepare report directory
mkdir -p "$REPORT_DIR"
BRANCH_DIR="$REPORT_DIR/cis_${CURRENT_BRANCH}_$TIMESTAMP"
mkdir -p "$BRANCH_DIR"

echo -e "${GREEN}Report directory: $BRANCH_DIR${NC}\n"

# Check if docker-bench-security is installed
CIS_DIR="$HOME/.docker-bench-security"

if [ ! -d "$CIS_DIR" ]; then
    echo "Installing docker-bench-security..."
    git clone https://github.com/docker/docker-bench-security.git "$CIS_DIR"
else
    echo "docker-bench-security already installed"
    # Update to latest version
    cd "$CIS_DIR" && git pull origin main 2>/dev/null || true
    cd "$PROJECT_ROOT"
fi

# Get PrediGrowee container IDs
echo "Detecting PrediGrowee containers..."
cd "$PROJECT_ROOT"
CONTAINER_IDS=$(docker compose ps -q 2>/dev/null | tr '\n' ',' | sed 's/,$//')

if [ -z "$CONTAINER_IDS" ]; then
    echo -e "${YELLOW}Warning: No running containers detected${NC}"
    echo "Starting containers..."
    docker compose up -d
    sleep 5
    CONTAINER_IDS=$(docker compose ps -q 2>/dev/null | tr '\n' ',' | sed 's/,$//')
fi

if [ -z "$CONTAINER_IDS" ]; then
    echo -e "${YELLOW}Error: Could not detect PrediGrowee containers${NC}"
    echo "Make sure docker-compose.yml is in the current directory"
    exit 1
fi

echo "Found container IDs: $CONTAINER_IDS"
echo ""

# Run CIS Docker Benchmark only for PrediGrowee containers
echo -e "\n${YELLOW}Running CIS Docker Benchmark...${NC}\n"
cd "$CIS_DIR"

# Run CIS scan - it will check all containers but we'll focus on container runtime
sudo sh docker-bench-security.sh 2>&1 | tee "$BRANCH_DIR/cis-docker-benchmark.txt"

cd "$PROJECT_ROOT"

# Parse results
if [ -f "$BRANCH_DIR/cis-docker-benchmark.txt" ]; then
    PASS_COUNT=$(grep -c "\[PASS\]" "$BRANCH_DIR/cis-docker-benchmark.txt" 2>/dev/null || echo "0")
    WARN_COUNT=$(grep -c "\[WARN\]" "$BRANCH_DIR/cis-docker-benchmark.txt" 2>/dev/null || echo "0")
    INFO_COUNT=$(grep -c "\[INFO\]" "$BRANCH_DIR/cis-docker-benchmark.txt" 2>/dev/null || echo "0")
    NOTE_COUNT=$(grep -c "\[NOTE\]" "$BRANCH_DIR/cis-docker-benchmark.txt" 2>/dev/null || echo "0")

    TOTAL_CHECKS=$((PASS_COUNT + WARN_COUNT))
    if [ "$TOTAL_CHECKS" -gt 0 ]; then
        PASS_PERCENTAGE=$(awk "BEGIN {printf \"%.0f\", ($PASS_COUNT * 100 / $TOTAL_CHECKS)}")
    else
        PASS_PERCENTAGE=0
    fi

    echo ""
    echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  CIS Docker Benchmark Results - branch: ${CURRENT_BRANCH}${NC}"
    echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "${GREEN}✓ PASS: $PASS_COUNT${NC}"
    echo -e "${YELLOW}⚠ WARN: $WARN_COUNT${NC}"
    echo "ℹ INFO: $INFO_COUNT"
    echo "📝 NOTE: $NOTE_COUNT"
    echo ""
    echo -e "${GREEN}Pass Rate: ${PASS_PERCENTAGE}% (${PASS_COUNT}/${TOTAL_CHECKS})${NC}"
    echo ""
    echo "Full report: $BRANCH_DIR/cis-docker-benchmark.txt"
    echo ""
    echo -e "${BLUE}════════════════════════════════════════════════════${NC}"
    echo ""

    # Show top WARN items
    echo -e "${YELLOW}Top WARN items:${NC}"
    grep "\[WARN\]" "$BRANCH_DIR/cis-docker-benchmark.txt" | head -10
    echo ""

    # Recommendations
    if [ $PASS_PERCENTAGE -lt 15 ]; then
        echo -e "${YELLOW}⚠ Recommendation: Implement security hardening${NC}"
        echo "See: $PROJECT_ROOT/SECURITY_HARDENING.md"
    elif [ $PASS_PERCENTAGE -lt 50 ]; then
        echo -e "${GREEN}✓ Good progress! Continue improving security.${NC}"
    else
        echo -e "${GREEN}✓ Excellent security posture!${NC}"
    fi
    echo ""
fi
