# URL Shortener

REST API для сокращения URL, написанный на Go.

Сервис позволяет создавать короткие ссылки для длинных URL и перенаправлять пользователя по исходному адресу. Для хранения данных используется PostgreSQL.

![Работа сервиса](https://raw.githubusercontent.com/DenisNikolsky/url-shortener/63b3b0e/example.png)

## Возможности

- создание коротких ссылок;
- перенаправление по короткому коду;
- подсчёт количества переходов;
- валидация URL;
- обработка ошибок;
- PostgreSQL;
- database migrations;
- unit-тесты;
- integration-тесты;
- HTTP request logging;
- graceful shutdown;
- Swagger UI;
- Docker Compose;
- автоматическая проверка через GitHub Actions.

## Стек

- **Go**
- **Echo** — HTTP framework
- **PostgreSQL** — база данных
- **pgx** — PostgreSQL driver
- **golang-migrate** — database migrations
- **Swagger** — API documentation
- **Docker / Docker Compose**
- **GitHub Actions**

## Handler

Отвечает за HTTP-уровень:

- **получение HTTP-запроса**
- **обработку request body**
- **вызов service**
- **формирование HTTP-ответа**
- **redirect**
## Service

Содержит бизнес-логику приложения:

- **проверка URL**
- **генерация короткого кода**
- **создание URL**
- **получение URL**
- **увеличение количества переходов**
## Repository

Отвечает за взаимодействие с PostgreSQL:

- **создание записи**
- **получение URL по short code**
- **увеличение количества кликов**

## Запуск проекта

### 1. Требования

Перед запуском необходимо установить:

- Go 1.25+
- Docker
- Docker Compose

### 2. Настройка переменных окружения

Создай файл `.env` в корне проекта на основе `.env.example`.

Пример:

```env
POSTGRES_DB=url_shortener
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_PORT=5432

POSTGRES_TEST_DB=url_shortener_test
POSTGRES_TEST_USER=postgres
POSTGRES_TEST_PASSWORD=postgres
POSTGRES_TEST_PORT=5433

SERVER_PORT=8080
```

## Запуск через Docker

Запуск сервиса:

```docker compose up --build```

После запуска будут доступны:

```
API: http://localhost:8080
Swagger UI: http://localhost:8080/swagger/index.html
```

Остановить контейнеры:

```docker compose down```

## Проверка API

Создать короткую ссылку:

```
curl -X POST http://localhost:8080/urls \
-H "Content-Type: application/json" \
-d "{\"url\":\"https://google.com\"}"
```

Пример ответа:
```
{
"short_code": "abc12345",
"original_url": "https://google.com"
}
```

После этого можно открыть:
```
http://localhost:8080/abc12345
```

Сервис выполнит HTTP redirect на исходный URL.
