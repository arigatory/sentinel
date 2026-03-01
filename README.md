# Sentinel

Сервер сбора метрик и системы алертинга для мониторинга производительности приложений.

# Терминал 1 - запуск сервера

## Запуск сервера с подключением к БД
DATABASE_DSN="postgres://sentinel:sentinel_password@localhost:5432/sentinel_db?sslmode=disable" \
go run ./cmd/server/ -a "localhost:8080"

## Вариант 2: через флаг командной строки
go run ./cmd/server/ -a "localhost:8080" -d "postgres://sentinel:sentinel_password@localhost:5432/sentinel_db?sslmode=disable"

## Указать порт только
go run ./cmd/server/ -a "localhost:8080"

## Терминал 2 - запуск агента
go run ./cmd/agent/ -a "localhost:8080" -r 10 -p 2



## О проекте

Sentinel — это легковесная система мониторинга, которая позволяет:
- Собирать метрики производительности (CPU, память, custom метрики)
- Хранить исторические данные
- Настраивать правила алертинга
- Получать уведомления о критических событиях

Проект разработан в рамках трека «Сервер сбора метрик и алертинга» курса Яндекс Практикум.

## Начало работы

### Инициализация проекта

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере
2. В корне репозитория выполните команду для создания Go-модуля:
```bash
go mod init github.com/<your-username>/sentinel
```

### Обновление шаблона

Чтобы получать обновления автотестов и других частей шаблона, добавьте upstream-репозиторий:
```bash
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов:
```bash
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки по шаблону `iter<number>`, где `<number>` — порядковый номер инкремента.

**Примеры:**
- Ветка `iter4` — запустятся тесты для инкрементов 1-4
- Мёрж в `main` — запустятся все автотесты

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## База данных PostgreSQL

Проект поддерживает интеграцию с PostgreSQL для хранения метрик и проверки соединения.

### Запуск PostgreSQL через Docker Compose

В корне проекта находится файл `docker-compose.yml` для быстрого запуска PostgreSQL:

```bash
# Запуск PostgreSQL в фоновом режиме
docker-compose up -d

# Просмотр логов
docker-compose logs -f postgres

# Остановка
docker-compose down

# Остановка с удалением данных
docker-compose down -v
```

### Подключение к базе данных

Строка подключения к БД может быть передана через:

1. **Переменную окружения** `DATABASE_DSN`:
```bash
export DATABASE_DSN="postgres://sentinel:sentinel_password@localhost:5432/sentinel_db?sslmode=disable"
go run ./cmd/server/
```

2. **Флаг командной строки** `-d`:
```bash
go run ./cmd/server/ -d "postgres://sentinel:sentinel_password@localhost:5432/sentinel_db?sslmode=disable"
```

3. **Файл .env** (см. `.env.example`):
```bash
cp .env.example .env
# Отредактируйте .env под свои нужды
```

### Проверка соединения с БД

Endpoint `GET /ping` проверяет соединение с базой данных:

```bash
# Успешное подключение
curl http://localhost:8080/ping
# Ответ: 200 OK

# Ошибка подключения
# Ответ: 500 Internal Server Error
```

**Примечание:** Если `DATABASE_DSN` не указан, сервер запустится без подключения к БД, и endpoint `/ping` вернет 500 ошибку.

### Используемые технологии

Для работы с PostgreSQL используется библиотека **[pgx v5](https://github.com/jackc/pgx)** - современный высокопроизводительный драйвер PostgreSQL для Go с поддержкой:
- Connection pooling
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