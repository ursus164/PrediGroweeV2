# Security Scanning Pipeline

Automatyczne skanowanie bezpieczeństwa dla projektu PrediGrowee zgodnie z wymaganiami pracy inżynierskiej.

## Cel

Implementacja podstawowych mechanizmów zabezpieczających w środowisku konteneryzacji:

- Refactor obrazów (ograniczenie możliwych podatności - Capabilities... AppSec...)
- Statyczna analiza Dockerfile
- Skanowanie podatności (CVE) w obrazach
- Generowanie i analiza SBOM (Software Bill of Materials)
- Wykrywanie podatnych zależności
- Detekcja wycieków sekretów
- **CIS Docker Benchmark** - zgodność z najlepszymi praktykami bezpieczeństwa kontenerów
- **OWASP ZAP** - skanowanie bezpieczeństwa aplikacji webowej

## Automatyczne Skanowanie (GitHub Actions)

### Kiedy uruchamia się pipeline?

1. **Push do main/develop** - pełne skanowanie
2. **Pull Request** - skanowanie przed merge
3. **Codziennie o 2:00** - sprawdzanie nowych CVE

### Co jest skanowane?

#### Backend (Go Microservices)

- **Hadolint** - best practices w Dockerfile
- **Dockle** - lint Dockerfile (non-root user, minimalizacja warstw)
- **Trivy** - skanowanie CVE w obrazach Docker
- **Syft** - generowanie SBOM
- **Grype** - analiza podatności w SBOM
- **govulncheck** - sprawdzanie podatności w Go dependencies
- **Docker Compose** - analiza konfiguracji bezpieczeństwa

#### Frontend (Next.js)

- **Hadolint** - best practices w Dockerfile
- **Dockle** - lint Dockerfile
- **Trivy** - skanowanie CVE
- **npm audit** - sprawdzanie podatności w pakietach NPM
- **Syft + Grype** - SBOM i analiza
- **ESLint Security** - statyczna analiza kodu
- **Gitleaks** - wykrywanie wycieków sekretów

### Wyniki skanowania

Wyniki są dostępne w:

- **GitHub Security** tab - SARIF reports
- **GitHub Actions** - szczegółowe logi
- **Artifacts** - SBOM files do pobrania

## Lokalne Skanowanie

### Instalacja narzędzi

**Automatyczna instalacja (zalecane):**

```bash
# Uruchom skrypt instalacyjny (Ubuntu/Debian)
./scripts/install-security-tools.sh

# Załaduj nowe zmienne środowiskowe
source ~/.bashrc
```

Skrypt zainstaluje:

- Trivy (CVE scanner)
- Syft (SBOM generator)
- Grype (vulnerability scanner)
- govulncheck (Go vulnerabilities)
- Dockle (Dockerfile linter)

**Manualna instalacja:**

<details>
<summary>Kliknij aby rozwinąć instrukcje manualnej instalacji</summary>

```bash
# Trivy (CVE scanner)
wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
echo "deb https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
sudo apt-get update
sudo apt-get install trivy

# Syft (SBOM generator)
curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin

# Grype (vulnerability scanner)
curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin

# govulncheck (Go vulnerabilities)
go install golang.org/x/vuln/cmd/govulncheck@latest

# Dockle (Dockerfile linter) - opcjonalnie
VERSION=$(curl --silent "https://api.github.com/repos/goodwithtech/dockle/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
curl -L -o dockle.deb https://github.com/goodwithtech/dockle/releases/download/v${VERSION}/dockle_${VERSION}_Linux-64bit.deb
sudo dpkg -i dockle.deb
```

</details>

### Uruchomienie skanowania

```bash
# Pełne skanowanie (CIS + OWASP ZAP)
cd PrediGroweeV2
./scripts/security-comparison.sh

# Tylko backend
./scripts/security-scan-local.sh

# Frontend
cd PrediGroweeV2-UI
npm audit
trivy image frontend:prod
```

### Pojedyncze narzędzia

```bash
# CIS Docker Benchmark
cd ~/.docker-bench-security
sudo sh docker-bench-security.sh

# OWASP ZAP Baseline Scan
docker run --rm --network host \
  -v $(pwd):/zap/wrk/:rw \
  -t ghcr.io/zaproxy/zaproxy:stable \
  zap-baseline.py -t http://localhost:8080

# Skanowanie konkretnego obrazu
trivy image predigrowee-auth:latest

# Generowanie SBOM
syft predigrowee-auth:latest -o spdx-json > auth-sbom.json

# Skanowanie SBOM
grype sbom:auth-sbom.json

# Sprawdzenie Go dependencies
cd auth && govulncheck ./...

# Lint Dockerfile
dockle predigrowee-auth:latest
```

## Pre-commit Hooks

### Instalacja Git Hooks

**Automatyczna instalacja (wraz z narzędziami):**

```bash
# Zainstaluj wszystkie narzędzia bezpieczeństwa i git hooki
./scripts/install-security-tools.sh
```

**Manualna instalacja git hooków:**

```bash
# Backend
cd PrediGroweeV2
ln -sf ../../scripts/git-hooks/pre-commit .git/hooks/pre-commit
chmod +x scripts/git-hooks/pre-commit

# Frontend
cd PrediGroweeV2-UI
ln -sf ../../scripts/git-hooks/pre-commit .git/hooks/pre-commit
chmod +x scripts/git-hooks/pre-commit
```

### Sprawdzenia wykonywane przez pre-commit hook

**Backend** (`PrediGroweeV2/scripts/git-hooks/pre-commit`):

- **Hardcoded secrets** - wykrywanie wrażliwych danych (passwords, API keys, tokens)
- **Wielkość plików** - blokada plików >1MB
- **Dockerfile lint**:
  - **Hadolint** - pełny linting Dockerfile
  - Fallback: sprawdzanie `:latest` tagów i `USER` directive
- **Go fmt** - automatyczne formatowanie kodu Go
- **govulncheck** - sprawdzanie podatności w zależnościach Go
- **npm audit** - sprawdzanie podatności npm (dla zmienionych package-lock.json)

**Frontend** (`PrediGroweeV2-UI/scripts/git-hooks/pre-commit`):

- **Hardcoded secrets** - wykrywanie wrażliwych danych
- **Wielkość plików** - blokada plików >1MB
- **Dockerfile lint**:
  - **Hadolint** - pełny linting Dockerfile
  - Fallback: sprawdzanie `:latest` i `USER` directive
- **ESLint** - linting kodu JavaScript/TypeScript
- **TypeScript type check** - sprawdzanie typów
- **npm audit** - sprawdzanie podatności (dla zmienionych package-lock.json)

### Użycie

Hooki uruchamiają się automatycznie przy każdym commicie:

```bash
git add .
git commit -m "feat: nowa funkcjonalność"
# Pre-commit hook uruchamia się automatycznie
```

# Lint Dockerfile

dockle predigrowee-auth:latest

````

## Interpretacja Wyników

### Poziomy severity:
- **CRITICAL** 🔴 - Natychmiastowa akcja wymagana
- **HIGH** 🟠 - Priorytetowe do naprawy
- **MEDIUM** 🟡 - Do naprawy w najbliższym czasie
- **LOW** 🟢 - Nice to have

### Typowe problemy i rozwiązania:

#### CVE w obrazach bazowych
```dockerfile
# ❌ Źle - stary obraz
FROM golang:1.20-alpine

# ✅ Dobrze - najnowszy obraz
FROM golang:1.23-alpine
````

#### Uruchamianie jako root

```dockerfile
# ❌ Źle
CMD ["./app"]

# ✅ Dobrze
RUN addgroup -g 1001 appuser && adduser -D -u 1001 -G appuser appuser
USER appuser
CMD ["./app"]
```

#### Nieużywane pakiety

```dockerfile
# ❌ Źle
RUN apk add curl wget git

# ✅ Dobrze - tylko to co potrzebne
RUN apk add --no-cache ca-certificates
```

## Konfiguracja

### Zmiana severity threshold

W `.github/workflows/security-scan.yml`:

```yaml
severity-cutoff: high # critical, high, medium, low
```

### Wyłączenie konkretnych CVE

Utwórz `.trivyignore`:

```
# Ignoruj CVE które są false positive
CVE-2023-12345
```

## Metryki Bezpieczeństwa

Pipeline generuje następujące metryki:

- Liczba CVE per severity
- Opis danego severity
- Rozmiar obrazów Docker
- Liczba podatnych zależności
- Błędy implementacyjne (Dockerfile)
- Błędy analizy statycznej i ich klasyfikacja po typie
- Coverage skanowania
- Czas naprawy podatności

## Dokumentacja Praca Inżynierska

Ten pipeline realizuje następujące punkty z pracy:

- ✅ Analiza standardów bezpieczeństwa w konteneryzacji
- ✅ Minimalizacja obrazów kontenerów
- ✅ Zabezpieczenie konfiguracji (non-root user)
- ✅ Automatyczne testowanie bezpieczeństwa (CI/CD)
- ✅ Logowanie zdarzeń bezpieczeństwa
- ✅ Dokumentacja procesu

## 📚 Referencje

- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [Syft Documentation](https://github.com/anchore/syft)
- [Grype Documentation](https://github.com/anchore/grype)
- [Docker Security Best Practices](https://docs.docker.com/develop/security-best-practices/)
- [OWASP Container Security](https://owasp.org/www-project-docker-top-10/)

```

```
