# Vault

Микросервис кошельков и переводов. Учебный проект для отработки конкурентных операций, DDD, GORM, MongoDB и нагрузочного тестирования.

## Архитектура

Проект следует принципам Domain-Driven Design (тактический дизайн) и чистой архитектуры.

- **`internal/domain`** — сущности, value objects, доменные ошибки. Не зависит от инфраструктуры.
- **`internal/application`** — сценарии использования (use cases). Оркеструет домен и репозитории.
- **`internal/infrastructure`** — реализации репозиториев (Postgres, Mongo), HTTP-транспорт, менеджеры блокировок.
- **`cmd/app`** — точка входа, композиция зависимостей.
- **`pkg`** — переиспользуемые утилиты (логгер, хеширование, ключи контекста), не привязанные к конкретному слою.

### Структура пакетов (актуальная)

```
cmd/app/main.go
internal/
  domain/
    user.go
    wallet.go           # (позже)
    transaction.go      # (позже)
    errors.go
  application/
    user/
      commands.go       # DTO для use case'ов
      repository.go     # интерфейс UserRepository
      usecase.go
    auth/
      commands.go
      contracts.go      # интерфейсы authService, sessionStore
      usecase.go
  infrastructure/
    persistence/
      postgres/
        models.go       # GORM-модели
        user_repo.go    # реализация UserRepository
        db.go           # подключение к БД, миграции
        errors.go       # mapGormError, DB-ошибки
    transport/
      http/
        common/
          errors.go     # HTTP-ошибки, маппинг домен → статус
          response.go   # WriteJSON, WriteError
          response_models.go # PaginatedResponse и пр.
          utils.go      # ReadSessionID, GetRequestID
        auth/
          handler.go    # Echo-хендлеры Auth
          contracts.go  # интерфейс authService (зависимость)
          request.go    # DTO запросов
          response.go   # DTO ответов
        server.go       # инициализация Echo, регистрация роутов
        middleware.go   # RequestID, auth (будущая)
pkg/
  contextkeys/
    contextkeys.go      # ключи для context.WithValue
  logger/
    logger.go           # logrus-обёртка с FromContext
  hash/
    argon2.go           # HashArgon2, CompareArgon2
```

### Почему интерфейсы репозиториев живут в application, а не в domain?

Это осознанный компромисс в духе идиоматического Go и чистой архитектуры. Use case'ы (`application`) используют репозитории, поэтому они объявляют нужные им контракты. Домен остаётся чистым, не зная о хранении. Реализация в `infrastructure` импортирует `application` и подстраивается под контракт. Это предотвращает циклические зависимости и сохраняет привычный Go-стиль «интерфейсы рядом с потребителем».

### Почему `common` в транспорте, а не всё в одном пакете?

HTTP-хендлеры разделены по сущностям (`auth`, скоро `user`, `wallet`), чтобы избежать мешанины. Общие функции (обработка ошибок, сериализация) вынесены в `common`, чтобы избежать циклических импортов между пакетами хендлеров. Константы вроде имени сессионной куки тоже живут в `common`, чтобы middleware мог их использовать, не импортируя конкретный хендлер.

### Почему контекстные ключи вынесены в `pkg/contextkeys`?

Ключи, используемые в `context.WithValue`, нужны и middleware (чтобы положить request ID), и логгеру (чтобы его прочитать), и потенциально use case'ам. Если оставить их в пакете логгера, middleware будет зависеть от логгера. Вынос в отдельный легковесный пакет позволяет всем слоям зависеть от него без лишних связей.

### Почему сессии вместо JWT?

Серверные сессии дают возможность мгновенной инвалидации (logout) и не раскрывают данные пользователя в токене. Для учебного проекта это упрощает реализацию OTP в будущем и позволяет попрактиковаться с хранилищем сессий (in-memory, позже Redis). Кука `session_id` содержит только случайный идентификатор; все данные сессии хранятся на сервере.

### Почему Echo v5?

Выбран для эксперимента с новым API `StartConfig.Start`, встроенным graceful shutdown на основе контекста. Раньше использовали v4 с ручным `Shutdown`. Это позволяет сравнить подходы и получить практический опыт с актуальной версией фреймворка.

## Стек

- **Go** 1.22+
- **PostgreSQL** (GORM)
- **MongoDB** (событийный аудит-лог) — появится позже
- **Echo v5** (HTTP-фреймворк)
- **golang-migrate** (миграции)
- **logrus** (логирование)
- **go-playground/validator** (валидация HTTP-запросов)
- **Argon2** (хэширование паролей)
- **Docker / docker-compose**

## Запуск

```bash
# 1. Запустить PostgreSQL и MongoDB (если уже настроен docker-compose)
docker-compose up -d

# 2. Применить миграции (пока можно запустить вручную или через cmd/migrate)
go run cmd/migrate/main.go up

# 3. Запустить сервер
go run cmd/app/main.go
```

Сервер стартует на порту `:8080`. Graceful shutdown срабатывает по `SIGINT`/`SIGTERM`, ожидание завершения текущих запросов — 10 секунд.