# 🛒 Gophermart — дипломный проект Go-разработчика

Gophermart — это система накопления и списания баллов, разработанная в рамках дипломного проекта курса "Go-разработчик".

---

## 🚀 Возможности

- Регистрация и логин пользователей  
- Прием номеров заказов и проверка по алгоритму Луна  
- Асинхронная интеграция с внешним сервисом Accrual  
- Баланс пользователя и история списаний  
- Вывод средств и просмотр истории выводов  
- Работа с PostgreSQL  
- Cookie-based аутентификация с HMAC-подписью  
- Очередь заказов и фоновый воркер  

---

## 🧱 Технологии

- **Go 1.21+**  
- **Gin** — веб-фреймворк  
- **PostgreSQL 13**  
- **Docker / Docker Compose**  
- **bcrypt**, **UUID**, **pgx**  

---

## 📂 Структура проекта

```bash
cmd/gophermart/           # точка входа: main.go
│
├── config/               # парсинг флагов и переменных окружения
├── internal/
│   ├── handlers/         # HTTP-обработчики (Gin)
│   ├── repository/       # интерфейсы и реализация доступа к БД
│   ├── services/         # бизнес-логика
│   ├── async/            # воркер для обработки заказов через Accrual
│   ├── models/           # структуры данных
│   ├── middlewares/      # middleware для авторизации и логирования
├── router/               # настройка маршрутов и middleware
```

---

## ⚙️ Переменные окружения

Задаются в `docker-compose.yml`:

```
DATABASE_DSN=postgres://user:password@db:5432/workdb?sslmode=disable  
ACCRUAL_SYSTEM_ADDRESS=http://host.docker.internal:8080
```

По умолчанию порт API: **8081**

---

## 🐳 Как запустить

1. Клонируйте проект:

```bash
git clone https://github.com/rfruffer/go-musthave-diploma-tpl.git
cd go-musthave-diploma-tpl
```

2. Запустите симулятор Accrual (замените бинарник под вашу ОС):

```bash
./accrual_darwin_arm64 -a localhost:8080
```

3. Запустите Gophermart и PostgreSQL через Docker:

```bash
docker-compose up --build
```

- Gophermart API будет доступен по адресу: `http://localhost:8081`  
- PostgreSQL будет работать на `localhost:5432`

---

## 🔐 Аутентификация

- При логине устанавливается cookie с `user_id`, подписанная `SECRET_KEY`  
- Все защищенные эндпоинты проверяют подпись и наличие куки через middleware

---

## 📬 Основные эндпоинты

| Метод | Путь                          | Описание                     |
|-------|-------------------------------|-------------------------------|
| POST  | `/api/user/register`          | Регистрация пользователя      |
| POST  | `/api/user/login`             | Логин пользователя            |
| POST  | `/api/user/orders`            | Загрузка заказа               |
| GET   | `/api/user/orders`            | История заказов               |
| GET   | `/api/user/balance`           | Баланс пользователя           |
| POST  | `/api/user/balance/withdraw`  | Списание баллов               |
| GET   | `/api/user/withdrawals`       | История выводов               |

---

## 🧪 Примеры запросов

### Регистрация

```http
POST /api/user/register
Content-Type: application/json

{
  "login": "alice",
  "password": "123456"
}
```

### Загрузка заказа

```http
POST /api/user/orders
Cookie: auth=<подписанный user_id>

2377225624
```

---

## 📌 Заметки

- База данных создаётся и инициализируется автоматически  
- Статусы заказов обновляются из внешнего сервиса  
- Если очередь заказов переполнена — заказы не теряются, они будут обработаны при следующем запуске  
- Проект реализован с соблюдением требований диплома
