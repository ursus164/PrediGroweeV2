#!/bin/bash

# Skrypt do lokalnego skanowania bezpieczeństwa przed pushem
# Użycie: ./scripts/security-scan-local.sh

set -e

echo "PrediGrowee Security Scan - local"
echo "======================================"
echo ""

# Kolory
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "Checking tools..."
TOOLS_MISSING=0

check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}❌ $1 is not installed${NC}"
        TOOLS_MISSING=1
    else
        echo -e "${GREEN}✅ $1${NC}"
    fi
}

check_tool docker
check_tool trivy
check_tool syft
check_tool grype

if command -v hadolint &> /dev/null; then
    echo -e "${GREEN}✅ hadolint${NC}"
    HADOLINT_AVAILABLE=1
else
    echo -e "${YELLOW}❌ hadolint (not installed)${NC}"
    HADOLINT_AVAILABLE=0
fi

# Go static analysis tools
if command -v go &> /dev/null; then
    echo -e "${GREEN}✅ go${NC}"
    GO_AVAILABLE=1
    
    check_tool govulncheck
    check_tool golangci-lint
    check_tool staticcheck
    check_tool gosec
else
    echo -e "${YELLOW}❌ go (not installed)${NC}"
    GO_AVAILABLE=0
fi

if [ $TOOLS_MISSING -eq 1 ]; then
    echo ""
    echo "Install required tools:"
    echo "  trivy: https://aquasecurity.github.io/trivy/latest/getting-started/installation/"
    echo "  syft: https://github.com/anchore/syft#installation"
    echo "  grype: https://github.com/anchore/grype#installation"
    echo "  hadolint (opcjonalny): https://github.com/hadolint/hadolint"
    exit 1
fi

echo ""

# Skanowanie Dockerfile z Hadolint (jeśli dostępne)
if [ $HADOLINT_AVAILABLE -eq 1 ]; then
    echo "Scanning Dockerfile (Hadolint)..."
    HADOLINT_ERRORS=0
    for service in auth quiz stats images admin; do
        if [ -f "./$service/Dockerfile" ]; then
            echo "  Checking $service/Dockerfile..."
            if ! hadolint --failure-threshold style "./$service/Dockerfile"; then
                echo -e "${RED}❌ Hadolint found problems in $service/Dockerfile${NC}"
                HADOLINT_ERRORS=1
            fi
        fi
    done

    if [ $HADOLINT_ERRORS -eq 0 ]; then
        echo -e "${GREEN}✅ All Dockerfile files have passed Hadolint${NC}"
    else
        echo -e "${YELLOW}⚠️  There were issues with Dockerfiles - consider repair${NC}"
    fi
    echo ""
fi

echo "Building Docker images..."

# Budowanie wszystkich mikroserwisów
SERVICES=("auth" "quiz" "stats" "images" "admin")
for service in "${SERVICES[@]}"; do
    echo "Building $service..."
    docker build -t predigrowee-$service:local ./$service -q
done

echo ""
echo "Scanning Trivy (CVE + Secrets)..."
for service in "${SERVICES[@]}"; do
    echo ""
    echo "--- $service ---"
    trivy image --severity HIGH,CRITICAL --scanners vuln,secret --ignore-unfixed predigrowee-$service:local
done

echo ""
echo "Generating SBOM (Syft)..."
mkdir -p ./security-reports/sbom
for service in "${SERVICES[@]}"; do
    syft predigrowee-$service:local -o spdx-json > ./security-reports/sbom/$service-sbom.json
    echo "✅ SBOM for $service has been generated"
done

echo ""
echo "Scanning SBOM (Grype)..."
mkdir -p ./security-reports/grype
for service in "${SERVICES[@]}"; do
    echo ""
    echo "--- $service ---"
    grype sbom:./security-reports/sbom/$service-sbom.json --fail-on medium
done

echo ""
echo "Checking Go dependencies (govulncheck)..."
for service in "${SERVICES[@]}"; do
    echo ""
    echo "--- $service ---"
    cd ./$service
    if command -v govulncheck &> /dev/null; then
        govulncheck ./... || true
    else
        echo "⚠️  govulncheck not installed, skipping..."
    fi
    cd ..
done

echo ""
echo "Running Go Static Analysis..."
if [ $GO_AVAILABLE -eq 1 ]; then
    for service in "${SERVICES[@]}"; do
        echo ""
        echo "=== $service ==="
        cd ./$service
        
        # golangci-lint
        if command -v golangci-lint &> /dev/null; then
            echo "  [1/4] golangci-lint..."
            golangci-lint run --timeout=5m || echo "    ⚠️  Found issues"
        fi
        
        # staticcheck
        if command -v staticcheck &> /dev/null; then
            echo "  [2/4] staticcheck..."
            staticcheck -checks all ./... || echo "    ⚠️  Found issues"
        fi
        
        # gosec
        if command -v gosec &> /dev/null; then
            echo "  [3/4] gosec (security)..."
            gosec -fmt=golint -quiet ./... || echo "    ⚠️  Security issues found"
        fi
        
        # go vet
        echo "  [4/4] go vet..."
        go vet ./... || echo "    ⚠️  Found issues"
        
        cd ..
    done
else
    echo "⚠️  Go not installed - skipping static analysis"
fi

echo ""
echo -e "${GREEN}✅ Scan has been finished!${NC}"
echo ""
echo "Reports have been saved in: ./security-reports/"
echo ""
