# System Referencyjny - PrediGrowee (Stan Wyjściowy)

**Opis:** Stan systemu przed wdrożeniem zabezpieczeń (gałąź `main`)
**Data analizy:** Listopad 2025
**Commit bazowy:** `c8ecf28` - Initial commit: PrediGrowee Backend

---

## 1. Cel i Charakterystyka Systemu

### 1.1 Cel Biznesowy

**PrediGrowee** to platforma dla ortodontów do interpretacji i predykcji wzrostu twarzy (twarzoczaszki) na podstawie zdjęć rentegnowskich i związanych z nimi parametrów (kątów pomiędzy danymi kościami itd...)

### 1.3 Charakterystyka Techniczna

**Architektura:** Microservices
**Backend:** Go 1.22.5
**Frontend:** Next.js 14 (Node 18)
**Baza danych:** PostgreSQL 13
**Konteneryzacja:** Docker + Docker Compose
**Reverse Proxy:** Nginx (Alpine)

---

## 2. Architektura Systemu

### 2.1 Diagram Architektury

```
                         ┌─────────────────┐
                         │   Użytkownik    │
                         │   (Przeglądarka)│
                         └────────┬────────┘
                                  │ HTTPS (8080, 3001)
                         ┌────────▼────────┐
                         │   Nginx Proxy   │
                         │  (API Gateway)  │
                         └────┬────────┬───┘
                              │        │
                  ┌───────────┘        └───────────┐
                  │                                 │
         ┌────────▼────────┐              ┌────────▼────────┐
         │  Backend APIs   │              │   Frontend      │
         │  (Port 8080)    │              │   Next.js       │
         └────┬────────────┘              │   (Port 3000)   │
              │                           └─────────────────┘
    ┌─────────┼──────────┬─────────┐
    │         │          │         │
┌───▼───┐ ┌──▼──┐ ┌─────▼───┐ ┌──▼───┐
│ Auth  │ │Quiz │ │ Stats   │ │Images│
│:8080  │ │:8080│ │ :8080   │ │:8080 │
└───┬───┘ └──┬──┘ └─────┬───┘ └──┬───┘
    │        │          │        │
┌───▼────┬───▼─────┬────▼────┬───▼────┐
│auth_db │ quiz_db │stats_db │images  │
│:5432   │ :5432   │ :5432   │_db:5432│
└────────┴─────────┴─────────┴────────┘
     PostgreSQL 13 (4 instancje)
```

### 2.2 Komponenty Systemu

#### Backend Microservices (Go)

| Serwis     | Port | Odpowiedzialność                                      | Zależności            |
| ---------- | ---- | ----------------------------------------------------- | --------------------- |
| **auth**   | 8080 | Autentykacja, autoryzacja, zarządzanie użytkownikami  | auth_db, JWT          |
| **quiz**   | 8080 | Zarządzanie pytaniami, sesje quizów, logika biznesowa | quiz_db, auth, stats  |
| **stats**  | 8080 | Agregacja statystyk, analiza wyników, rankingi        | stats_db, auth        |
| **images** | 8080 | Obsługa obrazów diagnostycznych                       | images_db, filesystem |
| **admin**  | 8080 | Panel administracyjny, zarządzanie systemem           | auth, quiz, stats     |

#### Frontend (Next.js)

| Komponent    | Port | Technologia                       | Funkcje                         |
| ------------ | ---- | --------------------------------- | ------------------------------- |
| **Frontend** | 3000 | Next.js 14, React 18, Material-UI | SSR, SPA routing, UI components |

#### Infrastruktura

| Komponent      | Image        | Rola                          |
| -------------- | ------------ | ----------------------------- |
| **Nginx**      | nginx:alpine | Reverse proxy, load balancing |
| **PostgreSQL** | postgres:13  | Persistent storage (4 bazy)   |

### 2.3 Sieci Docker

```yaml
networks:
  frontend_network: # Nginx ↔ Frontend
  backend_network: # Nginx ↔ Backend services
  db_network: # Backend ↔ Databases
```

**Izolacja:** Każda sieć jest separowana, ale bez zaawansowanych kontroli dostępu.

---

## 3. Przepływ Danych

### 3.1 Rejestracja i Logowanie

```
┌──────────┐     ┌────────┐     ┌──────┐     ┌─────────┐
│ Frontend │────▶│ Nginx  │────▶│ Auth │────▶│ auth_db │
└──────────┘     └────────┘     └──────┘     └─────────┘
     │                              │
     │◀─────── JWT Token ───────────┘
     │
┌────▼──────────────────────────────────┐
│ Token przechowywany w cookie          │
│ (httpOnly: true, secure: false)       │
└───────────────────────────────────────┘
```

**Przepływ:**

1. Użytkownik podaje email/hasło lub loguje przez Google OAuth
2. Auth weryfikuje dane w `auth_db.users`
3. Generuje JWT token (ważny 24h)
4. Token zapisywany w HttpOnly cookie
5. Frontend używa cookie do autoryzacji kolejnych żądań

### 3.2 Sesja Quizu (Educational Mode)

```
┌──────────┐     ┌────────┐     ┌──────┐     ┌─────────┐
│ Frontend │────▶│ Nginx  │────▶│ Quiz │────▶│ quiz_db │
└──────────┘     └────────┘     └──┬───┘     └─────────┘
                                    │
                                    │ verify token
                                ┌───▼───┐
                                │ Auth  │
                                └───┬───┘
                                    │ save stats
                                ┌───▼───┐     ┌──────────┐
                                │ Stats │────▶│ stats_db │
                                └───────┘     └──────────┘
```

**Przepływ:**

1. Frontend wywołuje `POST /api/quiz/start` z trybem
2. Quiz weryfikuje token z Auth
3. Quiz tworzy nową sesję w `quiz_db.quiz_sessions`
4. Quiz losuje pierwsze pytanie z grupy
5. Frontend wyświetla pytanie (obrazy z `/api/images/`)
6. Użytkownik odpowiada → `POST /api/quiz/submit`
7. Quiz zapisuje odpowiedź i zwraca wynik
8. Stats agreguje dane w `stats_db.responses`
9. Po zakończeniu: `POST /api/quiz/finish` → podsumowanie

### 3.3 Panel Administracyjny

```
┌─────────┐     ┌───────┐     ┌───────┐     ┌──────────┐
│ Frontend│────▶│ Admin │────▶│ Auth  │────▶│  auth_db │
└─────────┘     └───┬───┘     └───────┘     └──────────┘
                    │
                    ├──────────▶ Quiz  ────▶ quiz_db
                    │
                    └──────────▶ Stats ────▶ stats_db
```

**Funkcje:**

- Zarządzanie użytkownikami (CRUD)
- Tworzenie/edycja pytań
- Przeglądanie statystyk globalnych
- Eksport danych do analizy

### 3.4 Statystyki Użytkownika

```
┌──────────┐     ┌───────┐     ┌──────────┐
│ Frontend │────▶│ Stats │────▶│ stats_db │
└──────────┘     └───┬───┘     └──────────┘
                     │
                     │ verify user
                 ┌───▼───┐
                 │ Auth  │
                 └───────┘
```

**Agregowane dane:**

- Liczba odpowiedzi (poprawne/błędne)
- Accuracy per question
- Czas reakcji
- Ranking użytkowników
- Wykresy postępów

---

## 4. Bazy Danych

### 4.1 Schema: auth_db

**Główne tabele:**

```sql
-- Użytkownicy
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email VARCHAR(255) UNIQUE NOT NULL,
    google_id VARCHAR(255) UNIQUE,
    password VARCHAR(255),  -- bcrypt hash
    role VARCHAR(20) DEFAULT 'user',  -- user, teacher, admin
    created_at TIMESTAMP DEFAULT NOW(),
    verified BOOLEAN DEFAULT FALSE
);

-- Role (RBAC)
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

-- Sesje użytkowników
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    token VARCHAR(500) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    ip_address VARCHAR(45)
);

-- Reset hasła
CREATE TABLE password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    token VARCHAR(500) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE
);

-- OAuth providers
CREATE TABLE oauth_providers (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    provider VARCHAR(50),  -- google, facebook
    provider_user_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Relacje:**

- users ↔ roles (many-to-many przez user_roles)
- users → user_sessions (one-to-many)
- users → oauth_providers (one-to-many)

### 4.2 Schema: quiz_db

**Główne tabele:**

```sql
-- Parametry laboratoryjne
CREATE TABLE parameters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    reference_values VARCHAR(255),
    "order" INT DEFAULT 0
);

-- Opcje odpowiedzi
CREATE TABLE options (
    id SERIAL PRIMARY KEY,
    option VARCHAR(255) NOT NULL UNIQUE
);

-- Przypadki kliniczne (dzieci)
CREATE TABLE cases (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    patient_gender VARCHAR(10),  -- male, female
    age1 INT,  -- wiek w latach
    age2 INT,  -- wiek w miesiącach
    age3 INT   -- wiek w dniach
);

-- Wartości parametrów dla przypadku
CREATE TABLE parameters_values (
    id SERIAL PRIMARY KEY,
    case_id INT REFERENCES cases(id) ON DELETE CASCADE,
    parameter_id INT REFERENCES parameters(id),
    value1 NUMERIC(10,2),
    value2 NUMERIC(10,2),
    value3 NUMERIC(10,2)
);

-- Pytania (case + poprawna opcja)
CREATE TABLE questions (
    id SERIAL PRIMARY KEY,
    case_id INT REFERENCES cases(id) ON DELETE CASCADE,
    correct_option_id INT REFERENCES options(id),
    group_id INT,  -- grupowanie pytań
    created_at TIMESTAMP DEFAULT NOW()
);

-- Opcje dla pytania (możliwe diagnozy)
CREATE TABLE question_options (
    id SERIAL PRIMARY KEY,
    question_id INT REFERENCES questions(id) ON DELETE CASCADE,
    option_id INT REFERENCES options(id)
);

-- Sesje quizów
CREATE TABLE quiz_sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    mode VARCHAR(20) NOT NULL,  -- educational, exam, test
    current_question_id INT REFERENCES questions(id),
    current_group INT,
    group_order INT[],  -- kolejność pytań w grupie
    start_time TIMESTAMP DEFAULT NOW(),
    finish_time TIMESTAMP,
    test_code VARCHAR(50)  -- dla trybu test
);

-- Testy (zestawy pytań)
CREATE TABLE tests (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_by INT NOT NULL,  -- user_id nauczyciela
    created_at TIMESTAMP DEFAULT NOW()
);

-- Pytania w teście
CREATE TABLE test_questions (
    test_id INT REFERENCES tests(id) ON DELETE CASCADE,
    question_id INT REFERENCES questions(id) ON DELETE CASCADE,
    PRIMARY KEY (test_id, question_id)
);

-- Dostęp do quizu (approval system)
CREATE TABLE quiz_user_access (
    user_id INT PRIMARY KEY,
    approved BOOLEAN DEFAULT FALSE,
    approved_by INT,  -- admin user_id
    approved_at TIMESTAMP,
    registered_at TIMESTAMP DEFAULT NOW()
);

-- Ustawienia systemu
CREATE TABLE settings (
    name VARCHAR(100) PRIMARY KEY,
    value TEXT
);

-- Zgłoszenia błędów
CREATE TABLE case_reports (
    id SERIAL PRIMARY KEY,
    case_id INT REFERENCES cases(id),
    user_id INT,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Relacje:**

- cases → parameters_values → parameters
- questions → cases, options
- quiz_sessions → questions
- tests ↔ questions (many-to-many)

### 4.3 Schema: stats_db

**Główne tabele:**

```sql
-- Sesje quizów (kopia z quiz_db)
CREATE TABLE quiz_sessions (
    session_id INT PRIMARY KEY,
    user_id INT NOT NULL,
    mode VARCHAR(20) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    finish_time TIMESTAMP
);

-- Odpowiedzi użytkowników
CREATE TABLE responses (
    id SERIAL PRIMARY KEY,
    session_id INT REFERENCES quiz_sessions(session_id),
    question_id INT NOT NULL,
    selected_option_id INT NOT NULL,
    correct_option_id INT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    response_time INT,  -- ms
    answered_at TIMESTAMP DEFAULT NOW()
);

-- Ankieta użytkownika (demografia)
CREATE TABLE users_surveys (
    user_id INT PRIMARY KEY,
    name VARCHAR(100),
    surname VARCHAR(100),
    gender VARCHAR(20),
    age INT,
    country VARCHAR(100),
    vision_defect VARCHAR(50),  -- yes, no
    education VARCHAR(100),     -- medical_student, doctor, etc.
    experience VARCHAR(100)     -- years
);
```

**Agregowane dane:**

- Accuracy per user
- Accuracy per question
- Response time średni
- Statystyki grupowane (gender, age, education)

### 4.4 Schema: images_db

**Główne tabele:**

```sql
-- Metadata obrazów
CREATE TABLE images (
    id SERIAL PRIMARY KEY,
    case_id INT NOT NULL,
    filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    mime_type VARCHAR(50),
    size_bytes BIGINT,
    uploaded_at TIMESTAMP DEFAULT NOW(),
    uploaded_by INT  -- user_id
);
```

**Storage:** Obrazy przechowywane w `/app/images` (volume Docker)

---

## 5. Zależności Technologiczne

### 5.1 Backend (Go 1.22.5)

**Wspólne biblioteki dla wszystkich serwisów:**

```go
// go.mod (przykład: auth/go.mod)
module auth

go 1.22.5

require (
    github.com/go-playground/validator/v10 v10.22.1  // Walidacja
    github.com/golang-jwt/jwt/v5 v5.2.1              // JWT tokens
    github.com/lib/pq v1.10.9                        // PostgreSQL driver
    github.com/rs/cors v1.11.1                       // CORS middleware
    go.uber.org/zap v1.27.0                          // Structured logging
    golang.org/x/crypto v0.28.0                      // bcrypt, crypto
)
```

**Serwis-specyficzne:**

- **auth:** `github.com/jordan-wright/email` (email sending)
- **stats:** agregacja danych, statistyka
- **quiz:** logika losowania pytań, time limits

### 5.2 Frontend (Next.js 14)

```json
// package.json (główne zależności)
{
  "dependencies": {
    "next": "14.2.5",
    "react": "18.3.1",
    "react-dom": "18.3.1",
    "@mui/material": "^5.16.7", // UI components
    "@emotion/react": "^11.13.3", // CSS-in-JS
    "axios": "^1.7.4", // HTTP client
    "formik": "^2.4.6", // Forms
    "yup": "^1.4.0", // Validation
    "recharts": "^2.12.7", // Charts
    "typescript": "^5.5.4"
  }
}
```

### 5.3 Infrastruktura

**Docker Compose konfiguracja:**

```yaml
# docker-compose.yml (fragment)
services:
  nginx:
    image: nginx:alpine
    ports:
      - "8080:8080" # ⚠️ Exposed na 0.0.0.0 (wszędzie)
      - "3001:3000" # ⚠️ Exposed na 0.0.0.0
    # ❌ Brak limitów zasobów
    # ❌ Brak security_opt
    # ❌ Writable filesystem

  auth:
    build:
      context: ./auth
      dockerfile: Dockerfile
    expose:
      - "8080"
    environment:
      - DB_PASSWORD=auth_password # ⚠️ Plain text
      - JWT_SECRET=${JWT_SECRET} # ⚠️ Env var
    # ❌ Brak cap_drop
    # ❌ Brak read-only filesystem
    # ❌ Brak resource limits

  auth_db:
    image: postgres:13
    ports:
      - "5433:5432" # ⚠️ Exposed na 0.0.0.0
    environment:
      - POSTGRES_PASSWORD=auth_password # ⚠️ Plain text
    # ❌ Running as root (default postgres)
```

**Nginx konfiguracja (nginx.conf):**

```nginx
# ⚠️ STAN WYJŚCIOWY - BRAK SECURITY HEADERS
http {
    server {
        listen 8080;

        # ❌ Brak server_tokens off;
        # ❌ Brak CSP header
        # ❌ Brak X-Frame-Options
        # ❌ Brak X-Content-Type-Options
        # ❌ Brak Permissions-Policy

        location /api/auth/ {
            proxy_pass http://auth:8080/auth/;
            # ❌ Minimalne proxy headers
        }

        location /api/quiz/ {
            proxy_pass http://quiz:8080/quiz/;
        }
        # ... inne lokalizacje
    }
}
```

---

## 6. Identyfikacja Krytycznych Komponentów

### 6.1 Komponenty Krytyczne dla Bezpieczeństwa

| Komponent        | Krytyczność | Powód                                | Ryzyka                      |
| ---------------- | ----------- | ------------------------------------ | --------------------------- |
| **Auth Service** | 🔴 CRITICAL | Single point of authentication       | Compromise = pełny dostęp   |
| **auth_db**      | 🔴 CRITICAL | Przechowuje credentials, JWT secrets | Data breach = catastrophic  |
| **Nginx**        | 🟠 HIGH     | Entry point do systemu               | DDoS, injection attacks     |
| **JWT Token**    | 🟠 HIGH     | Bearer token autoryzacji             | Token theft = impersonation |
| **quiz_db**      | 🟡 MEDIUM   | Business logic data                  | Data integrity              |
| **stats_db**     | 🟡 MEDIUM   | Analytics, PII data                  | Privacy concerns            |

### 6.2 Single Points of Failure (SPOF)

1. **Auth Service**

   - Bez redundancji (1 instancja)
   - Failure → cały system niedostępny
   - Brak health check retry logic

2. **Nginx**

   - Single reverse proxy
   - Failure → brak dostępu do systemu
   - Brak load balancing

3. **PostgreSQL**
   - 4 separate instancje
   - Brak replikacji
   - Brak automated backups

### 6.3 Data Flow - Krytyczne Ścieżki

**Path 1: Authentication Flow**

```
User → Nginx → Auth → auth_db → JWT Token → Cookie
```

**Ryzyka:**

- MITM (brak TLS między serwisami)
- Token theft (HttpOnly, ale secure=false)
- SQL injection (używane prepared statements, ale brak dodatkowych walidacji)

**Path 2: Quiz Session**

```
User → Nginx → Quiz → quiz_db
                  ↓
                Stats → stats_db
```

**Ryzyka:**

- Race conditions (concurrent sessions)
- Data consistency (distributed transactions)
- Time manipulation (client-side time validation)

**Path 3: Admin Operations**

```
User → Nginx → Admin → [Auth, Quiz, Stats]
```

**Ryzyka:**

- Privilege escalation
- Mass data export (brak rate limiting)
- Unauthorized access (RBAC not enforced consistently)

### 6.4 Zewnętrzne Integracje

| Integracja       | Cel                   | Ryzyka                       |
| ---------------- | --------------------- | ---------------------------- |
| **Google OAuth** | SSO login             | OAuth token theft, phishing  |
| **SMTP (Gmail)** | Password reset emails | Email interception, spoofing |
| **Browser APIs** | Client-side storage   | XSS, storage manipulation    |

---

## 7. Stan Bezpieczeństwa Wyjściowego

### 7.1 Dockerfile - Przykład (auth/Dockerfile)

```dockerfile
# ⚠️ STAN PRZED HARDENING
FROM golang:1.22-alpine AS builder  # ❌ Stara wersja Go (1.22 vs 1.24)

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Runtime stage
FROM alpine:latest  # ❌ Tag :latest (nie pinned)

RUN apk --no-cache add ca-certificates curl  # ❌ Brak pinned versions

# ✅ Non-root user PRESENT (jedyne zabezpieczenie)
RUN addgroup -g 1001 appuser && \
    adduser -D -u 1001 -G appuser appuser && \
    chown -R appuser:appuser /app

COPY --from=builder --chown=appuser:appuser /app/main .
USER appuser

# ✅ Healthcheck PRESENT
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/auth/health || exit 1

CMD ["./main"]
# ❌ Brak EXPOSE (ale to minor issue)
```

**Co było OK:**

- ✅ Multi-stage build
- ✅ Non-root user (appuser:1001)
- ✅ Health check

**Problemy:**

- ❌ Go 1.22 (9 CVE w standardowej bibliotece)
- ❌ Alpine :latest (niepinned versions → CVE-2025-9230, 9231, 9232)
- ❌ Writable filesystem (brak read_only)
- ❌ Default capabilities (37 capabilities)
- ❌ Brak resource limits

### 7.2 docker-compose.yml - Stan Wyjściowy

**Całkowity brak security hardening:**

```yaml
# ⚠️ PRZYKŁAD: Serwis auth
auth:
  build:
    context: ./auth
    dockerfile: Dockerfile
  expose:
    - "8080"
  environment:
    - DB_PASSWORD=auth_password # ⚠️ Plain text
    - JWT_SECRET=${JWT_SECRET} # ⚠️ Env variable
  depends_on:
    - auth_db
  networks:
    - backend_network
    - db_network
  # ❌ Brak cap_drop
  # ❌ Brak security_opt (no-new-privileges)
  # ❌ Brak read_only filesystem
  # ❌ Brak resource limits (CPU, memory, PIDs)
  # ❌ Brak AppArmor profile
  # ❌ Brak healthcheck
```

**Porty exposed na wszystkich interfejsach:**

```yaml
ports:
  - "8080:8080" # ❌ Równoważne 0.0.0.0:8080
  - "3001:3000" # ❌ Exposed externally
  - "5433:5432" # ❌ PostgreSQL accessible from network
```

### 7.3 Nginx - Brak Security Headers

```nginx
# nginx.conf - STAN WYJŚCIOWY
events {
    worker_connections 1024;
}

http {
    server {
        listen 8080;
        # ❌ Brak server_tokens off;
        # ❌ Brak security headers (CSP, X-Frame-Options, etc.)
        # ❌ Brak rate limiting
        # ❌ Brak access control

        location /api/auth/ {
            proxy_pass http://auth:8080/auth/;
            # ❌ Minimalne proxy headers
        }
    }
}
```

### 7.4 CVE i Podatności

**Stan wyjściowy (przed security-fixes):**

| Źródło               | Liczba CVE | Severity                     |
| -------------------- | ---------- | ---------------------------- |
| **Go 1.22.5**        | 9 CVE      | 1 CRITICAL, 5 HIGH, 3 MEDIUM |
| **Alpine :latest**   | 3 CVE      | 2 HIGH, 1 MEDIUM             |
| **NPM dependencies** | 0          | -                            |

**Szczegóły Go CVE:**

- GO-2025-4006 (CRITICAL): RCE via encoding/gob
- GO-2025-4015 (HIGH): TLS handshake vulnerability
- GO-2025-4014 (HIGH): HTTP request smuggling
- ... i 6 innych

**OWASP ZAP Scan:**

- 1 MEDIUM: Missing Content Security Policy
- 3 LOW: Server version leak, missing headers
- 1 INFO: Cacheable sensitive content

---

## 8. Limitacje i Słabe Punkty

### 8.1 Bezpieczeństwo Runtime

| Problem                    | Impact                | CWE     |
| -------------------------- | --------------------- | ------- |
| **Kontenery jako root**    | Privilege escalation  | CWE-250 |
| **Brak capabilities drop** | Excessive permissions | CWE-250 |
| **Writable filesystem**    | Malware persistence   | CWE-732 |
| **Brak resource limits**   | DoS vulnerability     | CWE-400 |
| **Default Docker config**  | Weak isolation        | -       |

### 8.2 Network Security

| Problem                       | Impact            | Mitigacja           |
| ----------------------------- | ----------------- | ------------------- |
| **Ports na 0.0.0.0**          | External exposure | Bind do 127.0.0.1   |
| **Brak TLS między serwisami** | MITM attacks      | Implement mTLS      |
| **HTTP only (no HTTPS)**      | Packet sniffing   | Add TLS termination |

### 8.3 Secrets Management

**Problemy:**

```yaml
environment:
  - DB_PASSWORD=auth_password # ⚠️ Plain text w docker-compose
  - JWT_SECRET=${JWT_SECRET} # ⚠️ .env file (committed?)
  - INTERNAL_API_KEY=api_key # ⚠️ Hardcoded
```

**Recommendations:**

- ❌ Brak Docker Secrets
- ❌ Brak Vault/external secret store
- ❌ Secrets w environment variables (widoczne w `docker inspect`)

### 8.4 Monitoring & Logging

**Stan wyjściowy:**

- ✅ Structured logging (zap)
- ❌ Brak centralized logging (Loki)
- ❌ Brak metrics collection (Prometheus)
- ❌ Brak alerting
- ❌ Brak audit logs (security events)

### 8.5 Development vs Production

**Brak rozróżnienia:**

```yaml
# Jedna konfiguracja docker-compose.yml
# Brak docker-compose.prod.yml
# Brak różnych ustawień dla dev/prod
```

---

## 9. Metryki i Statystyki

### 9.1 Rozmiary Komponentów

| Komponent        | Rozmiar | Warstwy | Base Image                         |
| ---------------- | ------- | ------- | ---------------------------------- |
| **Backend (Go)** | ~45 MB  | 8-10    | golang:1.22-alpine → alpine:latest |
| **Frontend**     | ~1.8 GB | 15      | node:18-alpine                     |
| **Nginx**        | ~40 MB  | -       | nginx:alpine                       |
| **PostgreSQL**   | ~350 MB | -       | postgres:13                        |

**Total image size:** ~2.5 GB

### 9.2 Complexity Metrics

**Linie kodu (przybliżenie):**

- Backend (Go): ~15,000 LOC
- Frontend (TypeScript/React): ~25,000 LOC
- SQL schemas: ~500 LOC
- Config files: ~1,000 LOC

**Endpoints:**

- Auth: ~15 endpoints
- Quiz: ~25 endpoints
- Stats: ~10 endpoints
- Images: ~5 endpoints
- Admin: ~20 endpoints
- **Total:** ~75 REST endpoints

### 9.3 Database Metrics

| Database  | Tables | Estimated Rows (production)       |
| --------- | ------ | --------------------------------- |
| auth_db   | 8      | ~10,000 users                     |
| quiz_db   | 15     | ~500 questions, ~100,000 sessions |
| stats_db  | 3      | ~1,000,000 responses              |
| images_db | 1      | ~1,500 images                     |

---

## 10. Użyte Technologie - Stack Technologiczny

### 10.1 Backend Stack

```
Go 1.22.5
├── Web Framework: net/http (stdlib)
├── Router: gorilla/mux (implicit via handlers)
├── Database: lib/pq (PostgreSQL driver)
├── Validation: go-playground/validator
├── JWT: golang-jwt/jwt/v5
├── CORS: rs/cors
├── Logging: zap (structured logging)
└── Crypto: golang.org/x/crypto (bcrypt)
```

### 10.2 Frontend Stack

```
Next.js 14.2.5 (React 18.3.1)
├── UI Framework: Material-UI (MUI) 5.16
├── Styling: Emotion (CSS-in-JS)
├── Forms: Formik 2.4 + Yup validation
├── HTTP Client: Axios 1.7
├── Charts: Recharts 2.12
├── Auth: Custom (JWT cookies)
└── TypeScript: 5.5.4
```

### 10.3 Infrastructure Stack

```
Docker 24.x + Docker Compose
├── Reverse Proxy: Nginx (Alpine)
├── Database: PostgreSQL 13
├── Container Network: Bridge driver
└── Volumes: Named volumes (persistent data)
```

---

## 11. Przypadki Użycia (Use Cases)

### UC1: Student rozpoczyna quiz edukacyjny

**Aktorzy:** Student (authenticated user)

**Prekondycje:**

- Użytkownik zalogowany
- Wypełniona ankieta demograficzna

**Flow:**

1. Student wybiera tryb "Educational"
2. System tworzy sesję quizu (`quiz_sessions`)
3. System losuje pierwsze pytanie z grupy
4. Student analizuje parametry laboratoryjne i obrazy
5. Student wybiera diagnozę (opcję)
6. System sprawdza poprawność i wyświetla feedback
7. System przechodzi do następnego pytania
8. Student kończy quiz → podsumowanie

**Postkondycje:**

- Sesja zapisana w `stats_db`
- Statystyki zaktualizowane

### UC2: Nauczyciel tworzy test dla grupy

**Aktorzy:** Nauczyciel (role: teacher)

**Prekondycje:**

- Użytkownik ma rolę "teacher" lub "admin"

**Flow:**

1. Nauczyciel loguje się do panelu admin
2. Przechodzi do sekcji "Create Test"
3. Wybiera pytania z puli (multi-select)
4. Definiuje kod testu (np. "MED-2025-FINAL")
5. Definiuje nazwę testu
6. System tworzy test w `quiz_db.tests`
7. System generuje link: `/quiz?mode=test&code=MED-2025-FINAL`
8. Nauczyciel udostępnia link studentom

**Postkondycje:**

- Test dostępny dla studentów z kodem

### UC3: Admin analizuje statystyki

**Aktorzy:** Administrator (role: admin)

**Prekondycje:**

- Użytkownik ma rolę "admin"

**Flow:**

1. Admin loguje się do panelu admin
2. Wybiera sekcję "Statistics"
3. System agreguje dane z `stats_db`:
   - Accuracy per question
   - User rankings
   - Time metrics
   - Demographic breakdowns
4. Admin eksportuje dane do CSV/Excel
5. Admin identyfikuje trudne pytania
6. Admin może edytować/usunąć błędne pytania

**Postkondycje:**

- Eksportowane dane dostępne

---

## 12. Wnioski - Stan Przed Zabezpieczeniami

### 12.1 Pozytywne Aspekty

✅ **Architektura:**

- Dobrze zaprojektowana microservice architecture
- Separation of concerns (auth, quiz, stats)
- Scalable design (możliwość horizontal scaling)

✅ **Development Best Practices:**

- Multi-stage Docker builds
- Non-root users w kontenerach
- Health checks dla większości serwisów
- Structured logging (zap)
- Input validation (validator)

✅ **Database Design:**

- Normalized schema
- Foreign keys i constraints
- Indexes na często używanych kolumnach

### 12.2 Krytyczne Problemy Bezpieczeństwa

🔴 **HIGH SEVERITY:**

1. **9 CVE w Go standard library** (w tym 1 CRITICAL RCE)
2. **Brak security hardening kontenerów** (capabilities, read-only FS)
3. **Secrets w plain text** (passwords, JWT secrets w env vars)
4. **Porty exposed na 0.0.0.0** (dostępne z zewnątrz)
5. **Brak security headers** (CSP, X-Frame-Options)

🟡 **MEDIUM SEVERITY:**

1. Brak rate limiting (DDoS vulnerability)
2. Brak TLS między serwisami (MITM)
3. Brak resource limits (DoS)
4. Brak centralized logging/monitoring
5. Brak automated security scanning

### 12.3 Rekomendacje Priorytetowe

**Priorytet 1 (Critical):**

1. ✅ Update Go 1.22 → 1.24 (fix CVE)
2. ✅ Implement container hardening (capabilities, read-only FS)
3. ✅ Add security headers (CSP, etc.)
4. ✅ Localhost port binding (127.0.0.1)
5. ✅ Pin Alpine versions (fix CVE-2025-9230+)

**Priorytet 2 (High):**

1. Migrate to Docker Secrets
2. Add resource limits (CPU, memory, PIDs)
3. Implement monitoring (Prometheus + Grafana)
4. Add automated security scanning (Trivy, Grype)

**Priorytet 3 (Medium):**

1. Implement TLS/mTLS
2. Add rate limiting (nginx)
3. Implement audit logging
4. Add SIEM integration

---

## 13. Metryki Compliance - Stan Wyjściowy

| Standard                 | Score        | Komentarz                                   |
| ------------------------ | ------------ | ------------------------------------------- |
| **CIS Docker Benchmark** | 4/117 (3.4%) | Tylko podstawowe kontrole (non-root user)   |
| **OWASP Top 10**         | 3/10 covered | A05 (Security Misconfiguration) NOT covered |
| **NIST Cybersecurity**   | 20%          | Brak continuous monitoring                  |
| **ISO 27001 A.12.6**     | Partial      | Brak vulnerability management process       |

---

## 14. Podsumowanie

**PrediGrowee** w stanie wyjściowym (gałąź `main`) był **funkcjonalnym systemem edukacyjnym** z dobrą architekturą microservices, ale z **istotnymi lukami w bezpieczeństwie**:

✅ **Strengths:**

- Solid architecture (microservices + separation of concerns)
- Non-root users w kontenerach (baseline security)
- Multi-stage builds (smaller images)
- Input validation i prepared statements (SQL injection mitigation)

❌ **Critical Weaknesses:**

- **9 Critical/High CVE** w Go standard library
- **Brak runtime security hardening** (capabilities, read-only FS, resource limits)
- **Secrets exposure** (plain text passwords w config)
- **Missing security headers** (CSP, X-Frame-Options)
- **Network exposure** (ports na 0.0.0.0)
- **Brak automated security testing**

**Następne kroki:** Implementacja zabezpieczeń opisanych w rozdziałach 4-5 (security hardening).

---

**Dokumentacja stanu wyjściowego:** ✅ COMPLETE
**Kolejny rozdział:** Implementacja zabezpieczeń (gałąź `security-fixes`)
**Data:** Listopad 2025
