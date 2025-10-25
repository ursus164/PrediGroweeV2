# Quick Start: Monitoring Stack

## 1. Utwórz sieć monitoring

```bash
docker network create predigroweev2_monitoring
```

## 2. Uruchom monitoring stack

```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

## 3. Uruchom główną aplikację

```bash
docker-compose up -d
```

## 4. Dostęp do Grafana

Otwórz w przeglądarce: **http://localhost:3002**

- Login: `admin`
- Hasło: `admin123`

##  5. Zobacz logi w czasie rzeczywistym

1. W Grafana przejdź do **Explore**
2. Wybierz datasource **Loki**
3. Użyj query:

```logql
{service="auth"} |= "error"
```

## 6. Zobacz metryki

1. W Grafana przejdź do **Explore**
2. Wybierz datasource **Prometheus**
3. Użyj query:

```promql
rate(container_cpu_usage_seconds_total[5m])
```

## Usługi

- **Grafana**: http://localhost:3002 - Dashboardy
- **Prometheus**: internal:9090 - Metryki
- **Loki**: internal:3100 - Logi

## Zatrzymanie

```bash
# Zatrzymaj monitoring
docker-compose -f docker-compose.monitoring.yml down

# Zatrzymaj aplikację
docker-compose down
```

## Więcej informacji

Zobacz [MONITORING_SETUP.md](./MONITORING_SETUP.md) dla pełnej dokumentacji.
