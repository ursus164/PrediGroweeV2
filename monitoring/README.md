# PrediGrowee Monitoring & Security Stack

Kompletny stack monitoringu i security logging dla aplikacji PrediGrowee z wykorzystaniem Prometheus, Grafana, Loki i Promtail.

## 📋 Spis treści

- [Przegląd](#przegląd)
- [Wymagania](#wymagania)
- [Instalacja](#instalacja)
- [Uruchamianie](#uruchamianie)
- [Dostęp do interfejsów](#dostęp-do-interfejsów)
- [Konfiguracja](#konfiguracja)
- [Użytkowanie](#użytkowanie)
- [Troubleshooting](#troubleshooting)
- [Bezpieczeństwo](#bezpieczeństwo)

## 🎯 Przegląd

### Komponenty

| Komponent         | Funkcja                           | Port            |
| ----------------- | --------------------------------- | --------------- |
| **Prometheus**    | Zbieranie i przechowywanie metryk | 9090 (internal) |
| **Grafana**       | Wizualizacja i dashboardy         | 3002            |
| **Loki**          | Agregacja logów                   | 3100 (internal) |
| **Promtail**      | Zbieranie logów z kontenerów      | 9080 (internal) |
| **Node Exporter** | Metryki hosta                     | 9100 (internal) |
| **cAdvisor**      | Metryki kontenerów                | 8080 (internal) |

### Funkcje

✅ **Monitoring bezpieczeństwa**

- Śledzenie nieudanych prób logowania
- Monitoring błędów HTTP (4xx/5xx)
- Wykrywanie nieautoryzowanego dostępu
- Analiza wzorców zagrożeń

✅ **Monitoring wydajności**

- CPU i pamięć per kontener
- Ruch sieciowy
- Czasy odpowiedzi
- Wykorzystanie zasobów

✅ **Agregacja logów**

- Streaming logów w czasie rzeczywistym
- Filtrowanie zdarzeń bezpieczeństwa
- Wyszukiwanie i analiza
- Retention 31 dni

## 📦 Wymagania

- Docker 20.10+
- Docker Compose 2.0+
- 2GB wolnej pamięci RAM dla stacku monitoringu
- 10GB miejsca na dysku dla przechowywania logów i metryk

## 🚀 Instalacja

### Krok 1: Sklonuj repozytorium

```bash
git clone https://github.com/ursus164/PrediGroweeV2.git
cd PrediGroweeV2
```

### Krok 2: Utwórz sieć Docker

```bash
docker network create predigroweev2_monitoring
```

### Krok 3: Sprawdź konfigurację

Zweryfikuj, że pliki konfiguracyjne istnieją:

```bash
ls -la monitoring/prometheus/prometheus.yml
ls -la monitoring/loki/loki-config.yml
ls -la monitoring/promtail/promtail-config.yml
ls -la monitoring/grafana/provisioning/
```

### Krok 4: (Opcjonalnie) Zmień hasło Grafana

Edytuj `docker-compose.monitoring.yml`:

```yaml
grafana:
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=twoje_bezpieczne_haslo
```

## ▶️ Uruchamianie

### Uruchomienie stacku monitoringu

```bash
# Uruchom stack monitoringu
docker-compose -f docker-compose.monitoring.yml up -d

# Sprawdź status
docker-compose -f docker-compose.monitoring.yml ps
```

Powinieneś zobaczyć wszystkie serwisy w stanie `Up`:

```
NAME                COMMAND             STATUS          PORTS
prometheus          --config.file...    Up              9090/tcp
grafana             /run.sh             Up              0.0.0.0:3002->3000/tcp
loki                -config.file...     Up              3100/tcp
promtail            -config.file...     Up              9080/tcp
node-exporter       --path.rootfs...    Up              9100/tcp
cadvisor            /usr/bin/cadv...    Up (healthy)    8080/tcp
```

### Uruchomienie aplikacji głównej

```bash
# Uruchom aplikację PrediGrowee
docker-compose up -d

# Sprawdź status
docker-compose ps
```

### Sprawdzenie logów

```bash
# Logi monitoring stack
docker-compose -f docker-compose.monitoring.yml logs -f

# Logi konkretnego serwisu
docker-compose -f docker-compose.monitoring.yml logs -f grafana
docker-compose -f docker-compose.monitoring.yml logs -f prometheus
```

## 🌐 Dostęp do interfejsów

### Grafana - Dashboard główny

**URL:** http://localhost:3002

**Domyślne dane logowania:**

- Username: `admin`
- Password: `admin123`

**⚠️ WAŻNE:** Zmień hasło przy pierwszym logowaniu!

### Pierwsze kroki w Grafana

1. **Zaloguj się** do Grafana
2. **Explore** → Wybierz datasource
3. **Import dashboard** → Użyj ID: 1860 (Node Exporter Full)

### Prometheus (dostęp wewnętrzny)

Prometheus jest dostępny tylko wewnątrz sieci Docker. Aby uzyskać dostęp:

```bash
# Z localhost (edytuj docker-compose.monitoring.yml i dodaj ports)
ports:
  - "9090:9090"

# Lub przez kontener
docker-compose -f docker-compose.monitoring.yml exec grafana curl http://prometheus:9090/-/healthy
```

## ⚙️ Konfiguracja

### Dodanie własnych metryk do serwisów Go

1. Dodaj Prometheus client do Go service:

```go
// main.go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
}

func main() {
    // ... twój kod ...

    // Endpoint dla metryk
    http.Handle("/metrics", promhttp.Handler())

    // ... reszta kodu ...
}
```

2. Użyj metryki w handlerach:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
    // ... logika handlera ...
}
```

### Konfiguracja retention (przechowywanie danych)

**Prometheus** - edytuj `monitoring/prometheus/prometheus.yml`:

```yaml
global:
  evaluation_interval: 15s
  scrape_interval: 15s
```

W `docker-compose.monitoring.yml`:

```yaml
prometheus:
  command:
    - "--storage.tsdb.retention.time=30d" # Zmień na 90d dla 3 miesięcy
```

**Loki** - edytuj `monitoring/loki/loki-config.yml`:

```yaml
limits_config:
  retention_period: 744h # 31 dni - zmień na 2160h dla 90 dni
```

## 📊 Użytkowanie

### Przykładowe zapytania Prometheus (PromQL)

Użyj w Grafana → Explore → Prometheus

#### 1. CPU usage per container

```promql
sum(rate(container_cpu_usage_seconds_total{name!=""}[5m])) by (name) * 100
```

#### 2. Memory usage per container (MB)

```promql
sum(container_memory_usage_bytes{name!=""}) by (name) / 1024 / 1024
```

#### 3. HTTP error rate (5xx)

```promql
sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
```

#### 4. Failed authentication attempts

```promql
rate(auth_failed_attempts_total[5m])
```

#### 5. Network traffic RX/TX

```promql
# Received
rate(container_network_receive_bytes_total[5m])

# Transmitted
rate(container_network_transmit_bytes_total[5m])
```

### Przykładowe zapytania Loki (LogQL)

Użyj w Grafana → Explore → Loki

#### 1. Wszystkie logi z serwisu auth

```logql
{service="auth"}
```

#### 2. Logi z błędami

```logql
{service=~".*"} | json | level="error"
```

#### 3. Failed login attempts

```logql
{service="auth"} |= "failed" |= "login"
```

#### 4. Unauthorized access (401)

```logql
{service=~"auth|admin|nginx"} | json | status_code="401"
```

#### 5. Suspicious patterns

```logql
{service="auth"} |~ "(?i)(sql injection|xss|unauthorized|forbidden)"
```

#### 6. Error rate w ostatnich 5 minutach

```logql
sum(rate({service=~".*"} | json | level="error" [5m])) by (service)
```

#### 7. Slow requests (>1s)

```logql
{service=~".*"} | json | duration > 1s
```

### Tworzenie dashboardów w Grafana

#### Krok 1: Nowy dashboard

1. Grafana → **Dashboards** → **New Dashboard**
2. **Add visualization**
3. Wybierz datasource (Prometheus lub Loki)

#### Krok 2: Panel z metrykami

**Przykład: CPU Usage**

- Query: `sum(rate(container_cpu_usage_seconds_total[5m])) by (name)`
- Visualization: Time series
- Unit: Percent (0-100)
- Legend: `{{name}}`

#### Krok 3: Panel z logami

**Przykład: Recent Errors**

- Query: `{service=~".*"} | json | level="error"`
- Visualization: Logs
- Options: Show time, Show labels

#### Krok 4: Zapisz dashboard

- **Save dashboard**
- Nadaj nazwę: "PrediGrowee Overview"
- Dodaj do folderu: "General"

### Import gotowych dashboardów

1. Grafana → **Dashboards** → **Import**
2. Wpisz ID dashboardu:

   - **1860** - Node Exporter Full
   - **893** - Docker and Host Monitoring
   - **13639** - Loki Dashboard Quick Search
   - **12633** - Docker Container & Host Metrics

3. Wybierz datasource: Prometheus
4. Kliknij **Import**

## 🔍 Troubleshooting

### Problem: Prometheus nie zbiera metryk

**Sprawdź targets:**

```bash
docker-compose -f docker-compose.monitoring.yml exec prometheus wget -O- http://localhost:9090/api/v1/targets | jq
```

**Sprawdź connectivity:**

```bash
docker-compose -f docker-compose.monitoring.yml exec prometheus wget -O- http://auth:8080/metrics
```

**Rozwiązanie:** Upewnij się, że serwisy są w tej samej sieci (`monitoring`)

### Problem: Promtail nie zbiera logów

**Sprawdź uprawnienia do Docker socket:**

```bash
docker-compose -f docker-compose.monitoring.yml exec promtail ls -la /var/run/docker.sock
```

**Sprawdź logi Promtail:**

```bash
docker-compose -f docker-compose.monitoring.yml logs promtail
```

**Rozwiązanie:** Sprawdź, czy labels `logging: "promtail"` są dodane do serwisów

### Problem: Grafana nie łączy się z datasource

**Sprawdź status datasource:**

```bash
docker-compose -f docker-compose.monitoring.yml exec grafana curl http://prometheus:9090/-/healthy
docker-compose -f docker-compose.monitoring.yml exec grafana curl http://loki:3100/ready
```

**Sprawdź konfigurację:**

```bash
cat monitoring/grafana/provisioning/datasources/datasources.yml
```

**Rozwiązanie:** Zrestartuj Grafana

```bash
docker-compose -f docker-compose.monitoring.yml restart grafana
```

### Problem: Wysoka utilization dysków

**Sprawdź rozmiar volumes:**

```bash
docker system df -v | grep prometheus
docker system df -v | grep loki
docker system df -v | grep grafana
```

**Wyczyść stare dane:**

```bash
# Prometheus
docker-compose -f docker-compose.monitoring.yml exec prometheus rm -rf /prometheus/wal/*

# Loki
docker-compose -f docker-compose.monitoring.yml exec loki rm -rf /loki/chunks/*
```

**Lub zmniejsz retention:**

Edytuj `monitoring/loki/loki-config.yml`:

```yaml
limits_config:
  retention_period: 168h # 7 dni zamiast 31
```

### Problem: Kontener ciągle się restartuje

**Sprawdź logi:**

```bash
docker-compose -f docker-compose.monitoring.yml logs --tail=50 <service-name>
```

**Sprawdź zasoby:**

```bash
docker stats
```

**Rozwiązanie:** Zwiększ limity w `docker-compose.monitoring.yml`

## 🔒 Bezpieczeństwo

### Zmiana domyślnego hasła Grafana

**Przez interfejs:**

1. Zaloguj się → User icon → Change password

**Przez environment variable:**

Edytuj `docker-compose.monitoring.yml`:

```yaml
grafana:
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=nowe_silne_haslo_123!
```

```bash
docker-compose -f docker-compose.monitoring.yml up -d grafana
```

### Ograniczenie dostępu do Grafana

**Firewall (UFW):**

```bash
# Allow only from localhost
sudo ufw allow from 127.0.0.1 to any port 3002

# Allow from specific IP
sudo ufw allow from 192.168.1.100 to any port 3002
```

**Reverse proxy z Nginx + TLS:**

```nginx
server {
    listen 443 ssl;
    server_name monitoring.predigrowee.com;

    ssl_certificate /etc/ssl/certs/cert.pem;
    ssl_certificate_key /etc/ssl/private/key.pem;

    location / {
        proxy_pass http://localhost:3002;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Security hardening obecny w stacku

✅ Wszystkie kontenery z `cap_drop: ALL`
✅ `no-new-privileges: true` dla wszystkich
✅ `read_only: true` filesystem z tmpfs
✅ Resource limits (CPU/memory)
✅ Non-root user dla Grafana (uid: 472)

### Regularne przeglądy security logs

**Utwórz cron job:**

```bash
# /etc/cron.daily/security-review.sh
#!/bin/bash

docker-compose -f /path/to/docker-compose.monitoring.yml exec -T grafana \
  curl -s "http://localhost:3002/api/ds/query" \
  -H "Content-Type: application/json" \
  -d '{
    "queries": [{
      "refId": "A",
      "datasource": "Loki",
      "expr": "{service=\"auth\",alert=\"true\"}"
    }]
  }' | jq '.results[] | .frames[] | .data.values' > /var/log/security-review.log
```

## 📚 Dodatkowe zasoby

### Dokumentacja

- [MONITORING_SETUP.md](./MONITORING_SETUP.md) - Pełna dokumentacja techniczna
- [MONITORING_QUICKSTART.md](./MONITORING_QUICKSTART.md) - Szybki start
- [SECURITY_HARDENING.md](./SECURITY_HARDENING.md) - Docker security hardening

### Linki zewnętrzne

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [PromQL Cheat Sheet](https://promlabs.com/promql-cheat-sheet/)
- [LogQL Cheat Sheet](https://grafana.com/docs/loki/latest/logql/)

## 🛠️ Maintenance

### Backup danych

```bash
# Backup Prometheus data
docker run --rm -v predigroweev2_prometheus_data:/data -v $(pwd):/backup ubuntu tar czf /backup/prometheus-backup-$(date +%Y%m%d).tar.gz -C /data .

# Backup Grafana dashboards
docker run --rm -v predigroweev2_grafana_data:/data -v $(pwd):/backup ubuntu tar czf /backup/grafana-backup-$(date +%Y%m%d).tar.gz -C /data .

# Backup Loki data
docker run --rm -v predigroweev2_loki_data:/data -v $(pwd):/backup ubuntu tar czf /backup/loki-backup-$(date +%Y%m%d).tar.gz -C /data .
```

### Restore z backupu

```bash
# Restore Prometheus
docker run --rm -v predigroweev2_prometheus_data:/data -v $(pwd):/backup ubuntu tar xzf /backup/prometheus-backup-20251025.tar.gz -C /data

# Restart services
docker-compose -f docker-compose.monitoring.yml restart prometheus
```

### Update kontenerów

```bash
# Pull latest images
docker-compose -f docker-compose.monitoring.yml pull

# Restart with new images
docker-compose -f docker-compose.monitoring.yml up -d
```

## 🚦 Status i health checks

### Sprawdzenie zdrowia wszystkich serwisów

```bash
# Status wszystkich kontenerów
docker-compose -f docker-compose.monitoring.yml ps

# Health check Prometheus
curl http://localhost:9090/-/healthy

# Health check Loki
curl http://localhost:3100/ready

# Health check Grafana
curl http://localhost:3002/api/health
```

### Monitoring zdrowia z zewnętrznego narzędzia

```bash
#!/bin/bash
# healthcheck.sh

services=("prometheus:9090/-/healthy" "loki:3100/ready" "grafana:3002/api/health")

for service in "${services[@]}"; do
    IFS=':' read -r name endpoint <<< "$service"
    status=$(docker-compose -f docker-compose.monitoring.yml exec -T $name curl -s -o /dev/null -w "%{http_code}" "http://localhost:$endpoint")

    if [ "$status" = "200" ]; then
        echo "✅ $name is healthy"
    else
        echo "❌ $name is unhealthy (HTTP $status)"
    fi
done
```

## 📞 Support

W przypadku problemów:

1. Sprawdź [Troubleshooting](#troubleshooting)
2. Zobacz logi: `docker-compose -f docker-compose.monitoring.yml logs`
3. Otwórz issue na GitHub: https://github.com/ursus164/PrediGroweeV2/issues

## 📝 Licencja

MIT License - Zobacz [LICENSE](../LICENSE) dla szczegółów.

---

**Ostatnia aktualizacja:** 2025-10-25
**Wersja:** 1.0.0
