# Queue service

gRPC микросервис бизнес-логики очередей. Работает с Postgres, миграции через goose, фильтрует очереди по `group_code`.

## Конфигурация

Файл `config/config.yaml`, можно переопределять через ENV с префиксом `ENV_` (как в auth):  
`ENV_DB_USERNAME`, `ENV_DB_PASSWORD`, `ENV_DB_HOST`, `ENV_DB_PORT`, `ENV_DB_NAME`, `ENV_GRPC_PORT`, `ENV_ENV`.

DSN формируется как `postgres://user:pass@host:port/db?sslmode=disable`.

## Миграции

Миграции вшиты в бинарь (`services/queue/migrations`). При старте сервиса вызывается `goose.Up`, поэтому достаточно запустить сервис на доступной БД. Локально можно прогнать вручную:

```bash
goose -dir services/queue/migrations postgres "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up
```

Создаются таблицы `queues`, `queue_participants` и enum `queue_mode/queue_status`.

## Прото

`protos/proto/queue/queue.proto`, go-код в `protos/gen/go/queue`. Команда генерации: `make generate-proto-queue`.

## Запуск

```bash
cd services/queue
go run ./cmd/queue
```

gRPC по умолчанию на `:44045`.
