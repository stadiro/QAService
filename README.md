## QAService — API‑сервис для вопросов и ответов

### Описание

REST API‑сервис для работы с вопросами и ответами.  
Две основные сущности:

- **Question**
  - `id: int`
  - `text: string`
  - `created_at: datetime`
- **Answer**
  - `id: int`
  - `question_id: int`
  - `user_id: string`
  - `text: string`
  - `created_at: datetime`

Основная логика:

- **Нельзя** создать ответ на несуществующий вопрос.
- Один и тот же пользователь может оставить несколько ответов к одному вопросу.
- При удалении вопроса его ответы удаляются **каскадно** (через FK `ON DELETE CASCADE`).

Доступные методы API:

- **GET** `/questions` — список всех вопросов  
- **POST** `/questions` — создать вопрос  
- **GET** `/questions/{id}` — получить вопрос и все его ответы  
- **DELETE** `/questions/{id}` — удалить вопрос (и его ответы)  
- **POST** `/questions/{id}/answers` — добавить ответ к вопросу  
- **GET** `/answers/{id}` — получить конкретный ответ  
- **DELETE** `/answers/{id}` — удалить ответ  

---

### Запуск через Docker Compose


```bash
docker-compose up --build
```

Будут подняты два сервиса:

- **db** — PostgreSQL (пользователь/пароль/БД: `postgres/postgres/qa_service`)
- **app** — Go‑приложение на порту `8080`, которое при старте прогоняет миграции через `goose`

После успешного старта API будет доступен по адресу `http://localhost:8080`.

Быстрая проверка:

```bash
curl http://localhost:8080/questions
```

Если всё ок, вернётся пустой JSON‑массив `[]`.

---

### Примеры запросов

**Создать вопрос**

```bash
curl -X POST http://localhost:8080/questions ^
  -H "Content-Type: application/json" ^
  -d "{\"text\": \"Что такое GORM?\"}"
```

**Добавить ответ к вопросу с id=1**

```bash
curl -X POST http://localhost:8080/questions/1/answers ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\": \"user-123\", \"text\": \"Это ORM для Go\"}"
```

**Получить вопрос с ответами**

```bash
curl http://localhost:8080/questions/1
```
