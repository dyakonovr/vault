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
cmd/
  app/
    main.go               # точка входа: инициализация БД, репозиториев, use case'ов, запуск HTTP-сервера
internal/
  domain/
    user.go               # сущность User, конструктор, инварианты
    wallet.go             # сущность Wallet, методы Deposit/Withdrawal, инварианты баланса
    errors.go             # общие доменные ошибки (ErrUserNotFound, ErrWalletNotFound и др.)
  application/
    auth/
      commands.go         # DTO: LoginCommand, RegisterCommand
      contracts.go        # интерфейсы authService, sessionStore
      usecase.go          # AuthUsecase: Login, Register, Me, Logout
    user/
      commands.go         # DTO: CreateUserCommand, UpdateUserCommand, ListUsersParams
      contracts.go        # интерфейс UserRepository
      usecase.go          # UserUsecase: Create, GetByID, Update, List, Delete
    wallet/
      commands.go         # DTO: CreateWalletCommand, UpdateWalletCommand
      contracts.go        # интерфейс WalletRepository
      errors.go           # ошибки уровня приложения (ErrWalletAccessDenied)
      usecase.go          # WalletUsecase: Create, Deposit, Withdrawal, GetByID
  infrastructure/
    persistence/
      postgres/
        models.go         # GORM-модели (UserModel, WalletModel, TransactionModel)
        user_repo.go      # реализация UserRepository
        wallet_repo.go    # реализация WalletRepository
        db.go             # подключение к БД, запуск golang-migrate
        errors.go         # mapGormError, базовые ошибки БД (ErrDBNoRows, ErrDBUniqueViolation...)
    session/
      memory_store.go     # in-memory реализация SessionStore
    transport/
      http/
        common/
          errors.go       # HttpError, MapDomainError → HTTP-статусы
          response.go     # HTTPSuccessResponse, HTTPErrorResponse
          response_models.go # PaginatedResponse, ErrorResponse
          utils.go        # ReadSessionID, GetRequestID, куки
        auth/
          handler.go      # Echo-хендлеры: /auth/login, /auth/register, /auth/me, /auth/logout
          contracts.go    # интерфейс authService (зависимость хендлера)
          request.go      # LoginRequest, RegisterRequest
          response.go     # UserResponse
        wallet/
          handler.go      # Echo-хендлеры: /api/wallets, /api/wallets/:id/balance, /deposit, /withdrawal
          contracts.go    # интерфейс walletService
          request.go      # CreateWalletRequest, UpdateWalletRequest
          response.go     # WalletResponse
        server.go         # инициализация Echo, группы роутов, gracefull shutdown
        middleware.go     # RequestIDMiddleware (request_id в контекст и заголовок)
pkg/
  ctxkeys/
    ctxkeys.go            # ключи контекста (RequestIDKey и др.)
  logger/
    logger.go             # logrus-обёртка, FromContext(ctx) → *Entry с request_id
  hash/
    argon2.go             # HashArgon2, CompareArgon2
  crypto/
    utils.go              # GenerateSessionID (crypto/rand)
migrations/
  000001_init.up.sql
  000001_init.down.sql
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
- В доменных методах (`Wallet.Deposit`, `Wallet.Withdrawal`) — потому что бизнес-операция фиксирует время изменения.
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

### Транзакционность и Unit of Work

**Проблема:** Бизнес-операции часто требуют атомарного изменения нескольких агрегатов и/или записи в журнал транзакций. Например, зачисление средств на кошелёк включает обновление баланса (`Wallet.Deposit`) и создание записи `Transaction`. Эти две операции должны выполниться как единое целое — либо обе, либо ни одна. Возникает вопрос: как обеспечить транзакционность, не допуская утечки инфраструктурных деталей (GORM, `*sql.Tx`) в домен или use case'ы?

**Решение — паттерн Unit of Work (UoW):**  
Use case'ы, требующие транзакционности, объявляют собственные контракты на доступ к репозиториям внутри транзакции. Эти контракты живут в application-слое и не зависят от конкретной БД.

- Общий контракт `TransactionalResources` и `UnitOfWork` для операций с кошельком и транзакциями размещён в `application/wallet`. Он предоставляет `WalletRepository` и `TransactionRepository`.
- Use case'ы `DepositUseCase` и `WithdrawUseCase` (и будущий `TransferUseCase`) используют этот контракт, не дублируя его.
- Инфраструктурный слой (`postgres`) создаёт тонкий адаптер `WalletRepositories`, реализующий `wallet.TransactionalResources`, и метод `BeginWalletTransaction`, который оборачивает вызов в транзакцию GORM.

**Почему не единый UoW на всё приложение?**  
Создание одного универсального `UnitOfWork`, возвращающего все возможные репозитории, привело бы к god object'у и нарушению Interface Segregation Principle. Вместо этого:

- Для семейства операций «изменение баланса + запись транзакции» используется один контракт (`wallet.TransactionalResources`).
- Для принципиально другого сценария (например, регистрация пользователя с автоматическим созданием кошелька) будет создан отдельный контракт (`registration.TransactionalResources` с `UserRepo` и `WalletRepo`), не влияющий на существующие.

**Как это работает (на примере `DepositUseCase`):**
- Контракт определён в `application/wallet/contracts.go` (`TransactionalResources`, `UnitOfWork`).
- Use case принимает `wallet.UnitOfWork` и вызывает `Begin(ctx, func(resources wallet.TransactionalResources) error { ... })`.
- В коллбэке use case проверяет идемпотентность, загружает кошелёк, вызывает `wallet.Deposit`, сохраняет кошелёк и создаёт транзакцию — все операции внутри одной транзакции PostgreSQL.
- `postgres/wallet_uow.go` реализует `wallet.TransactionalResources` и `wallet.UnitOfWork`, связывая GORM-транзакцию с адаптерами.

Таким образом, use case полностью контролирует логику, оставаясь независимым от БД, а инфраструктура предоставляет атомарность. Это соответствует принципам чистой архитектуры и DDD: домен не знает о транзакциях, application управляет сценариями, infrastructure реализует хранение.

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