# CIS Docker Benchmark - Przewodnik

## Czym jest CIS Docker Benchmark?

**CIS Docker Benchmark** to zestaw najlepszych praktyk bezpieczeństwa dla kontenerów Docker, opracowany przez [Center for Internet Security (CIS)](https://www.cisecurity.org/).

**Wersja:** 1.6.0
**Łączna liczba kontroli:** 117

## Sekcje CIS Docker Benchmark

| Sekcja | Nazwa                       | Kontrole   | Opis                       |
| ------ | --------------------------- | ---------- | -------------------------- |
| **1**  | Host Configuration          | 1.1 - 1.2  | Konfiguracja systemu hosta |
| **2**  | Docker Daemon Configuration | 2.1 - 2.18 | Daemon Docker              |
| **3**  | Docker Daemon Files         | 3.1 - 3.24 | Uprawnienia plików         |
| **4**  | Container Images            | 4.1 - 4.11 | Bezpieczeństwo obrazów     |
| **5**  | Container Runtime           | 5.1 - 5.31 | Runtime security           |
| **6**  | Docker Security Operations  | 6.1 - 6.2  | Monitoring i auditing      |
| **7**  | Docker Swarm Configuration  | 7.1 - 7.10 | Swarm mode (jeśli używane) |

## Uruchomienie Skanu

### Automatyczny (wraz z OWASP ZAP)

```bash
cd /home/ursus/personal/engineer/PrediGroweeV2
./scripts/security-comparison.sh
```

### Manualny (tylko CIS)

```bash
# Instalacja
git clone https://github.com/docker/docker-bench-security.git ~/.docker-bench-security

# Uruchomienie
cd ~/.docker-bench-security
sudo sh docker-bench-security.sh
```

## Interpretacja Wyników

### Status Codes

| Status     | Znaczenie         | Akcja                               |
| ---------- | ----------------- | ----------------------------------- |
| **[PASS]** | ✅ Test przeszedł | OK - zachowaj obecną konfigurację   |
| **[WARN]** | ⚠️ Ostrzeżenie    | Rozważ naprawę (nie krytyczne)      |
| **[INFO]** | ℹ️ Informacja     | FYI - może wymagać uwagi            |
| **[NOTE]** | 📝 Notatka        | Nie dotyczy lub nie można sprawdzić |

### Przykładowe Wyniki

```
[PASS] 5.4  - Ensure that Linux kernel capabilities are restricted within containers
[PASS] 5.11 - Ensure that CPU priority is set appropriately on containers
[PASS] 5.12 - Ensure that the container's root filesystem is mounted as read only
[WARN] 4.1  - Ensure that a user for the container has been created
[INFO] 1.1.1 - Ensure a separate partition for containers has been created
```

## PrediGrowee - Wyniki

### Przed Zabezpieczeniami (main)

```
PASS: 4/117 (3.4%)
WARN: ~90
INFO: ~20
```

**Główne problemy:**

- ❌ Brak resource limits
- ❌ Default capabilities
- ❌ Writable filesystem
- ❌ Brak AppArmor profiles

### Po Zabezpieczeniach (security-fixes)

```
PASS: 18+/117 (15.4%)
WARN: ~75
INFO: ~20
```

**Naprawione kontrole:**

#### Sekcja 2: Docker Daemon Configuration

- ✅ 2.1 - User namespace remapping (`userns-remap: default`)
- ✅ 2.2 - Container isolation (`icc: false`)
- ✅ 2.6 - Live restore (`live-restore: true`)
- ✅ 2.14 - Userland proxy disabled (`userland-proxy: false`)

#### Sekcja 5: Container Runtime

- ✅ 5.4 - Restricted capabilities (`cap_drop: ALL`)
- ✅ 5.11 - CPU limits (`cpus: 0.5-1.0`)
- ✅ 5.12 - Memory limits (`memory: 256M-1024M`)
- ✅ 5.13 - PIDs limits (`pids: 50-200`)
- ✅ 5.14 - Localhost binding (`127.0.0.1:*`)
- ✅ 5.26 - Health checks (`HEALTHCHECK`)
- ✅ 5.27 - no-new-privileges (`no-new-privileges:true`)
- ✅ 5.29 - Read-only FS (`read_only: true`)
- ✅ 5.2-5.3 - AppArmor profiles (`apparmor=docker-default`)

## Dlaczego Nie 100%?

### Wymagają Zmian na Poziomie Hosta (Sekcja 1)

```bash
# Przykłady kontroli wymagających root access na hoście:

# 1.1.1 - Separate partition for containers
sudo mkdir /var/lib/docker
sudo mount /dev/sdb1 /var/lib/docker

# 1.1.3-1.1.5 - Audit rules
sudo auditctl -w /var/lib/docker -k docker
sudo auditctl -w /etc/docker -k docker
sudo auditctl -w /usr/bin/docker -k docker
```

### PostgreSQL Limitations (Sekcja 4)

**4.1 - Ensure a user for the container has been created**

```yaml
# ⚠️ WARN - PostgreSQL wymaga specyficznych uprawnień
auth_db:
  image: postgres:13
  # user: postgres  # Musimy użyć domyślnego użytkownika
```

**Kompromis:** Runtime security controls (capabilities, read-only FS) kompensują.

### Docker Content Trust (Sekcja 4)

**4.5 - Ensure Content trust for Docker is Enabled**

```bash
# Wymaga Docker Registry z Notary
export DOCKER_CONTENT_TRUST=1
```

**Status:** Planowane dla produkcji.

## Best Practices - Implementacja

### 1. Resource Limits (5.11-5.13)

```yaml
services:
  auth:
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
          pids: 100
```

### 2. Capabilities Drop (5.4)

```yaml
services:
  auth:
    cap_drop:
      - ALL # Drop wszystkich 37 capabilities
```

### 3. Read-Only Filesystem (5.29)

```yaml
services:
  auth:
    read_only: true
    tmpfs:
      - /tmp:size=100M,mode=1777
```

### 4. no-new-privileges (5.27)

```yaml
services:
  auth:
    security_opt:
      - no-new-privileges:true
```

### 5. AppArmor Profile (5.2)

```yaml
services:
  auth:
    security_opt:
      - apparmor=docker-default
```

### 6. Network Security (5.14)

```yaml
services:
  nginx:
    ports:
      - "127.0.0.1:8080:8080" # Localhost only
```

## Produkcja - Checklist

### Host Level (Wymagane dla Production)

- [ ] Separate partition dla `/var/lib/docker`
- [ ] Audit rules dla Docker daemon
- [ ] SELinux/AppArmor w enforcing mode
- [ ] Kernel hardening (sysctl)
- [ ] Log rotation configured
- [ ] Automated security updates

### Docker Daemon

- [ ] `daemon.json` installed w `/etc/docker/`
- [ ] Docker service restarted
- [ ] User namespace remapping verified
- [ ] TLS dla remote access (jeśli używane)

### Container Level

- [ ] All containers z resource limits
- [ ] All containers z capabilities drop
- [ ] All containers z read-only FS
- [ ] All containers z non-root user
- [ ] Health checks dla wszystkich serwisów
- [ ] Secrets via Docker Secrets (nie env vars)

### Monitoring

- [ ] CIS scan w CI/CD pipeline
- [ ] Daily automated scans
- [ ] Alerting dla WARN/FAIL statuses
- [ ] Trend analysis (tracking score over time)

## Narzędzia Pomocnicze

### docker-bench-security

```bash
# GitHub: https://github.com/docker/docker-bench-security
# Oficjalne narzędzie CIS do audytu Docker

# Instalacja
git clone https://github.com/docker/docker-bench-security.git

# Uruchomienie
cd docker-bench-security
sudo sh docker-bench-security.sh
```

### Dockle

```bash
# Linter dla Dockerfile (podobny do Hadolint)
# Sprawdza best practices CIS

docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  goodwithtech/dockle:latest [IMAGE_NAME]
```

### Trivy

```bash
# CVE scanner + CIS checks
trivy image --security-checks vuln,config [IMAGE_NAME]
```

## Referencje

- [CIS Docker Benchmark v1.6.0](https://www.cisecurity.org/benchmark/docker)
- [docker-bench-security GitHub](https://github.com/docker/docker-bench-security)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [NIST Container Security Guide](https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-190.pdf)

## FAQ

### Q: Dlaczego niektóre testy pokazują [NOTE]?

**A:** [NOTE] oznacza, że test nie ma zastosowania (np. Docker Swarm gdy nie używamy Swarm mode) lub nie może być sprawdzony automatycznie.

### Q: Czy muszę naprawić wszystkie WARN?

**A:** Nie wszystkie WARN są krytyczne. Priorytetyzuj:

1. **P1:** Runtime security (capabilities, resource limits)
2. **P2:** Network security (port binding)
3. **P3:** Host-level configuration (dla produkcji)

### Q: Jak często uruchamiać CIS scan?

**A:** Rekomendacje:

- **Development:** Po każdej zmianie w Dockerfile/docker-compose
- **CI/CD:** Przy każdym PR i merge do main
- **Production:** Codziennie (automated)

### Q: Co zrobić gdy score nie poprawia się?

**A:** Sprawdź:

1. Czy `daemon.json` został zainstalowany i Docker zrestartowany
2. Czy zmiany w `docker-compose.yml` zostały zastosowane (`docker-compose down && docker-compose up`)
3. Czy uprawnienia plików są poprawne

---

**Ostatnia aktualizacja:** Listopad 2025
**Wersja dokumentu:** 1.0
**Projekt:** PrediGroweeV2
