# Бронирование досуговых

Система бронирования досуговых комнат в общежитиях ВШЭ — **Go**, **React**, **PostgreSQL**, **Telegram**.

**Деплой:** [booking.dubki.fun](https://booking.dubki.fun)

## Возможности

- Авторизация через корпоративную почту `@hse.ru` / `@edu.hse.ru` (OTP на email)
- Бронирование комнат с выбором временного слота
- Три режима отображения: карточки, таблица, бегущая строка
- Мгновенные обновления у всех клиентов через SSE
- Уведомления в Telegram (группа + личные сообщения)
- Telegram-бот [@dubki_booking_bot](https://t.me/dubki_booking_bot) — `/book`, `/list`, `/cancel`, `/link`
- Привязка Telegram-аккаунта к аккаунту сайта
- Светлая и тёмная тема
- Панель администратора (для студ. советов)
- Форс-пуш для администраторов (обход бизнес-правил)
- Динамическое управление комнатами из панели администратора

## Правила бронирования

| Правило | Значение |
|---------|----------|
| Максимальная длительность | 4 часа |
| Лимит на человека | 1 бронь в день |
| ЧП (частные посиделки) | Разрешены в любой день, в т.ч. пт–вс |
| ЧП на комнату | Не более 3 в день, не более 1 вечером (с 18:00) |

## Технологии

| Слой | Стек |
|------|------|
| Backend | Go 1.23, chi router, pgx |
| База данных | PostgreSQL 16 |
| Frontend | React 18 + TypeScript + Vite + Tailwind CSS |
| Авторизация | Email OTP (`@hse.ru` / `@edu.hse.ru`) |
| Real-time | Server-Sent Events (SSE) |
| Уведомления | Telegram Bot API |
| Инфраструктура | Docker Compose |

## Структура проекта

```
booking/
├── backend/
│   ├── cmd/run/           # Точка входа
│   ├── deploy/migrations/ # SQL-миграции
│   └── internal/
│       ├── application/   # Бизнес-логика (booking, auth, admin, room, tglink)
│       ├── domain/        # Модели и интерфейсы репозиториев
│       └── infrastructure/# Postgres, сервер, Telegram-бот, mailer
├── frontend/
│   └── src/
│       ├── app/           # Корень приложения
│       ├── auth/          # Страницы авторизации
│       ├── pages/         # Страницы приложения
│       ├── admin/         # Панель администратора
│       ├── components/    # UI-компоненты
│       └── api/           # HTTP-клиент
├── docker-compose.yml
└── .env.example
```

## Запуск

### Требования

- Docker + Docker Compose
- SMTP-сервер для OTP (или Yandex 360 / Gmail)
- Telegram-бот (опционально)

### Настройка

```bash
cp .env.example .env
# заполните переменные в .env
```

### Переменные окружения (`.env`)

| Переменная | Описание |
|------------|----------|
| `DB_PASSWORD` | Пароль PostgreSQL |
| `SMTP_HOST` / `SMTP_PORT` | SMTP-сервер для OTP |
| `SMTP_FROM` / `SMTP_PASSWORD` | Учётные данные SMTP |
| `TELEGRAM_BOT_TOKEN` | Токен бота |
| `TELEGRAM_CHAT_ID` | ID чата для уведомлений (опционально) |
| `BOT_INTERNAL_SECRET` | Секрет для внутреннего API бота |
| `BACKEND_INTERNAL_URL` | URL бекенда для бота (обычно `http://dorm_backend:8080`) |

### Запуск контейнеров

```bash
docker compose up -d --build
```

### Применение миграций

Миграции применяются автоматически при старте. При обновлении существующей БД:

```bash
# 002, 003, 004 — применять вручную при наличии старой БД
docker exec -i dorm_db psql -U postgres -d dormitory < backend/deploy/migrations/002_tg_links.sql
docker exec -i dorm_db psql -U postgres -d dormitory < backend/deploy/migrations/003_bot_users.sql
docker exec -i dorm_db psql -U postgres -d dormitory < backend/deploy/migrations/004_booking_tg_msg_id.sql
```

## Команда

| Участник | Роль |
|----------|------|
| **Антон Барабуля** | Backend, инфраструктура, авторизация, Telegram, Docker |
| **Илья Носов** | Frontend, UI, панель администратора, таблицы, calendar |
