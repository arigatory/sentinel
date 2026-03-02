# Sentinel

Сервер сбора метрик для мониторинга производительности приложений.

## Быстрый старт

```bash
# Запустить PostgreSQL
docker compose up -d

# Запустить сервер
DATABASE_DSN="postgres://sentinel:sentinel_password@localhost:5432/sentinel_db?sslmode=disable" \
go run ./cmd/server/ -a "localhost:8080"

# В другом терминале - запустить агент
go run ./cmd/agent/ -a "localhost:8080" -r 10 -p 2
```



## Возможности

- Сбор метрик производительности (CPU, память, custom метрики)
- Хранение в PostgreSQL или в памяти
- REST API для получения и обновления метрик
- Автоматические миграции БД

*Проект Яндекс Практикума*

## PostgreSQL

### Запуск

```bash
docker compose up -d
```

### Подключение

Через переменную окружения:
```bash
export DATABASE_DSN="postgres://sentinel:sentinel_password@localhost:5432/sentinel_db?sslmode=disable"
```

Или флаг `-d`:
```bash
go run ./cmd/server/ -d "postgres://..."
```

### Миграции

При запуске сервер автоматически применит миграции из `migrations/`. Используется [golang-migrate](https://github.com/golang-migrate/migrate).

Создать новую миграцию:
```bash
migrate create -ext sql -dir ./migrations -seq your_feature_name
```

Подробнее см. `migrations/README.md`

### Хранение метрик

Приоритет выбора:
1. PostgreSQL (если указан `DATABASE_DSN`)
2. Файл (если указан `FILE_STORAGE_PATH`)
3. Память (по умолчанию)

### Проверка

```bash
curl http://localhost:8080/ping  # статус БД
curl http://localhost:8080/      # все метрики
```

## Технологии

- Go 1.24
- PostgreSQL 16
- [pgx v5](https://github.com/jackc/pgx) - драйвер PostgreSQL
- [golang-migrate](https://github.com/golang-migrate/migrate) - миграции
- [chi](https://github.com/go-chi/chi) - HTTP router
- Context-aware операций
- Prepared statements
- Copy protocol
- Listen/Notify

## Архитектура проекта

Структура проекта является рекомендуемой, но не обязательной. Вы можете использовать любые архитектурные подходы:

- **Clean Architecture**
- **Domain-Driven Design (DDD)**
- **Hexagonal Architecture**
- **Layered Architecture**

Выбирайте подход, который лучше всего подходит для решения конкретных задач проекта.

## Технологический стек

- Go 1.21+
- HTTP/REST API
- PostgreSQL (опционально)
- Docker (опционально)

---

Разработано в рамках курса [Go-разработчик](https://practicum.yandex.ru/go/) от Яндекс Практикум