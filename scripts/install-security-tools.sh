#!/bin/bash
# Instalacja narzędzi bezpieczeństwa dla lokalnego developmentu
# Wymaga: Ubuntu/Debian, sudo privileges

set -e

# Kolory dla outputu
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Instalacja narzędzi bezpieczeństwa - PrediGroweeV2       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}\n"

# Sprawdź system operacyjny
if [[ ! -f /etc/debian_version ]]; then
    echo -e "${RED}❌ Ten skrypt działa tylko na Ubuntu/Debian${NC}"
    exit 1
fi

# 1. Trivy - CVE Scanner
echo -e "\n${YELLOW}[1/5] Instalacja Trivy (CVE Scanner)...${NC}"
if command -v trivy &> /dev/null; then
    echo -e "${GREEN}✓ Trivy już zainstalowany ($(trivy --version | head -n1))${NC}"
else
    echo "  Dodaję repozytorium Trivy..."
    sudo apt-get update
    sudo apt-get install -y wget apt-transport-https gnupg lsb-release
    wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo gpg --dearmor -o /usr/share/keyrings/trivy.gpg
    echo "deb [signed-by=/usr/share/keyrings/trivy.gpg] https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee /etc/apt/sources.list.d/trivy.list
    sudo apt-get update
    sudo apt-get install -y trivy
    echo -e "${GREEN}✓ Trivy zainstalowany${NC}"
fi

# 2. Syft - SBOM Generator
echo -e "\n${YELLOW}[2/5] Instalacja Syft (SBOM Generator)...${NC}"
if command -v syft &> /dev/null; then
    echo -e "${GREEN}✓ Syft już zainstalowany ($(syft version | grep Version))${NC}"
else
    echo "  Pobieram i instaluję Syft..."
    curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sudo sh -s -- -b /usr/local/bin
    echo -e "${GREEN}✓ Syft zainstalowany${NC}"
fi

# 3. Grype - Vulnerability Scanner
echo -e "\n${YELLOW}[3/5] Instalacja Grype (Vulnerability Scanner)...${NC}"
if command -v grype &> /dev/null; then
    echo -e "${GREEN}✓ Grype już zainstalowany ($(grype version | grep Version))${NC}"
else
    echo "  Pobieram i instaluję Grype..."
    curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sudo sh -s -- -b /usr/local/bin
    echo -e "${GREEN}✓ Grype zainstalowany${NC}"
fi

# 4. govulncheck - Go Vulnerabilities
echo -e "\n${YELLOW}[4/5] Instalacja govulncheck (Go Vulnerabilities)...${NC}"
if command -v govulncheck &> /dev/null; then
    echo -e "${GREEN}✓ govulncheck już zainstalowany${NC}"
else
    if command -v go &> /dev/null; then
        echo "  Instaluję govulncheck..."
        go install golang.org/x/vuln/cmd/govulncheck@latest

        # Dodaj $GOPATH/bin do PATH jeśli nie ma
        if [[ ":$PATH:" != *":$HOME/go/bin:"* ]]; then
            echo "  Dodaję ~/go/bin do PATH..."
            echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
            export PATH=$PATH:$HOME/go/bin
        fi
        echo -e "${GREEN}✓ govulncheck zainstalowany${NC}"
    else
        echo -e "${YELLOW}⚠️  Go nie jest zainstalowany - pomijam govulncheck${NC}"
        echo -e "${YELLOW}   Zainstaluj Go: https://go.dev/doc/install${NC}"
    fi
fi

# 5. Hadolint - Dockerfile Linter
echo -e "\n${YELLOW}[5/6] Instalacja Hadolint (Dockerfile Linter)...${NC}"
if command -v hadolint &> /dev/null; then
    echo -e "${GREEN}✓ Hadolint już zainstalowany ($(hadolint --version))${NC}"
else
    echo "  Pobieram i instaluję Hadolint..."
    HADOLINT_VERSION=$(curl -s "https://api.github.com/repos/hadolint/hadolint/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
    sudo curl -sL -o /usr/local/bin/hadolint \
      "https://github.com/hadolint/hadolint/releases/download/v${HADOLINT_VERSION}/hadolint-Linux-x86_64"
    sudo chmod +x /usr/local/bin/hadolint
    echo -e "${GREEN}✓ Hadolint zainstalowany${NC}"
fi

# 6. Dockle - Dockerfile Security Linter
echo -e "\n${YELLOW}[6/7] Instalacja Dockle (Dockerfile Security Linter)...${NC}"
if command -v dockle &> /dev/null; then
    echo -e "${GREEN}✓ Dockle już zainstalowany ($(dockle --version))${NC}"
else
    echo "  Pobieram i instaluję Dockle..."
    VERSION=$(curl --silent "https://api.github.com/repos/goodwithtech/dockle/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
    curl -L -o /tmp/dockle.deb https://github.com/goodwithtech/dockle/releases/download/v${VERSION}/dockle_${VERSION}_Linux-64bit.deb
    sudo dpkg -i /tmp/dockle.deb
    rm /tmp/dockle.deb
    echo -e "${GREEN}✓ Dockle zainstalowany${NC}"
fi

# 7. Instalacja Git Hooks
echo -e "\n${YELLOW}[7/7] Instalacja Git Hooks...${NC}"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

if [ -d "$REPO_ROOT/.git" ]; then
    echo "  Linkuję pre-commit hook..."
    ln -sf ../../scripts/git-hooks/pre-commit "$REPO_ROOT/.git/hooks/pre-commit"
    chmod +x "$REPO_ROOT/scripts/git-hooks/pre-commit"
    echo -e "${GREEN}✓ Git hook zainstalowany${NC}"
else
    echo -e "${YELLOW}⚠️  Katalog .git nie znaleziony - pomiń instalację git hooka${NC}"
fi

# Weryfikacja instalacji
echo -e "\n${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Weryfikacja instalacji                                    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}\n"

echo -e "${GREEN}Zainstalowane narzędzia:${NC}"
echo "  • Trivy:         $(trivy --version 2>&1 | head -n1 | cut -d' ' -f2)"
echo "  • Syft:          $(syft version 2>&1 | grep Version | cut -d' ' -f2)"
echo "  • Grype:         $(grype version 2>&1 | grep Version | cut -d' ' -f2)"
if command -v govulncheck &> /dev/null; then
    echo "  • govulncheck:   ✓"
else
    echo "  • govulncheck:   ❌ (wymaga Go)"
fi
if command -v dockle &> /dev/null; then
    echo "  • Dockle:        $(dockle --version 2>&1 | cut -d' ' -f2)"
else
    echo "  • Dockle:        ❌"
fi
if command -v hadolint &> /dev/null; then
    echo "  • Hadolint:      $(hadolint --version 2>&1 | cut -d' ' -f2)"
else
    echo "  • Hadolint:      ❌"
fi

if [ -L "$REPO_ROOT/.git/hooks/pre-commit" ]; then
    echo "  • Git hook:      ✓"
else
    echo "  • Git hook:      ❌"
fi

echo -e "\n${GREEN}✅ Instalacja zakończona!${NC}\n"
echo -e "${BLUE}Możesz teraz uruchomić:${NC}"
echo -e "  ${YELLOW}./scripts/security-scan-local.sh${NC} - pełne skanowanie bezpieczeństwa"
echo -e "  ${YELLOW}git commit${NC} - pre-commit hook automatycznie sprawdzi podstawowe rzeczy"
echo -e "\n${YELLOW}Pamiętaj o załadowaniu nowych zmiennych środowiskowych:${NC}"
echo -e "  ${YELLOW}source ~/.bashrc${NC}\n"
