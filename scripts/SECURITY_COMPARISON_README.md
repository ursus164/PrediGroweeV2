# Security Comparison Script

Automatyczny skrypt porównujący zabezpieczenia między branch `main` (przed zmianami) a `security-fixes` (po zmianach).

## 🛠️ Narzędzia używane

### 1. **CIS Docker Benchmark** (Docker Bench Security)

- **Waga:** 40%
- **Co testuje:** Zgodność z CIS Docker Benchmark v1.6.0
- **Obszary:** Host configuration, Docker daemon, images, containers, security operations
- **Źródło:** https://github.com/docker/docker-bench-security

### 2. **OWASP ZAP** (Zed Attack Proxy)

- **Waga:** 40%
- **Co testuje:** OWASP Top 10 vulnerabilities
- **Typ:** Automated baseline security scan
- **Źródło:** https://www.zaproxy.org/

### 3. **Custom Security Checks**

- **Waga:** 20%
- **Co testuje:**
  - Container hardening (cap_drop, read_only, security_opt)
  - Non-root user configuration
  - Secrets management
  - Network isolation
  - Privileged containers
  - Image tags

## 🚀 Jak uruchomić

```bash
cd /home/ursus/personal/engineer/PrediGroweeV2
sudo ./scripts/security-comparison.sh
```

⚠️ **Wymagania:**

- Docker i docker-compose zainstalowane
- Uprawnienia sudo (dla Docker Bench Security)
- Wolne porty 8080, 3000 (dla OWASP ZAP scan)
- ~10-15 minut na pełny test

## 📊 Interpretacja wyników

### Scoring System

- **A (90-100):** Excellent - Production ready
- **B (80-89):** Good - Minor improvements needed
- **C (70-79):** Fair - Some hardening recommended
- **D (60-69):** Poor - Significant issues present
- **F (0-59):** Fail - Critical security gaps

### Wagi testów

```
CIS Docker Benchmark:    40% (infrastructure security)
OWASP ZAP:              40% (application security)
Custom Security Checks: 20% (configuration hardening)
```

## 📁 Output

Skrypt generuje:

1. **Raport końcowy:** `security-comparison-report/FINAL_REPORT_[timestamp].md`
2. **Szczegółowe logi:**
   - `main_[timestamp]/` - wyniki dla branch main
   - `security-fixes_[timestamp]/` - wyniki dla branch security-fixes

### Zawartość raportów:

- `cis-docker-benchmark.txt` - pełny output CIS tests
- `owasp-zap.txt` - tekstowy log OWASP ZAP
- `owasp-zap-report.html` - HTML raport OWASP ZAP
- `owasp-zap-report.json` - JSON raport OWASP ZAP
- `custom-security-checks.txt` - własne testy konfiguracji

## 🎯 Co pokazuje comparison

### Przykładowy output:

```
╔═══════════════════════════════════════════════════════════╗
║                  COMPARISON COMPLETE                      ║
╠═══════════════════════════════════════════════════════════╣
║  main:           65/100 (C - Fair)                        ║
║  security-fixes: 87/100 (B - Good)                        ║
║  Improvement:    +22 points (+33%)                        ║
╚═══════════════════════════════════════════════════════════╝
```

### Metrics tracked:

| Metric                | Description                      |
| --------------------- | -------------------------------- |
| CIS PASS/WARN/INFO    | Docker infrastructure compliance |
| OWASP HIGH/MEDIUM/LOW | Application vulnerability count  |
| Hardening score       | Container security configuration |
| Non-root users        | Service user configuration       |
| Secrets management    | Credential handling              |

## 🔧 Troubleshooting

### "Docker not available"

```bash
sudo systemctl start docker
```

### "Permission denied"

```bash
sudo chmod +x scripts/security-comparison.sh
```

### "Port already in use"

```bash
docker-compose down -v
sudo lsof -ti:8080 | xargs kill -9
```

### "Application not responding"

- Zwiększ sleep time w skrypcie (linia: `sleep 30` → `sleep 60`)
- Sprawdź logi: `docker-compose logs`

## 📚 Referencje

- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP ZAP Documentation](https://www.zaproxy.org/docs/)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)

## 🤝 Contributing

Aby dodać nowe testy:

1. Dodaj test w funkcji `test_branch()`
2. Aktualizuj scoring w `calculate_final_score()`
3. Rozszerz raport w `generate_final_report()`

## 📝 License

Part of PrediGroweeV2 project - Educational/Internal use
