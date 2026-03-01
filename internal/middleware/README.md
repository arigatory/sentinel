# Middleware

Пакет содержит HTTP middleware для сервера метрик.

## Logger Middleware

Middleware для структурированного логирования HTTP запросов и ответов с использованием `slog`.

### Функциональность

- **URI запроса** - полный путь запроса
- **HTTP метод** - GET, POST, и т.д.
- **Время выполнения** - длительность обработки запроса
- **Код статуса** - HTTP статус код ответа
- **Размер ответа** - количество байт в теле ответа

### Использование

```go
import (
    "log/slog"
    "os"
    customMiddleware "github.com/arigatory/sentinel/internal/middleware"
)

// Создаём логгер
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Подключаем middleware
r := chi.NewRouter()
r.Use(customMiddleware.Logger(logger))
```

### Пример вывода

```json
{
  "time": "2026-02-13T16:45:56.484461+03:00",
  "level": "INFO",
  "msg": "HTTP request",
  "method": "POST",
  "uri": "/update/counter/myCounter/5",
  "status": 200,
  "size": 0,
  "duration": 52666
}
```

### Реализация

Middleware использует паттерн "обёртка" (wrapper) для перехвата:
1. **responseWriter** - обёртка над `http.ResponseWriter`
2. Перехватывает вызовы `WriteHeader()` для получения кода статуса
3. Перехватывает вызовы `Write()` для подсчёта размера ответа
4. Измеряет время выполнения с помощью `time.Now()` и `time.Since()`
