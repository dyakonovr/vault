# Vault

Микросервис кошельков и переводов. Учебный проект для отработки конкурентных операций, DDD, GORM, MongoDB и нагрузочного тестирования.

## Архитектура

Проект следует принципам Domain-Driven Design (тактический дизайн) и чистой архитектуры.

- **`internal/domain`** — сущности, value objects, доменные ошибки. Не зависит от инфраструктуры.
- **`internal/application`** — сценарии использования (use cases). Оркеструет домен и репозитории.
- **`internal/infrastructure`** — реализации репозиториев (Postgres, Mongo), HTTP-транспорт, менеджеры блокировок.
- **`cmd/app`** — точка входа, композиция зависимостей.

## Стек

- **Go** 1.22+
- **PostgreSQL** (GORM)
- **MongoDB** (событийный аудит-лог)
- **Echo** (HTTP-фреймворк)
- **golang-migrate** (миграции)
- **Docker / docker-compose**

## Запуск

```bash
# 1. Запустить PostgreSQL и MongoDB
docker-compose up -d

# 2. Применить миграции
go run cmd/migrate/main.go up

# 3. Запустить сервер
go run cmd/app/main.go