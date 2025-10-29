# PrediGroweeV2 - Backend

Backend dla aplikacji Predigrowee 2.0 - projekt pracy inżynierskiej.
Aplikacja dostępna pod adresem `predigrowee.agh.edu.pl`.

## Szybki Start

### Uruchomienie lokalne

```bash
# Uruchom wszystkie serwisy
docker compose up -d

# Backend dostępny na: http://localhost:8080
```

### Instalacja Narzędzi Bezpieczeństwa

Po sklonowaniu repozytorium zainstaluj narzędzia bezpieczeństwa i git hooki:

```bash
# Instalacja wszystkich narzędzi (Trivy, Syft, Grype, govulncheck, Dockle)
./scripts/install-security-tools.sh

# Załaduj nowe zmienne środowiskowe
source ~/.bashrc
```

**Co zostanie zainstalowane:**

- Trivy - skanowanie CVE w obrazach Docker
- Syft - generowanie SBOM
- Grype - skanowanie podatności
- govulncheck - sprawdzanie podatności Go
- Dockle - linting Dockerfile
- Pre-commit hook - automatyczne sprawdzenia przed commitem

**Pre-commit hook automatycznie sprawdzi:**

- Hardcoded secrets
- Wielkość plików
- Dockerfile best practices
- Go formatting
- Podatności w zależnościach

## Dokumentacja

- [SECURITY-PIPELINE.md](./SECURITY-PIPELINE.md) - Pipeline bezpieczeństwa i skanowanie
- [SECURITY_HARDENING.md](./SECURITY_HARDENING.md) - Docker security hardening
- [monitoring/README.md](./monitoring/README.md) - **Monitoring i Security Logging (Prometheus, Grafana, Loki)**
- [MONITORING_QUICKSTART.md](./MONITORING_QUICKSTART.md) - Szybki start monitoringu
- [database-init/README.md](./database-init/README.md) - Inicjalizacja baz danych

## Monitoring & Security Logging 📊

Stack monitoringu zapewnia:

- **Prometheus** - zbieranie metryk (CPU, RAM, błędy HTTP)
- **Grafana** - wizualizacja i dashboardy (http://localhost:3002)
- **Loki** - agregacja logów w czasie rzeczywistym
- **Promtail** - automatyczne zbieranie logów z kontenerów
- **Security monitoring** - śledzenie failed logins, 401/403, suspicious patterns

### Szybki start monitoringu

```bash
# 1. Utwórz sieć
docker network create predigroweev2_monitoring

# 2. Uruchom monitoring
docker-compose -f docker-compose.monitoring.yml up -d

# 3. Uruchom aplikację
docker-compose up -d

# 4. Otwórz Grafana
# http://localhost:3002 (admin/admin123)
```

Pełna dokumentacja: [monitoring/README.md](./monitoring/README.md)
