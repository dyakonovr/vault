# Go Testing & Mockery: Patterns Explained

## Содержание
1. [Основы: .On, .Return, .Maybe](#1-основы)
2. [Сопоставление аргументов](#2-сопоставление)
3. [Почему context.Background()](#3-context)
4. [Нейминг тестов](#4-нейминг)
5. [Внешний пакет _test](#5-внешний-пакет)
6. [Сложные конструкции](#6-сложные)
7. [Полезные советы](#7-советы)

---

## 1. Основы: .On, .Return, .Maybe

### .On — "Когда вызывается этот метод с этими аргументами..."

```go
userService.On("GetByLogin", context.Background(), "alice").Return(u, nil)
```

**Что происходит:** мы говорим mock-объекту: "Если кто-то вызовет метод `GetByLogin` с аргументами `(context.Background(), "alice")`, верни `(u, nil)`."

Это как настройка робота: "Если тебя спросят время — скажи 'полночь'."

### .Return — "Возвращай именно это"

```go
// Когда вызовут GetById с ID=1 → верни пользователя с ошибкой nil
sessionStore.On("GetByID", context.Background(), int64(1)).Return("session-abc", nil)

// Когда вызовут GetById с ID=999 → верни ошибку
sessionStore.On("GetByID", context.Background(), int64(999)).Return("", domain.ErrSessionNotFound)
```

**Ключевой момент:** `.Return` задаёт фиксированный ответ. Mock всегда вернёт то, что ты указал, даже если это не имеет смысла.

### .Maybe — "Может быть, этот метод вызовут, а может и нет"

```go
txRepo.On("Create", context.Background(), mock.AnythingOfType("*domain.Transaction")).
    Return(nil).Maybe()
```

**Когда нужно:** когда метод *может* быть вызван, но не обязан. Без `.Maybe()` mock выдаст ошибку в конце теста: "ожидался вызов, но не был".

Без `.Maybe()` mock ведёт себя так: "Я *обязан* быть вызван". С `.Maybe()`: "Я *могу* быть вызван, но это не критично."

### .Times(n) — "Вызови ровно N раз"

```go
sessionStore.On("Delete", context.Background(), "session-abc").Return(nil).Times(1)
```

Стандартно mock ожидает вызов ровно 1 раз. `.Times(3)` скажет: "Этот метод должен быть вызван 3 раза."

---

## 2. Сопоставление аргументов: .MatchedBy, .AnythingOfType

### mock.MatchedBy(func) — "Сопоставь с помощью функции"

Самая мощная конструкция. Вместо точного значения передаёшь функцию-предикат:

```go
walletRepo.On("Update", context.Background(), mock.MatchedBy(func(w *domain.Wallet) bool {
    return w.ID == 10 && w.UserID == 1 && w.Balance == 100
})).Return(nil)
```

**Зачем:** часто ты не знаешь точное значение аргумента (оно вычисляется в тестируемом коде), но знаешь *условие*.

Здесь мы не знаем точное время обновления (`UpdatedAt`), но знаем, что баланс должен стать 100. Функция проверяет только то, что нам важно.

**Как это работает внутри:**
1. Тестируемый код вызывает `walletRepo.Update(ctx, wallet)`
2. Mock проверяет каждый зарегистрированный `.On` вызов
3. Для `.MatchedBy` вызывается твоя функция с реальным аргументом
4. Если функция вернула `true` — это совпадение, возвращается `.Return(...)`

### mock.AnythingOfType("*domain.Transaction") — "Любой объект этого типа"

```go
txRepo.On("Create", context.Background(), mock.AnythingOfType("*domain.Transaction")).Return(nil)
```

**Когда нужно:** когда ты знаешь тип аргумента, но не можешь проверить его точное значение (например, UUID генерируется внутри).

Это короче чем `mock.MatchedBy(func(tx *domain.Transaction) bool { return true })`.

### mock.Anything — "Любой аргумент"

```go
mockService.On("Login", mock.Anything, mock.Anything).Return(session, nil)
```

Буквально: "Меня не волнует первый (или второй, третий...) аргумент — я отвечу на любой вызов."

**Осторожно:** слишком много `mock.Anything` делает тест бессмысленным — он будет проходить даже если тестируемый код передаёт неправильные аргументы.

### Порядок совпадения

Mock ищет совпадение **сверху вниз**. Если подходящих несколько — используется первое:

```go
// Сначала проверяется специальный случай:
walletRepo.On("GetByIdForUpdate", ctx, int64(10)).Return(w1, nil)

// Потом общий случай (если первый не совпал):
walletRepo.On("GetByIdForUpdate", ctx, mock.Anything).Return(domain.Wallet{}, domain.ErrNotFound)
```

---

## 3. Почему context.Background(), а не t.Context()?

### t.Context()

`t.Context()` появился в **Go 1.24** (февраль 2025). Он возвращает контекст, который автоматически отменяется при завершении теста:

```go
ctx := t.Context()  // отменится когда тест завершится
```

### context.Background()

`context.Background()` — это "чистый" контекст без дедлайнов и отмены:

```go
ctx := context.Background()  // живёт вечно
```

### Почему в тестах используется context.Background()

1. **Изоляция:** каждый вызов mock-метода с `context.Background()` — это точное совпадение. Если использовать `t.Context()`, то `context.Background() != t.Context()` и mock не сработает.

2. **Предсказуемость:** mock настраивается на конкретный контекст. `context.Background()` всегда один и тот же (нулевой).

3. **Генерация моков:** mockery генерирует моки с `mock.Anything` или конкретными значениями. В production коде контекст приходит из HTTP-запроса, а в тестах его нет.

### Когда использовать t.Context()

Когда тестируемый код **создаёт** контекст внутри:
```go
func TestSomething(t *testing.T) {
    ctx := context.Background()  // ← тут мы создаём, а не получаем
    result := myFunc(ctx, ...)
}
```

Но когда мы **настраиваем mock**, нужно точно знать, какой контекст передаст production-код:
```go
mock.On("GetUser", context.Background(), int64(1)).Return(user, nil)
```

---

## 4. Нейминг тестов: TestDo_Success vs TestDeposit_Create_Success

### Конвенция Go: `TestИмяФункции_Сценарий`

```go
func TestDo_Success(t *testing.T)        // функция Do, сценарий успеха
func TestDo_Idempotent(t *testing.T)     // функция Do, идемпотентность
func TestDo_OwnershipError(t *testing.T) // функция Do, ошибка владения
```

### Почему НЕ `TestDeposit_Create_Success`?

Потому что Go-тесты называются по имени тестируемой функции, а не по имени сущности.

Это как в таблицах умножения: `TestMultiplication_TwoTimesThree` или `TestMultiply_TwoTimesThree`.

### Структура названия

```
Test<ИмяФункции>_<Сценарий>
```

- `TestDo` — функция, которую тестируем
- `Success` — что происходит в этом тесте (успех, ошибка, boundary-case)

### Примеры из кодовой базы

```go
// deposit/usecase_test.go
func TestDo_Success(t *testing.T)           // пополнение успешно
func TestDo_Idempotent(t *testing.T)        // повторный вызов
func TestDo_OwnershipError(t *testing.T)    // кошелёк не принадлежит
func TestDo_WalletNotFound(t *testing.T)    // кошелёк не найден

// auth/usecase_test.go
func TestLogin_Success(t *testing.T)        // вход успешен
func TestLogin_UserNotFound(t *testing.T)   // пользователь не найден
func TestLogin_WrongPassword(t *testing.T)  // неверный пароль

// wallet/handler_test.go
func TestCreate_Success(t *testing.T)
func TestCreate_AlreadyExists(t *testing.T)
func TestGetBalanceByID_Success(t *testing.T)
func TestGetBalanceByID_WalletNotFound(t *testing.T)
```

### HTTP-хендлеры используют имя хендлера

```go
func TestLogin_Success(t *testing.T)     // AuthHandler.Login
func TestDeposit_Success(t *testing.T)   // WalletHandler.Deposit
```

Это работает потому что хендлеры вызываются через HTTP, а имя метода совпадает.

---

## 5. Внешний пакет _test: Почему package deposit_test, а не package deposit?

```go
// deposit/usecase_test.go
package deposit_test   // ← внешний пакет
```

### Внутренний пакет (package deposit)

```go
package deposit  // может видеть приватные поля и методы
```

### Внешний пакет (package deposit_test)

```go
package deposit_test  // видит только экспортированное
```

### Зачем?

1. **Изоляция:** тесты используют только публичный API (как реальные пользователи)
2. **Нет import cycles:** если пакет A зависит от B, B не может импортировать A
3. **Проверка экспортируемости:** если функция не экспортируема — тест не скомпилируется

### Когда использовать внутренний пакет?

Когда нужно протестировать приватную функцию или приватный метод. Это редкость — обычно лучше сделать метод экспортируемым.

---

## 6. Сложные конструкции

### Сценарий 1: Транзакция с функцией-коллбэком

```go
unitOfWork.On("StartTransaction", context.Background(), mock.MatchedBy(
    func(caller func(walletapp.TransactionalResources) error) bool {
        err := caller(resources)  // вызываем переданную функцию с нашими моками
        return err == nil         // проверяем что функция не упала
    },
)).Return(nil)
```

**Что происходит:**
1. Тестируемый код вызывает `unitOfWork.StartTransaction(ctx, func(r TransactionalResources) error { ... })`
2. Mock перехватывает вызов
3. `mock.MatchedBy` получает функцию-коллбэк
4. Мы вызываем её с нашим мок-объектом `resources`
5. Проверяем что функция не вернула ошибку

Это нужно когда `StartTransaction` принимает функцию, которая работает с ресурсами (кошелёк, транзакции).

### Сценарий 2: Отдельные вызовы Update для разных кошельков

```go
// Первый кошелёк: ID=10, баланс уменьшается 100→90
walletRepo.On("Update", context.Background(), mock.MatchedBy(func(w *domain.Wallet) bool {
    return w.ID == 10 && w.UserID == 1 && w.Balance == 90
})).Return(nil)

// Второй кошелёк: ID=20, баланс увеличивается 50→60
walletRepo.On("Update", context.Background(), mock.MatchedBy(func(w *domain.Wallet) bool {
    return w.ID == 20 && w.UserID == 2 && w.Balance == 60
})).Return(nil)
```

**Как mock выберет нужный?**
- При первом вызове `Update(ctx, wallet)` mock ищет совпадение
- Если `wallet.ID == 10 && wallet.Balance == 90` → первый `.On` подходит
- Если `wallet.ID == 20 && wallet.Balance == 60` → второй `.On` подходит

**Порядок важен!** Если вызовы идут в обратном порядке, mock пройдёт по первому подходящему.

### Сценарий 3: Проверка что метод НЕ был вызван

```go
// Если нужно убедиться что某些 метод не вызывается:
unitOfWork.AssertNotCalled(t, "SomeMethod")
```

### Сценарий 4: HTTP тест с middleware

```go
func setupMiddleware(userID int64) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c *echo.Context) error {
            req := c.Request()
            ctx := context.WithValue(req.Context(), ctxkeys.UserIDKey, userID)
            c.SetRequest(req.WithContext(ctx))
            return next(c)
        }
    }
}

// Использование:
e.POST("/api/wallets", handler.Create, setupMiddleware(10))
```

Middleware подменяет контекст запроса, добавляя туда userID. Это имитирует настоящую middleware, которая извлекает userID из сессионной cookie.

---

## 7. Полезные советы

### Совет 1: Не используй слишком много mock.Anything

```go
// Плохо: тест пройдёт даже с неправильными аргументами
mock.On("GetUser", mock.Anything, mock.Anything).Return(user, nil)

// Хорошо: проверяем конкретные значения
mock.On("GetUser", context.Background(), int64(1)).Return(user, nil)
```

### Совет 2: MatchedBy для сложных структур

```go
// Плохо: нужно знать точное время создания транзакции
txRepo.On("Create", ctx, domain.Transaction{
    WalletID: 10, Amount: 100, CreatedAt: ???,
}).Return(nil)

// Хорошо: проверяем только то, что нам важно
txRepo.On("Create", ctx, mock.MatchedBy(func(tx *domain.Transaction) bool {
    return tx.WalletID == 10 && tx.Amount == 100
})).Return(nil)
```

### Совет 3: Тестируй один сценарий в каждом тесте

```go
func TestDo_Success(t *testing.T) { ... }      // один успех
func TestDo_InsufficientFunds(t *testing.T) { ... } // одна ошибка
```

Не пытайся уместить 5 сценариев в один тест.

### Совет 4: Требования через require, а не assert

```go
require.NoError(t, err)  // если ошибка — тест остановится сразу
assert.NoError(t, err)   // если ошибка — тест продолжится
```

`require` безопаснее — зачем продолжать если уже упали?

### Совет 5: Порядок Arrange-Act-Assert

```go
func TestDo_Success(t *testing.T) {
    // Arrange: настраиваем моки и входные данные
    ownershipChecker := depositmocks.NewMockWalletOwnershipChecker(t)
    cmd := deposit.DepositCommand{...}
    
    // Act: вызываем тестируемую функцию
    got, err := ucase.Do(context.Background(), cmd)
    
    // Assert: проверяем результат
    require.NoError(t, err)
    require.Equal(t, domain.TransactionTypeDeposit, got.Type)
}
```
