# Database Initialization

## Konfiguracja lokalna

1. Umieść swoje dumpy SQL w odpowiednich katalogach:

   - `auth/` - dump bazy auth_db
   - `quiz/` - dump bazy quiz_db
   - `stats/` - dump bazy stats_db
   - `images/` - dump bazy images_db

2. PostgreSQL automatycznie zaimportuje wszystkie pliki `.sql` z tych katalogów przy pierwszym uruchomieniu kontenera

3. Aby zresetować bazy danych:
   ```bash
   docker compose down -v  # usuwa volume z danymi
   docker compose up -d    # tworzy nowe volume i importuje dumpy
   ```

## Struktura

```
database-init/
├── auth/       # pliki *.sql dla auth_db
├── quiz/       # pliki *.sql dla quiz_db
├── stats/      # pliki *.sql dla stats_db
└── images/     # pliki *.sql dla images_db
```

**Uwaga:** Pliki `*.sql` NIE SĄ commitowane do repozytorium
