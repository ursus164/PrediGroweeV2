# Raport naprawy problemów bezpieczeństwa OWASP ZAP

**Data:** 2 listopada 2025
**Wersja ZAP:** 2.16.1
**Target:** http://localhost:8080

## Podsumowanie wykrytych problemów

### Problemy przed naprawą:

- **High:** 0
- **Medium:** 1
- **Low:** 3
- **Informational:** 1

---

## Szczegóły naprawionych problemów

### 1. ✅ Content Security Policy (CSP) Header Not Set [MEDIUM]

**Problem:**
Brak nagłówka Content-Security-Policy zwiększa ryzyko ataków XSS i wstrzykiwania danych.

**Naprawa:**
Dodano nagłówek CSP do obu plików konfiguracyjnych nginx:

```nginx
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'self';" always;
```

**Lokalizacja:**

- `/PrediGroweeV2/nginx.conf` - linie 14-15, 69-70
- `/PrediGroweeV2/nginx.prod.conf` - linia 72

**Wyjaśnienie polityki:**

- `default-src 'self'` - domyślnie tylko zasoby z tej samej domeny
- `script-src 'self' 'unsafe-inline' 'unsafe-eval'` - skrypty z tej samej domeny i inline (potrzebne dla React/Next.js)
- `style-src 'self' 'unsafe-inline'` - style z tej samej domeny i inline
- `img-src 'self' data: https:` - obrazy z tej samej domeny, data URIs i HTTPS
- `connect-src 'self'` - połączenia API tylko do tej samej domeny
- `frame-ancestors 'self'` - strona może być umieszczona tylko we własnych frame'ach

---

### 2. ✅ Permissions Policy Header Not Set [LOW]

**Problem:**
Brak nagłówka Permissions-Policy umożliwia nieautoryzowany dostęp do funkcji przeglądarki (kamera, mikrofon, lokalizacja).

**Naprawa:**
Dodano nagłówek Permissions-Policy:

```nginx
add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;
```

**Lokalizacja:**

- `/PrediGroweeV2/nginx.conf` - linie 16, 71
- `/PrediGroweeV2/nginx.prod.conf` - linia 75

**Wyjaśnienie:**

- `geolocation=()` - wyłączona geolokalizacja
- `microphone=()` - wyłączony dostęp do mikrofonu
- `camera=()` - wyłączony dostęp do kamery

---

### 3. ✅ Server Leaks Version Information via "Server" HTTP Response Header Field [LOW]

**Problem:**
Serwer ujawniał wersję nginx (nginx/1.29.0) w nagłówku odpowiedzi HTTP, co ułatwia atakującym identyfikację podatności.

**Naprawa:**
Dodano dyrektywę ukrywającą informacje o wersji:

```nginx
server_tokens off;
```

**Lokalizacja:**

- `/PrediGroweeV2/nginx.conf` - linia 8
- `/PrediGroweeV2/nginx.prod.conf` - linia 28

**Efekt:**

- Zamiast `Server: nginx/1.29.0` nagłówek będzie zawierał tylko `Server: nginx`

---

### 4. ✅ In Page Banner Information Leak [LOW]

**Problem:**
Informacje o wersji serwera (nginx/1.29.0) pojawiały się w treści odpowiedzi dla stron błędów.

**Naprawa:**
Dyrektywa `server_tokens off;` usuwa również informacje o wersji z domyślnych stron błędów nginx.

**Lokalizacja:**

- Naprawione przez `server_tokens off;` w obu plikach konfiguracyjnych

---

### 5. ✅ Storable and Cacheable Content [INFORMATIONAL]

**Problem:**
Zawartość dynamiczna mogła być cachowana przez serwery proxy, co może prowadzić do wycieku wrażliwych danych.

**Naprawa:**
Dodano nagłówki kontroli cache dla zawartości dynamicznej:

```nginx
add_header Cache-Control "no-cache, no-store, must-revalidate" always;
add_header Pragma "no-cache" always;
add_header Expires "0" always;
```

**Lokalizacja:**

- `/PrediGroweeV2/nginx.conf` - linie 21-23
- `/PrediGroweeV2/nginx.prod.conf` - linie 78-80

**Wyjaśnienie:**

- `no-cache` - wymusza rewalidację przed użyciem zakeszowanej wersji
- `no-store` - zabrania przechowywania odpowiedzi w cache
- `must-revalidate` - wymusza sprawdzenie aktualności
- `Pragma: no-cache` - kompatybilność z HTTP/1.0
- `Expires: 0` - natychmiastowe wygaśnięcie

---

## Dodatkowe usprawnienia

### Dodatkowe nagłówki bezpieczeństwa

Dodano również standardowe nagłówki bezpieczeństwa:

```nginx
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
```

**Wyjaśnienie:**

- `X-Frame-Options: SAMEORIGIN` - zapobiega atakom clickjacking
- `X-Content-Type-Options: nosniff` - zapobiega MIME type sniffing
- `X-XSS-Protection: 1; mode=block` - włącza filtr XSS przeglądarki

### Ulepszone proxy headers

Dodano standardowe nagłówki proxy dla wszystkich lokalizacji:

```nginx
proxy_set_header Host $host;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $scheme;
```

---

## Pliki zmodyfikowane

1. `/PrediGroweeV2/nginx.conf`

   - Dodano `server_tokens off;`
   - Dodano nagłówki CSP, Permissions-Policy, Cache-Control
   - Dodano dodatkowe nagłówki bezpieczeństwa
   - Dodano proxy headers dla wszystkich lokalizacji

2. `/PrediGroweeV2/nginx.prod.conf`
   - Dodano `server_tokens off;`
   - Rozszerzono nagłówki CSP, Permissions-Policy, Cache-Control
   - Dodano nagłówek X-XSS-Protection

---

## Testowanie

Po wdrożeniu zmian, należy:

1. Zrestartować kontenery nginx:

   ```bash
   docker-compose restart nginx
   # lub dla produkcji
   docker-compose -f docker-compose.prod.yml restart nginx
   ```

2. Uruchomić ponownie skan OWASP ZAP, aby zweryfikować naprawę:

   ```bash
   # Przykład skanu
   zap-cli quick-scan http://localhost:8080
   ```

3. Sprawdzić nagłówki odpowiedzi:

   ```bash
   curl -I http://localhost:8080
   ```

   Oczekiwane nagłówki:

   - `Content-Security-Policy: ...`
   - `Permissions-Policy: ...`
   - `Cache-Control: no-cache, no-store, must-revalidate`
   - `X-Frame-Options: SAMEORIGIN`
   - `X-Content-Type-Options: nosniff`
   - `X-XSS-Protection: 1; mode=block`
   - `Server: nginx` (bez numeru wersji)

4. Sprawdzić działanie aplikacji:
   - Upewnić się, że frontend działa poprawnie
   - Sprawdzić, czy wszystkie API endpoints odpowiadają
   - Zweryfikować, czy nie ma problemów z CORS
   - Sprawdzić, czy zasoby statyczne (obrazy, style, skrypty) ładują się poprawnie

---

## Potencjalne problemy

### Content Security Policy

Jeśli aplikacja używa zewnętrznych zasobów (CDN, zewnętrzne API), może być konieczne dostosowanie polityki CSP:

- Dla zewnętrznych API: dodaj domenę do `connect-src`
- Dla Google Fonts: dodaj `https://fonts.googleapis.com https://fonts.gstatic.com` do `font-src`
- Dla zewnętrznych skryptów: dodaj domenę do `script-src`

### Rozwój lokalne

W środowisku deweloperskim może być potrzebne złagodzenie niektórych polityk, szczególnie dla Hot Module Replacement (HMR):

```nginx
# Development only - more permissive CSP
add_header Content-Security-Policy "default-src 'self' 'unsafe-inline' 'unsafe-eval'; connect-src 'self' ws: wss:;" always;
```

---

## Zgodność z standardami

Naprawione problemy są zgodne z:

- **OWASP Top 10 2021**
- **OWASP Application Security Verification Standard (ASVS)**
- **CWE-693** (Protection Mechanism Failure)
- **CWE-497** (Exposure of Sensitive System Information)
- **CWE-524** (Use of Cache Containing Sensitive Information)

---

## Referencje

- [MDN - Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
- [OWASP CSP Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html)
- [MDN - Permissions Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Permissions-Policy)
- [OWASP Secure Headers Project](https://owasp.org/www-project-secure-headers/)
- [nginx server_tokens documentation](https://nginx.org/en/docs/http/ngx_http_core_module.html#server_tokens)

---

## Status naprawy

| Alert                                        | Risk Level    | Status        | CWE |
| -------------------------------------------- | ------------- | ------------- | --- |
| Content Security Policy (CSP) Header Not Set | Medium        | ✅ Naprawiony | 693 |
| Permissions Policy Header Not Set            | Low           | ✅ Naprawiony | 693 |
| Server Leaks Version Information             | Low           | ✅ Naprawiony | 497 |
| In Page Banner Information Leak              | Low           | ✅ Naprawiony | 497 |
| Storable and Cacheable Content               | Informational | ✅ Naprawiony | 524 |

**Wszystkie wykryte problemy zostały naprawione! ✅**
