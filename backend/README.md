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

### Атомарные переводы и конкурентность

**Проблема:** Перевод средств между кошельками — это операция, затрагивающая как минимум два агрегата (`Wallet` отправителя и `Wallet` получателя) и требующая создания двух записей `Transaction` (`TRANSFER_OUT` и `TRANSFER_IN`). Всё должно быть выполнено атомарно, чтобы избежать частичного списания или «пропажи» денег. Кроме того, при параллельных операциях с одним кошельком возможны гонки (lost update), когда два запроса читают один и тот же баланс и перезаписывают изменения друг друга.

**Решение — транзакции БД + явные блокировки строк (`SELECT ... FOR UPDATE`):**
- Все операции изменения баланса (`Deposit`, `Withdraw`, `Transfer`) выполняются внутри одной транзакции PostgreSQL через наш Unit of Work (`wallet.UnitOfWork`).
- Для предотвращения гонок при загрузке кошелька используется `SELECT ... FOR UPDATE`, который эксклюзивно блокирует строку до конца транзакции. Это гарантирует, что никто другой не изменит баланс параллельно.
- Чтобы избежать взаимных блокировок (deadlock) при переводе, кошельки блокируются в **детерминированном порядке** (по возрастанию их ID).

**Проверка идемпотентности при переводе:**
- Клиент передаёт `Idempotency-Key` в запросе.
- Внутри транзакции сначала проверяется существование транзакции с таким ключом (достаточно проверить `TRANSFER_OUT`). Если она найдена — операция уже была выполнена, возвращается успех без повторения.
- Составной уникальный индекс `UNIQUE(idempotency_key, type)` в таблице `transactions` позволяет обеим записям перевода иметь одинаковый ключ, но разные типы, и защищает от дублирования при повторных запросах.

**Что происходит при сбое:**
- Если в процессе перевода происходит ошибка (например, недостаточно средств) или сетевая проблема, транзакция БД автоматически откатывается. Никаких «половинчатых» изменений не остаётся. Это гарантируется свойствами ACID PostgreSQL.

**Ключевые моменты реализации:**
- `GetByIdForUpdate` добавлен в интерфейс `wallet.WalletRepository` и используется во всех сценариях изменения баланса.
- `TransferUseCase` координирует проверку идемпотентности, блокировку, вызов доменных методов (`Withdraw`, `Deposit`) и сохранение транзакций через `wallet.TransactionalResources`.
- Уровень изоляции транзакций оставлен по умолчанию (`READ COMMITTED`), так как он корректно работает с явными блокировками `FOR UPDATE`.

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

## Тесты

### Запуск

```bash
# Все тесты (кроме интеграционных — они требуют Docker)
go test ./internal/domain/... ./internal/application/... ./internal/infrastructure/transport/http/...

# Только интеграционные (требуют Docker, поднимают PostgreSQL в контейнере)
go test ./internal/infrastructure/persistence/postgres/...

# Все сразу
go test ./internal/... -timeout 180s
```

### Domain (20 тестов)

Unit-тесты доменных сущностей. Без моков — чистая бизнес-логика.

| Файл | Что проверяется |
|------|-----------------|
| `domain/wallet_test.go` | `Deposit`, `Withdraw` — нормальные сценарии, недостаточно средств, невалидная сумма, переполнение баланса |
| `domain/user_test.go` | Конструктор `NewUser` — валидные/невалидные логин и пароль |
| `domain/transaction_test.go` | Конструктор `NewTransaction` — все типы (deposit, withdrawal, transfer_out, transfer_in), невалидный тип |

### Application (39 тестов)

Unit-тесты use case'ов с моками. Каждый мок-объект автоматически проверяет ожидания в `t.Cleanup`.

| Пакет | Файл | Тесты | Что проверяется |
|-------|------|-------|-----------------|
| `auth` | `usecase_test.go` | 7 | `Login` (успех, пользователь не найден, неверный пароль, дублирование при регистрации), `Register` (успех, уже существует), `Me` (успех, сессия не найдена), `Logout` |
| `deposit` | `usecase_test.go` | 4 | `Do` (успех, идемпотентность, ошибка владения, кошелёк не найден) |
| `withdrawal` | `usecase_test.go` | 4 | `Do` (успех, идемпотентность, недостаточно средств, кошелёк не найден) |
| `transfer` | `usecase_test.go` | 6 | `Do` (успех, идемпотентность, ошибка владения, недостаточно средств, кошелёк не найден, порядок блокировки) |
| `wallet` | `usecase_test.go` | 5 | `Create`, `GetById` (успех, кошелёк не найден, доступ запрещён) |
| `user` | `usecase_test.go` | 5 | `Create`, `GetById`, `Update`, `Delete`, `List` |
| `transaction` | `usecase_test.go` | 4 | `List` (успех, пустой список, с фильтрами), `GetById` (успех, не найден) |

### Infrastructure: HTTP-хендлеры (28 тестов)

Unit-тесты хендлеров через Echo-роутер. Моки подменяют сервисы, middleware подменяют контекст (userID, idempotency key).

| Пакет | Файл | Тесты | Что проверяется |
|-------|------|-------|-----------------|
| `auth` | `handler_test.go` | 8 | `Login` (успех, неверные credentials, валидация), `Register` (успех, дубликат), `Me` (успех, нет сессии), `Logout` |
| `wallet` | `handler_test.go` | 12 | `Create` (успех, уже существует), `GetBalanceByID` (успех, не найден, доступ запрещён, нет авторизации), `Deposit` (успех, невалидная сумма), `Withdrawal` (успех, недостаточно средств), `Transfer` (успех, валидация) |
| `transaction` | `handler_test.go` | 8 | `List` (успех, пусто, с фильтрами, пагинация, нет авторизации), `GetByID` (успех, не найден, невалидный ID) |

### Infrastructure: PostgreSQL (18 тестов)

Интеграционные тесты реальных репозиториев. Используют testcontainers — поднимают PostgreSQL в Docker-контейнере, выполняют миграции, работают с реальной БД.

| Файл | Тесты | Что проверяется |
|------|-------|-----------------|
| `user_repo_test.go` | 7 | `Create`, `GetById`, `GetByLogin`, `GetById` (не найден), дубликат логина, `Update`, `Delete`, `List` |
| `wallet_repo_test.go` | 5 | `Create`, `GetById`, `GetByUserId`, `Update` баланса, дубликат `user_id` |
| `transaction_repo_test.go` | 4 | `Create`, `GetByIdempotencyKey`, дубликат idempotency key, `List` по `wallet_id` |
| `concurrent_test.go` | 1 | Параллельное пополнение одного кошелька 10 горутинами — проверяет `SELECT ... FOR UPDATE` |