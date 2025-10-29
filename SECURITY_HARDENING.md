# Security Hardening - Instrukcja użycia

## ✅ Zaimplementowane zabezpieczenia

### 1. **Capabilities** ✓
Wszystkie kontenery aplikacji mają zredukowane capabilities:
```yaml
cap_drop:
  - ALL
```

Nginx dodatkowo ma tylko niezbędne:
```yaml
cap_add:
  - NET_BIND_SERVICE  # Port 80/443
  - CHOWN
  - SETGID
  - SETUID
```

### 2. **no-new-privileges** ✓
Zapobiega eskalacji uprawnień:
```yaml
security_opt:
  - no-new-privileges:true
```

### 3. **Read-only filesystem** ✓
Wszystkie kontenery mają read-only filesystem z tmpfs dla katalogów tymczasowych:
```yaml
read_only: true
tmpfs:
  - /tmp:size=100M,mode=1777
```

### 4. **Resource limits** ✓
Ograniczenia CPU i pamięci:
```yaml
deploy:
  resources:
    limits:
      cpus: '0.5'
      memory: 512M
```

## 🔐 Produkcja: Docker Secrets

Dla `docker-compose.prod.yml`:

### 1. Stwórz pliki sekretów

```bash
mkdir -p secrets
echo "super_secret_password_123" > secrets/auth_db_password.txt
echo "your_jwt_secret_here" > secrets/jwt_secret.txt
chmod 600 secrets/*.txt
```

### 2. Zdefiniuj secrets w compose

```yaml
secrets:
  auth_db_password:
    file: ./secrets/auth_db_password.txt
  jwt_secret:
    file: ./secrets/jwt_secret.txt
  quiz_db_password:
    file: ./secrets/quiz_db_password.txt
  # ... itd
```

### 3. Użyj w serwisach

```yaml
services:
  auth:
    secrets:
      - auth_db_password
      - jwt_secret
    environment:
      - DB_PASSWORD_FILE=/run/secrets/auth_db_password
      - JWT_SECRET_FILE=/run/secrets/jwt_secret
```

### 4. Zmień kod Go aby czytał z pliku

```go
// utils/secrets.go
package utils

import (
    "io/ioutil"
    "os"
    "strings"
)

func GetSecret(envVar, secretFile string) string {
    // Najpierw sprawdź czy jest _FILE env var
    if secretPath := os.Getenv(envVar + "_FILE"); secretPath != "" {
        content, err := ioutil.ReadFile(secretPath)
        if err == nil {
            return strings.TrimSpace(string(content))
        }
    }
    // Fallback do zwykłej env var (dla dev)
    return os.Getenv(envVar)
}
```

Użycie:
```go
dbPassword := utils.GetSecret("DB_PASSWORD", "/run/secrets/auth_db_password")
```

---

## ✅ Weryfikacja zabezpieczeń

### 1. Sprawdź capabilities

```bash
docker inspect predigrowee-auth-1 | jq '.[0].HostConfig.CapDrop'
# Powinno pokazać: ["ALL"]
```

### 2. Sprawdź security_opt

```bash
docker inspect predigrowee-auth-1 | jq '.[0].HostConfig.SecurityOpt'
# Powinno pokazać: ["no-new-privileges:true"]
```

### 3. Sprawdź read-only

```bash
docker exec predigrowee-auth-1 touch /test
# Powinno failować: "Read-only file system"
```

### 4. Sprawdź user

```bash
docker exec predigrowee-auth-1 whoami
# Powinno pokazać: "appuser" lub "nextjs" (NIE "root")
```

### 5. Sprawdź resource limits

```bash
docker stats --no-stream
# Sprawdź czy limity są ustawione
```

---

## 📊 CIS Docker Benchmark

Uruchom audit zabezpieczeń:

```bash
# Pobierz docker-bench-security
git clone https://github.com/docker/docker-bench-security.git
cd docker-bench-security

# Uruchom audit
sudo sh docker-bench-security.sh
```

Oczekiwany wynik: **> 85% PASS**

## 📈 Monitoring

Dodaj do `docker-compose.yml`:

```yaml
  prometheus:
    image: prom/prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp
```

---

## 🔄 Aktualizacje

Po dodaniu nowych serwisów, **zawsze dodaj**:
1. `cap_drop: [ALL]`
2. `security_opt: [no-new-privileges:true]`
3. `read_only: true` + odpowiednie `tmpfs`
4. `deploy.resources.limits`

---

## 📚 Referencje

- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [Seccomp Profile Guide](https://docs.docker.com/engine/security/seccomp/)
- [AppArmor Documentation](https://gitlab.com/apparmor/apparmor/-/wikis/Documentation)
