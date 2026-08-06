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

## Ключевые архитектурные решения и заметки

### Почему интерфейсы репозиториев в `application`, а не в `domain`?
Осознанный компромисс: в духе идиоматического Go «интерфейсы рядом с потребителем». Use case’ы (`application`) объявляют нужные им контракты. Домен остаётся чистым, не зная о хранении. Реализация в `infrastructure` импортирует `application` и подстраивается под контракт. Это предотвращает циклические зависимости.

### Почему `common` в транспорте?
HTTP-хендлеры разделены по сущностям (`auth`, `wallet`), чтобы избежать мешанины. Общие функции (обработка ошибок, сериализация) вынесены в `common`, чтобы избежать циклических импортов. Константы типа имени сессионной куки живут там же, чтобы middleware мог их использовать, не импортируя конкретный хендлер.

### Почему `pkg/contextkeys` отдельно?
Ключи нужны middleware, логгеру и потенциально use case’ам. Вынос в отдельный легковесный пакет позволяет всем зависеть от него без лишних связей (например, middleware не зависит от логгера).

### Почему сессии вместо JWT?
Серверные сессии — мгновенная инвалидация (logout), безопасность (нет данных пользователя в токене), проще реализовать OTP в будущем. Кука `session_id` содержит только случайный идентификатор.

### Почему `Create` кошелька/пользователя без предварительной проверки?
Мы полагаемся на атомарную проверку уникальности через constraint базы данных. Последовательность `SELECT + INSERT` неатомарна и подвержена гонкам. Опора на `UNIQUE` constraint и обработка ошибки `unique_violation` даёт гарантию без лишних запросов и гонок.

### Где обновляется `UpdatedAt`?
- В доменных методах (`Wallet.Deposit`, `Wallet.Withdraw`) — потому что бизнес-операция фиксирует время изменения.
- В use case’е (`UserUsecase.Update`) — если операция не инкапсулирована в доменный метод, время устанавливается перед сохранением.
- Репозиторий **не** устанавливает `UpdatedAt` — он только сохраняет переданное состояние.

### Обработка ошибок и маппинг
- Домен возвращает чистые ошибки (`ErrInsufficientFunds`, `ErrUserNotFound`).
- Репозиторий преобразует GORM/DB-ошибки в доменные (`gorm.ErrRecordNotFound` → `domain.ErrUserNotFound`).
- HTTP-слой через `common.MapDomainError` преобразует доменные ошибки в `HttpError` со статусами и сообщениями для клиента.

### Валидация
- HTTP-запросы валидируются через `go-playground/validator` (теги в структурах запросов).
- Бизнес-правила защищены доменными методами и конструкторами (`NewUser` может возвращать ошибку при невалидных данных).

### Конкурентность и атомарность
На этапе 1 конкурентные операции не поддерживаются. На этапе 2 появится `SELECT ... FOR UPDATE` и/или оптимистичная блокировка через `version` для обеспечения согласованности при переводах и списаниях.

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