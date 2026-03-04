# Миграции БД

SQL-миграции для PostgreSQL. Используется [golang-migrate](https://github.com/golang-migrate/migrate).

## Установка CLI (опционально)

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Структура

Каждая миграция = 2 файла:
- `{version}_{name}.up.sql` - применение
- `{version}_{name}.down.sql` - откат

**000001_create_metrics_tables** - создает таблицы `gauges` и `counters`

## Создание миграции

```bash
migrate create -ext sql -dir ./migrations -seq название_фичи
```

## Применение

Автоматически при запуске сервера с `DATABASE_DSN`.

Вручную:
```bash
DSN="postgres://sentinel:sentinel_password@localhost:5432/sentinel_db?sslmode=disable"

migrate -database $DSN -path ./migrations up      # применить
migrate -database $DSN -path ./migrations down 1  # откатить одну
```

## Важно

- Не редактируйте примененные миграции - создавайте новые
- Всегда пишите `.down.sql` для отката
- Тестируйте откат перед prod
