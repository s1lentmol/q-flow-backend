# Auth service

Микросервис авторизации на gRPC (Go + Fiber для других сервисов). Хранит пользователей и приложения в Postgres, миграции — через goose.

## Конфигурация

- Настройки лежат в `config/config.yaml`, переменные можно переопределить через ENV (`ENV_DB_USERNAME`, `ENV_DB_PASSWORD`, `ENV_DB_HOST`, `ENV_DB_PORT`, `ENV_DB_NAME`, `ENV_GRPC_PORT`, `ENV_TOKEN_TTL`, `ENV_ENV`).
- DSN формируется как `postgres://user:pass@host:port/db?sslmode=disable`.

## Миграции (goose)

1. Установите утилиту: `go install github.com/pressly/goose/v3/cmd/goose@latest`.
2. Запустите миграции для auth:  
   `goose -dir services/auth/migrations postgres "postgres://user:pass@localhost:5432/postgres?sslmode=disable" up`
3. Для отката используйте `down` или `reset`.

Таблицы: `users` (email, pass_hash, is_admin) и `apps` (name, secret). Добавьте запись в `apps` вручную, чтобы логин выдавал JWT (app_id передается в Login).

## Запуск

```bash
cd services/auth
go run ./cmd/auth
```

Сервис стартует на порту из `config.yaml` (по умолчанию 44044). Остановка по SIGINT/SIGTERM.
