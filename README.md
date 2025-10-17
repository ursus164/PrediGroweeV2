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
- [database-init/README.md](./database-init/README.md) - Inicjalizacja baz danych
