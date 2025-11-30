# Szybki Przewodnik - CIS Docker Benchmark

## Jak Uruchomić?

```bash
# Opcja 1: Pełne skanowanie (CIS + OWASP ZAP)
./scripts/security-comparison.sh

# Opcja 2: Tylko CIS Docker Benchmark
./scripts/cis-scan.sh

# Opcja 3: Manualnie
cd ~/.docker-bench-security
sudo sh docker-bench-security.sh
```

## Jak Czytać Wyniki?

### Status Codes

```
[PASS] ✅ - Test przeszedł (dobra konfiguracja)
[WARN] ⚠️ - Ostrzeżenie (rozważ poprawkę)
[INFO] ℹ️ - Informacja (może wymagać uwagi)
[NOTE] 📝 - Nie dotyczy lub nie można sprawdzić
```

### Przykładowe Wyniki

```bash
# ✅ Dobre - te kontrole przeszły
[PASS] 5.4  - Ensure Linux kernel capabilities are restricted
[PASS] 5.12 - Ensure container's root filesystem is mounted read only
[PASS] 5.27 - Ensure no-new-privileges security option is set

# ⚠️ Ostrzeżenia - warto naprawić
[WARN] 4.1  - Ensure a user for the container has been created
[WARN] 5.9  - Ensure host's network namespace is not shared

# ℹ️ Informacje - opcjonalne
[INFO] 1.1.1 - Ensure separate partition for containers
[INFO] 2.1  - Restrict network traffic between containers
```

## Interpretacja Score

### PrediGrowee - Przed vs Po

| Branch             | Score          | Status                   |
| ------------------ | -------------- | ------------------------ |
| **main**           | 4/117 (3.4%)   | ⚠️ Needs hardening       |
| **security-fixes** | 18/117 (15.4%) | ✅ Good runtime security |

### Co Oznacza Mój Score?

```
0-10%   ⚠️ Critical - Implement basic hardening
10-20%  ✅ Good - Runtime security OK
20-50%  ⭐ Very Good - Host + runtime security
50-80%  ⭐⭐ Excellent - Production ready
80-100% ⭐⭐⭐ Outstanding - Maximum security
```

## Top 5 Najważniejszych Kontroli

### 1. Capabilities (5.4)

```yaml
# ✅ PASS
cap_drop:
  - ALL
```

### 2. Read-Only Filesystem (5.12)

```yaml
# ✅ PASS
read_only: true
tmpfs:
  - /tmp:size=100M
```

### 3. Resource Limits (5.11-5.13)

```yaml
# ✅ PASS
deploy:
  resources:
    limits:
      cpus: "0.5"
      memory: 512M
      pids: 100
```

### 4. no-new-privileges (5.27)

```yaml
# ✅ PASS
security_opt:
  - no-new-privileges:true
```

### 5. AppArmor Profile (5.2)

```yaml
# ✅ PASS
security_opt:
  - apparmor=docker-default
```

## Częste WARN i Jak Je Naprawić

### WARN: 4.1 - Container running as root

**Problem:** PostgreSQL wymaga specific UID

**Rozwiązanie:**

```yaml
# Kompensuj przez runtime security
cap_drop:
  - ALL
read_only: true
security_opt:
  - no-new-privileges:true
```

### WARN: 5.9 - Host network namespace shared

**Problem:** Porty exposed na 0.0.0.0

**Rozwiązanie:**

```yaml
ports:
  - "127.0.0.1:8080:8080" # Localhost only
```

### WARN: 2.18 - Experimental features enabled

**Problem:** Docker daemon w trybie experimental

**Rozwiązanie:**

```json
// /etc/docker/daemon.json
{
  "experimental": false
}
```

## FAQ

**Q: Dlaczego mój score nie wzrasta powyżej 20%?**

A: 60% kontroli CIS dotyczy konfiguracji **hosta** (nie kontenera):

- Separate partition dla Docker
- Audit rules (auditd)
- SELinux/AppArmor na hoście
- TLS dla Docker daemon

Te kontrole są **ważne dla produkcji**, ale nie są krytyczne dla development.

**Q: Co powinienem naprawić najpierw?**

A: Priorytet:

1. ✅ **Capabilities** (5.4) - NAJWAŻNIEJSZE
2. ✅ **Resource limits** (5.11-13)
3. ✅ **Read-only FS** (5.12)
4. ✅ **no-new-privileges** (5.27)
5. ⚠️ **Port binding** (5.14)

**Q: Czy muszę osiągnąć 100%?**

A: **NIE**. Dla większości projektów:

- **15-20% = Good** (runtime security OK)
- **30-40% = Very Good** (+ host security)
- **50%+ = Excellent** (production ready)

100% jest praktycznie niemożliwe bez dedykowanej infrastruktury i specjalistycznych wymagań.

## Gdzie Znaleźć Raporty?

```bash
# Ostatni raport
ls -lt security-comparison-report/ | head -2

# Przykład
security-comparison-report/
└── security-fixes_20251115_143022/
    ├── cis-docker-benchmark.txt  ← TUTAJ
    └── owasp-zap-report.html
```

## Co Dalej?

1. ✅ Przeczytaj raport: `security-comparison-report/*/cis-docker-benchmark.txt`
2. 📚 Zobacz pełny przewodnik: [CIS_DOCKER_BENCHMARK_GUIDE.md](./CIS_DOCKER_BENCHMARK_GUIDE.md)
3. 🔧 Implementuj poprawki: [SECURITY_HARDENING.md](./SECURITY_HARDENING.md)
4. 🔄 Uruchom ponownie scan aby zweryfikować

## Przydatne Linki

- [CIS Docker Benchmark Official](https://www.cisecurity.org/benchmark/docker)
- [docker-bench-security GitHub](https://github.com/docker/docker-bench-security)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)

---

**Tip:** Uruchamiaj scan regularnie (np. przed każdym merge do main), aby monitorować postęp zabezpieczeń! 🛡️
