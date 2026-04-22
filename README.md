# Startup Platform

Полный гайд по запуску проекта **с нуля на новом компьютере** (Windows / macOS / Linux).

> Важно: почти все команды ниже нужно выполнять **из корня проекта**: `.../startup-platform`.

---

## 1) Требования

- Go **1.25+**
- PostgreSQL **14+** (рекомендуется 17)
- Git

Проверка:

- `go version`
- `psql --version`
- `git --version`

---

## 2) Клонирование проекта

```bash
git clone <URL_ВАШЕГО_РЕПОЗИТОРИЯ>
cd startup-platform
```

Все дальнейшие команды (если не написано иное) — **только из `startup-platform`**.

---

## 3) Настройка PostgreSQL

### Вариант A (быстрый, как в проекте по умолчанию)

Создайте БД `startup_platform` и пользователя `postgres` с паролем `12345678`.

### Вариант B (кастомный)

Можно использовать любого пользователя/пароль/БД, но тогда обязательно обновите `DATABASE_URL` в `.env`.

Пример SQL (выполнить в `psql` от суперпользователя):

```sql
CREATE USER postgres WITH PASSWORD '12345678';
ALTER USER postgres CREATEDB;
CREATE DATABASE startup_platform OWNER postgres;
```

---

## 4) Переменные окружения

В проекте уже есть файл `.env` с примером:

```env
PORT=8080
DATABASE_URL=postgres://postgres:12345678@localhost:5432/startup_platform?sslmode=disable
JWT_SECRET=CHANGE_ME_TO_A_RANDOM_LONG_SECRET
UPLOAD_DIR=./uploads
```

Рекомендуется заменить `JWT_SECRET` на свой длинный случайный ключ.

> Примечание: приложение использует fallback на значения по умолчанию из `internal/config/config.go`, но для нового окружения лучше держать всё в `.env`.

---

## 5) Установка зависимостей Go

```bash
go mod download
```

---

## 6) Запуск приложения

### Рекомендуемый dev-запуск

```bash
go run ./cmd/server/
```

### Или через сборку бинаря

```bash
go build -o startup-platform.exe ./cmd/server/
./startup-platform.exe
```

После старта:

- `http://localhost:8080`

---

## 7) Что создаётся автоматически при старте

При запуске `cmd/server/main.go` приложение автоматически:

1. подключается к БД,
2. прогоняет миграции,
3. запускает сиды (`internal/seed/seed.go`) **только если таблица users пуста**,
4. создаёт папку загрузок (`./uploads`), если её нет.

---

## 8) Тесты

Из корня проекта:

```bash
go test ./tests/ -count=1 -timeout 120s
```

Расширенный запуск с покрытием:

```bash
go test ./tests/ -count=1 -timeout 120s -coverprofile cover.out -coverpkg startup-platform/internal/config,startup-platform/internal/database,startup-platform/internal/handlers,startup-platform/internal/middleware,startup-platform/internal/models
```

---

## 9) Тестовые учётные записи (сиды)

### Главный админ

- Email: `admin@startup.by`
- Nickname: `admin`
- Password: `admin123`
- Role: `admin`

### Все остальные сид-пользователи

Пароль для всех ниже: **`user123`**

| Email | Nickname | Role |
|---|---|---|
| ivan@startup.by | ivan_dev | user |
| maria@startup.by | maria_design | premium |
| alex@startup.by | alex_pm | user |
| olga@startup.by | olga_expert | expert |
| dmitry@startup.by | dmitry_ml | user |
| anna@startup.by | anna_mark | premium |
| sergey@startup.by | sergey_back | user |
| elena@startup.by | elena_data | user |
| nikita@startup.by | nikita_front | user |
| kate@startup.by | kate_hr | user |
| pavel@startup.by | pavel_devops | user |
| yulia@startup.by | yulia_content | user |
| maxim@startup.by | maxim_sec | expert |
| natasha@startup.by | natasha_qa | user |
| artem@startup.by | artem_game | user |
| light@startup.by | light_startup | premium |
| viktor@startup.by | viktor_mobile | user |
| daria@startup.by | daria_ux | user |
| roman@startup.by | roman_cto | user |
| polina@startup.by | polina_ai | user |

---

## 10) Частые проблемы и решения

### 10.1 `go.mod file not found`
Вы запустили команду не из той директории.

Решение: перейти в корень проекта:

```bash
cd <путь>/startup-platform
```

### 10.2 Сервер не стартует / шаблоны не находятся
Сервер должен запускаться из `startup-platform`, потому что используются относительные пути `./web/templates` и `./web/static`.

### 10.3 Ошибка подключения к PostgreSQL
Проверьте:

- PostgreSQL запущен
- существует БД `startup_platform`
- `DATABASE_URL` корректный

---

## 11) Минимальный smoke-check после запуска

1. Войти как `admin@startup.by / admin123`
2. Открыть `/admin`
3. Открыть `/profile/me`
4. Для premium/expert аккаунта открыть `/premium-stats`
5. Создать пост (`/post/new`)

---

## 12) Структура запуска (коротко)

- Рабочая директория команд: **`startup-platform`**
- Запуск: `go run ./cmd/server/`
- Порт по умолчанию: `8080`
- БД по умолчанию: `startup_platform`

Готово ✅
