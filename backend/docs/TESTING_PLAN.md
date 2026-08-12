# План тестирования Vault Backend

## Обзор

Реализация полного покрытия тестами: domain-юнит-тесты, application-юнит-тесты (с моками), infrastructure-интеграционные тесты (PostgreSQL через testcontainers), HTTP-тесты хендлеров.

## Стек тестирования

- `testify` — assertion + require + suite
- `mockery` — генерация моков из интерфейсов (`//go:generate`)
- `testcontainers-go` — контейнеры PostgreSQL для интеграционных тестов
- `httptest` — тестирование HTTP-хендлеров

## Этап 0: Подготовка

### 0.1 Экспорт интерфейсов

Интерфейсы в `application`-слое неэкспортированные. Нужно сделать их публичными для mockery:

| Файл | Интерфейс | Было → Стало |
|---|---|---|
| `application/wallet/contracts.go` | `walletRepository` → `WalletReader` |
| `application/deposit/contracts.go` | `walletOwnershipChecker` → `WalletOwnershipChecker` |
| `application/withdrawal/contracts.go` | `walletOwnershipChecker` → `WalletOwnershipChecker` |
| `application/transfer/contracts.go` | `walletOwnershipChecker` → `WalletOwnershipChecker` |
| `application/auth/contracts.go` | `userService` → `UserService` |
| `application/auth/contracts.go` | `sessionStore` → `SessionStore` |
| `application/user/contracts.go` | `userRepository` → `UserRepository` |
| `application/transaction/contracts.go` | `transactionRepository` → `TransactionRepository` |
| `application/transaction/contracts.go` | `walletOwnershipChecker` → `WalletOwnershipChecker` |

Директивы `//go:generate` добавляются перед каждым интерфейсом.

### 0.2 Конфигурация mockery

Создать `.mockery.yaml` в корне проекта.

### 0.3 Зависимости

```
go get github.com/stretchr/testify
go get github.com/testcontainers/testcontainers-go
go install github.com/vektra/mockery/v2@latest
```

### 0.4 Генерация моков

```
go generate ./internal/application/...
```

---

## Этап 1: Domain-юнит-тесты

Чистые тесты без моков. Проверяют бизнес-правила.

### `internal/domain/wallet_test.go`

| Тест | Что проверяем |
|---|---|
| `TestNewWallet` | userID > 0 → создание с балансом 0 |
| `TestNewWallet_InvalidUserID` | userID ≤ 0 → `ErrWalletIncorrectUserID` |
| `TestDeposit_Success` | баланс увеличивается на сумму |
| `TestDeposit_ZeroAmount` | amount = 0 → `ErrWalletInvalidAmount` |
| `TestDeposit_NegativeAmount` | amount < 0 → `ErrWalletInvalidAmount` |
| `TestWithdraw_Success` | баланс уменьшается, funds足够的 |
| `TestWithdraw_ZeroAmount` | amount = 0 → `ErrWalletInvalidAmount` |
| `TestWithdraw_InsufficientFunds` | amount > balance → `ErrInsufficientFunds` |
| `TestWithdraw_EntireBalance` | списание всей суммы → balance = 0 |

### `internal/domain/user_test.go`

| Тест | Что проверяем |
|---|---|
| `TestNewUser` | валидный логин + пароль → создание |
| `TestNewUser_EmptyLogin` | пустой логин → `ErrUserEmptyLogin` |
| `TestNewUser_EmptyPassword` | пустой пароль → `ErrUserEmptyPasswordHash` |

### `internal/domain/transaction_test.go`

| Тест | Что проверяем |
|---|---|
| `TestNewTransaction_Deposit` | тип deposit → статус completed |
| `TestNewTransaction_TransferOut` | тип transfer_out |
| `TestNewTransaction_TransferIn` | тип transfer_in |
| `TestNewTransaction_Withdrawal` | тип withdrawal |
| `TestNewTransaction_InvalidType` | недопустимый тип → `ErrTransactionIncorrectType` |
| `TestNewTransaction_ZeroAmount` | amount = 0 → `ErrTransactionInvalidAmount` |
| `TestNewTransaction_NegativeAmount` | amount < 0 → `ErrTransactionInvalidAmount` |
| `TestNewTransaction_InvalidWalletID` | walletID ≤ 0 → `ErrTransactionIncorrectWalletID` |

---

## Этап 2: Application-юнит-тесты

Моки генерируются mockery. Use case тестируется изолированно от БД.

### `internal/application/wallet/usecase_test.go`

| Тест | Моки | Проверяем |
|---|---|---|
| `TestGetById_Owner` | WalletReader.GetById → валидный кошелёк | Возвращает кошелёк |
| `TestGetById_NotOwner` | WalletReader.GetById → кошелёк другого юзера | `ErrWalletAccessDenied` |
| `TestGetByUserId` | WalletReader.GetByUserId | Возвращает кошелёк |
| `TestCreate_Success` | WalletReader.Create → nil | Возвращает созданный кошелёк |
| `TestCreate_InvalidUserID` | — | `ErrWalletIncorrectUserID` |

### `internal/application/deposit/usecase_test.go`

| Тест | Моки | Проверяем |
|---|---|---|
| `TestDo_Success` | OwnershipChecker → nil, UoW → success | Транзакция создана |
| `TestDo_Idempotent` | UoW → существующая транзакция | Возвращает существующую, без изменений |
| `TestDo_OwnershipError` | OwnershipChecker → error | Ошибка владения |
| `TestDo_WalletNotFound` | UoW → `ErrWalletNotFound` | Ошибка пробрасывается |

### `internal/application/withdrawal/usecase_test.go`

| Тест | Моки | Проверяем |
|---|---|---|
| `TestDo_Success` | OwnershipChecker → nil, UoW → success | Транзакция создана |
| `TestDo_Idempotent` | UoW → существующая транзакция | Возвращает существующую |
| `TestDo_InsufficientFunds` | UoW → `ErrInsufficientFunds` | Ошибка пробрасывается |
| `TestDo_OwnershipError` | OwnershipChecker → error | Ошибка владения |

### `internal/application/transfer/usecase_test.go`

| Тест | Моки | Проверяем |
|---|---|---|
| `TestDo_Success` | OwnershipChecker → nil, UoW → success | Две транзакции (OUT + IN) |
| `TestDo_Idempotent` | UoW → существующая транзакция | Возвращает существующую |
| `TestDo_InsufficientFunds` | UoW → `ErrInsufficientFunds` | Ошибка пробрасывается |
| `TestDo_OwnershipError` | OwnershipChecker → error | Ошибка владения |
| `TestDo_WalletLockingOrder` | Проверка порядка блокировки | FromID < ToID → correct order |

### `internal/application/user/usecase_test.go`

| Тест | Моки | Проверяем |
|---|---|---|
| `TestCreate_Success` | UserRepository.Create → nil | Хэширование пароля, создание |
| `TestGetById_Success` | UserRepository.GetById → user | Возвращает пользователя |
| `TestGetById_NotFound` | UserRepository.GetById → error | Ошибка пробрасывается |
| `TestUpdate_Success` | UserRepository → user + hash, Update → nil | Пароль обновлён |
| `TestUpdate_WrongPassword` | Старый пароль не совпадает | `ErrWrongPassword` |

### `internal/application/auth/usecase_test.go`

| Тест | Моки | Проверяем |
|---|---|---|
| `TestLogin_Success` | UserService.GetByLogin, SessionStore.Create | Сессия возвращена |
| `TestLogin_InvalidCredentials` | UserService → `ErrInvalidLoginOrPassword` | Ошибка аутентификации |
| `TestRegister_Success` | UserService.GetByLogin → not found, Create → nil | Регистрация |
| `TestRegister_AlreadyExists` | UserService.GetByLogin → user | `ErrUserAlreadyExists` |
| `TestMe_Success` | SessionStore.GetUserID, UserService.GetById | Возвращает пользователя |
| `TestMe_InvalidSession` | SessionStore → error | `ErrSessionNotFound` |
| `TestLogout_Success` | SessionStore.Delete → nil | Успешный выход |

---

## Этап 3: Infrastructure — интеграционные тесты (PostgreSQL)

### `internal/infrastructure/persistence/postgres/testutil_test.go`

Хелпер `SetupTestDB(t)`:
- Запуск PostgreSQL через testcontainers
- Применение мigrations
- Возврат `*gorm.DB` + функции cleanup

### `internal/infrastructure/persistence/postgres/user_repo_test.go`

| Тест | Проверяем |
|---|---|
| `TestCreate` | Создание, присвоение ID |
| `TestFindByID` | Поиск по ID, сравнение полей |
| `TestFindByLogin` | Поиск по login |
| `TestFindByID_NotFound` | → `domain.ErrUserNotFound` |
| `TestCreate_DuplicateLogin` | → `domain.ErrUserAlreadyExists` |
| `TestUpdate` | Обновление пароля |
| `TestDelete` | Удаление → NotFound |

### `internal/infrastructure/persistence/postgres/wallet_repo_test.go`

| Тест | Проверяем |
|---|---|
| `TestCreateWallet` | Создание, присвоение ID |
| `TestFindByID` | Поиск по ID |
| `TestFindByUserID` | Поиск по user_id |
| `TestUpdateBalance` | Обновление баланса |
| `TestCreateWallet_DuplicateUserID` | → `domain.ErrWalletAlreadyExists` |

### `internal/infrastructure/persistence/postgres/transaction_repo_test.go`

| Тест | Проверяем |
|---|---|
| `TestCreateTransaction` | Создание, присвоение ID |
| `TestFindByIdempotencyKey` | Поиск по ключу |
| `TestCreateTransaction_DuplicateKey` | → `domain.ErrTransactionAlreadyExists` |
| `TestListByWalletID` | Фильтрация по кошельку |

### `internal/infrastructure/persistence/postgres/concurrent_test.go`

| Тест | Проверяем |
|---|---|
| `TestConcurrentDeposits` | 2 параллельных депозита с FOR UPDATE — корректный баланс |

---

## Этап 4: HTTP-транспорт — тесты хендлеров

### `internal/infrastructure/transport/http/auth/handler_test.go`

| Тест | Маршрут | Ожидаем |
|---|---|---|
| `TestLogin_Success` | POST /api/auth/login | 204 + cookie |
| `TestLogin_InvalidCredentials` | POST /api/auth/login | 401 |
| `TestLogin_ValidationError` | POST /api/auth/login | 422 |
| `TestRegister_Success` | POST /api/auth/register | 204 |
| `TestRegister_AlreadyExists` | POST /api/auth/register | 409 |
| `TestRegister_ValidationError` | POST /api/auth/register | 422 |
| `TestMe_Success` | GET /api/auth/me | 200 + user |
| `TestMe_NoSession` | GET /api/auth/me | 401 |
| `TestLogout_Success` | POST /api/auth/logout | 204 |
| `TestLogout_NoSession` | POST /api/auth/logout | 401 |

### `internal/infrastructure/transport/http/wallet/handler_test.go`

| Тест | Маршрут | Ожидаем |
|---|---|---|
| `TestCreate_Success` | POST /api/wallets | 201 |
| `TestCreate_AlreadyExists` | POST /api/wallets | 409 |
| `TestGetBalance_Success` | GET /api/wallets/:id/balance | 200 |
| `TestGetBalance_NotFound` | GET /api/wallets/:id/balance | 404 |
| `TestGetBalance_Unauthorized` | GET /api/wallets/:id/balance | 401 |
| `TestDeposit_Success` | POST /api/wallets/:id/deposit | 200 |
| `TestDeposit_InsufficientFunds` | POST /api/wallets/:id/deposit | 409 |
| `TestWithdraw_Success` | POST /api/wallets/:id/withdrawal | 200 |
| `TestWithdraw_InsufficientFunds` | POST /api/wallets/:id/withdrawal | 409 |
| `TestTransfer_Success` | POST /api/wallets/transfers | 200 |
| `TestTransfer_InsufficientFunds` | POST /api/wallets/transfers | 409 |

---

## Структура файлов (итог)

```
.mockery.yaml
TESTING_PLAN.md
internal/
  domain/
    wallet_test.go
    user_test.go
    transaction_test.go
  application/
    wallet/
      mocks/
        mock_WalletReader.go
        mock_WalletRepository.go
        mock_TransactionRepository.go
        mock_UnitOfWork.go
      usecase_test.go
    deposit/
      mocks/
        mock_WalletOwnershipChecker.go
      usecase_test.go
    withdrawal/
      mocks/
        mock_WalletOwnershipChecker.go
      usecase_test.go
    transfer/
      mocks/
        mock_WalletOwnershipChecker.go
      usecase_test.go
    user/
      mocks/
        mock_UserRepository.go
      usecase_test.go
    auth/
      mocks/
        mock_UserService.go
        mock_SessionStore.go
      usecase_test.go
  infrastructure/
    persistence/postgres/
      testutil_test.go
      user_repo_test.go
      wallet_repo_test.go
      transaction_repo_test.go
      concurrent_test.go
    transport/http/
      auth/handler_test.go
      wallet/handler_test.go
```

---

## Порядок выполнения

1. [ ] Этап 0.1 — Экспорт интерфейсов + `//go:generate`
2. [ ] Этап 0.2 — `.mockery.yaml`
3. [ ] Этап 0.3 — `go get` зависимостей
4. [ ] Этап 0.4 — `go generate` + проверка моков
5. [ ] Этап 1 — Domain-тесты
6. [ ] Этап 2 — Application-тесты
7. [ ] Этап 3 — Infrastructure-интеграционные тесты
8. [ ] Этап 4 — HTTP-тесты
9. [ ] `go test ./... -v -count=1` — финальная проверка
