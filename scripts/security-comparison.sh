#!/bin/bash

# Security Scan Script
# Skanuje zabezpieczenia bieżącego brancha używając CIS Docker Benchmark + OWASP ZAP

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

# Pobierz nazwę bieżącego brancha
cd "$PROJECT_ROOT"
CURRENT_BRANCH=$(git branch --show-current)

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   SECURITY SCAN: ${CURRENT_BRANCH}${NC}"
echo -e "${BLUE}║   CIS Docker Benchmark + OWASP ZAP                       ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Przygotowanie katalogu na raporty
mkdir -p "$REPORT_DIR"
BRANCH_DIR="$REPORT_DIR/${CURRENT_BRANCH}_$TIMESTAMP"
mkdir -p "$BRANCH_DIR"

echo -e "${GREEN}Report directory: $BRANCH_DIR${NC}\n"

# ============================================================================
# PRZYGOTOWANIE: Zbuduj i uruchom aplikację
# ============================================================================
echo -e "${YELLOW}[Setup] Building and starting application${NC}"

# Zbuduj obrazy
echo "Building Docker images..."
docker-compose build 2>&1 | grep -E "(Building|Successfully)" || true

# Uruchom aplikację w tle
echo "Starting application stack..."
docker-compose up -d 2>&1 | grep -E "(Creating|Started)" || true

# Poczekaj aż aplikacja będzie gotowa
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
echo -e "\n${YELLOW}[1/2] CIS Docker Benchmark${NC}"

bench_dir="$HOME/.docker-bench-security"
if [ ! -d "$bench_dir" ]; then
    echo "Downloading Docker Bench Security..."
    git clone https://github.com/docker/docker-bench-security.git "$bench_dir" -q
fi

echo "Running CIS Docker Benchmark..."
cd "$bench_dir"
sh docker-bench-security.sh 2>&1 | sed 's/\x1b\[[0-9;]*m//g' > "$BRANCH_DIR/cis-docker-benchmark.txt" || true
cd "$PROJECT_ROOT"

# Wyciągnij wynik z raportu
cis_pass=$(grep -c "\[PASS\]" "$BRANCH_DIR/cis-docker-benchmark.txt" 2>/dev/null || echo 0)
cis_warn=$(grep -c "\[WARN\]" "$BRANCH_DIR/cis-docker-benchmark.txt" 2>/dev/null || echo 0)
cis_total=$((cis_pass + cis_warn))

if [ $cis_total -gt 0 ]; then
    cis_score=$((cis_pass * 100 / cis_total))
else
    cis_score=0
fi

echo -e "${GREEN}✓ CIS Benchmark complete: ${cis_pass} PASS, ${cis_warn} WARN (Score: ${cis_score}%)${NC}"

# ============================================================================
# TEST 2: OWASP ZAP Baseline Scan
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
    
    # Oblicz wynik
    if [ -f "$BRANCH_DIR/owasp-zap-report.json" ]; then
        zap_high=$(grep -c '"risk": "High"' "$BRANCH_DIR/owasp-zap-report.json" 2>/dev/null || echo 0)
        zap_medium=$(grep -c '"risk": "Medium"' "$BRANCH_DIR/owasp-zap-report.json" 2>/dev/null || echo 0)
        zap_low=$(grep -c '"risk": "Low"' "$BRANCH_DIR/owasp-zap-report.json" 2>/dev/null || echo 0)
        
        zap_score=$((100 - (zap_high * 15) - (zap_medium * 5) - (zap_low * 1)))
        [ $zap_score -lt 0 ] && zap_score=0
        
        echo -e "${GREEN}✓ OWASP ZAP complete: ${zap_high} High, ${zap_medium} Medium, ${zap_low} Low (Score: ${zap_score}%)${NC}"
    else
        zap_high=0
        zap_medium=0
        zap_low=0
        zap_score=0
        echo -e "${YELLOW}WARNING: OWASP ZAP report not generated${NC}"
    fi
else
    echo -e "${RED}ERROR: Application not responding, skipping OWASP ZAP scan${NC}"
    zap_high=0
    zap_medium=0
    zap_low=0
    zap_score=0
fi

# ============================================================================
# CLEANUP: Zatrzymaj aplikację
# ============================================================================
echo -e "\n${YELLOW}[Cleanup] Stopping application${NC}"
docker-compose down -v 2>&1 | grep -E "(Stopping|Removing)" || true

# ============================================================================
# GENERUJ RAPORT KOŃCOWY
# ============================================================================
echo -e "\n${YELLOW}Generating summary report...${NC}"

final_score=$(( (cis_score + zap_score) / 2 ))

# Określ ocenę
if [ $final_score -ge 90 ]; then
    grade="A"
    grade_color=$GREEN
elif [ $final_score -ge 80 ]; then
    grade="B"
    grade_color=$GREEN
elif [ $final_score -ge 70 ]; then
    grade="C"
    grade_color=$YELLOW
elif [ $final_score -ge 60 ]; then
    grade="D"
    grade_color=$YELLOW
else
    grade="F"
    grade_color=$RED
fi

# Zapisz raport tekstowy
cat > "$BRANCH_DIR/SUMMARY.txt" << EOF
═══════════════════════════════════════════════════════════
SECURITY SCAN SUMMARY
═══════════════════════════════════════════════════════════

Branch: $CURRENT_BRANCH
Date: $(date)
Report Directory: $BRANCH_DIR

───────────────────────────────────────────────────────────
TEST RESULTS
───────────────────────────────────────────────────────────

1. CIS Docker Benchmark
   • Passed:    $cis_pass tests
   • Warnings:  $cis_warn tests
   • Score:     $cis_score%

2. OWASP ZAP Baseline
   • High:      $zap_high vulnerabilities
   • Medium:    $zap_medium vulnerabilities
   • Low:       $zap_low vulnerabilities
   • Score:     $zap_score%

───────────────────────────────────────────────────────────
FINAL SCORE
───────────────────────────────────────────────────────────

Overall Score: $final_score%
Grade: $grade

───────────────────────────────────────────────────────────
DETAILED REPORTS
───────────────────────────────────────────────────────────

• CIS Docker Benchmark:  $BRANCH_DIR/cis-docker-benchmark.txt
• OWASP ZAP (text):      $BRANCH_DIR/owasp-zap.txt
• OWASP ZAP (HTML):      $BRANCH_DIR/owasp-zap-report.html
• OWASP ZAP (Markdown):  $BRANCH_DIR/owasp-zap-report.md
• OWASP ZAP (JSON):      $BRANCH_DIR/owasp-zap-report.json

═══════════════════════════════════════════════════════════
EOF

# Wyświetl podsumowanie
echo ""
cat "$BRANCH_DIR/SUMMARY.txt"
echo ""

echo -e "${grade_color}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${grade_color}║   FINAL GRADE: $grade ($final_score%)${NC}"
echo -e "${grade_color}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}Reports saved to: $BRANCH_DIR${NC}"
echo ""
