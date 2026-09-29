# Aspirin Production

Учебное веб-приложение по предметной области производства аспирина.

## Предметная область

Услуги приложения — этапы синтеза аспирина.

Пользователь может создавать этап синтеза в статусе черновика, добавлять изображение и короткое видео, заполнять описание этапа, его продолжительность и процент выхода чистого вещества, после чего публиковать услугу.

В ЛР-3 реализован REST API для дальнейшего использования SPA.

## Технологии

- Go
- Gin
- GORM
- PostgreSQL
- MinIO
- Docker Compose
- HTML SSR
- CSS
- Adminer

# REST API

Все методы ЛР-3 начинаются с `/api`.

## Домен этапов синтеза

### GET `/api/aspirin-stages`

Получение списка опубликованных этапов синтеза.

Удалённые и черновые записи клиенту не передаются.

Поддерживается фильтрация:

```text
?max_synthesis_duration_minutes=60
```

Каждый объект содержит поле `is_creator`:

- `1` — текущий пользователь является создателем;
- `0` — текущий пользователь не является создателем.

### GET `/api/aspirin-stages/draft`

Получение единственного черновика текущего пользователя.

ID черновика клиентом не передаётся.

Если черновика нет:

```json
{
  "status": "success",
  "data": null
}
```

### POST `/api/aspirin-stages`

Создание нового черновика.

Content-Type:

```text
multipart/form-data
```

Поля:

- `aspirin_synthesis_stage_name`
- `aspirin_synthesis_stage_image`
- `aspirin_synthesis_stage_video`

Изображение и видео передаются как файлы.

Файлы хранятся в MinIO.

В PostgreSQL сохраняются только сгенерированные сервером имена файлов.

Системные поля клиентом не передаются.

Статус автоматически устанавливается в `draft`.

Создатель определяется через singleton текущего пользователя.

Дата создания задаётся сервером.

### PUT `/api/aspirin-stages/draft/publication`

Публикация единственного черновика текущего пользователя.

JSON:

```json
{
  "aspirin_stage_description": "Описание этапа",
  "synthesis_duration_minutes": 45,
  "pure_aspirin_yield_percent": 82.5
}
```

При публикации:

- `draft` меняется на `published`;
- устанавливается дата формирования;
- повторная публикация невозможна;
- возврат опубликованной записи обратно в `draft` отсутствует.

### GET `/api/aspirin-stages/feed`

Получение первого опубликованного элемента ленты.

### GET `/api/aspirin-stages/feed/:id`

Получение указанного опубликованного элемента ленты.

Пример:

```text
GET /api/aspirin-stages/feed/5
```

### GET `/api/aspirin-stages/feed/:id?next=true`

Получение следующего опубликованного элемента ленты.

### POST `/api/aspirin-stages/:id/like`

Поставить или снять лайк от текущего пользователя.

JSON:

```json
{
  "like": 1
}
```

Значения:

- `1` — поставить лайк;
- `0` — убрать лайк.

Один пользователь не может поставить одной услуге несколько лайков.

### DELETE `/api/aspirin-stages/:id`

Логическое удаление услуги.

Удалять можно только услуги текущего пользователя.

Физически запись из PostgreSQL не удаляется.

Статус меняется на:

```text
deleted
```

Удалённые записи больше не передаются клиенту.

# Домен пользователей

### POST `/api/users`

Регистрация нового пользователя.

JSON:

```json
{
  "username": "student",
  "password": "password"
}
```

ID пользователя вычисляется PostgreSQL.

### POST `/api/authentication`

Заглушка аутентификации для ЛР-4.

### POST `/api/deauthentication`

Заглушка деавторизации для ЛР-4.

# Singleton текущего пользователя

В ЛР-3 полноценная авторизация отсутствует.

Текущий пользователь фиксирован константой:

```go
const currentAspirinProductionUserID = 1

func GetCurrentAspirinProductionUserID() int {
	return currentAspirinProductionUserID
}
```

Все методы, зависящие от пользователя, получают его ID через эту функцию.

# Таблицы PostgreSQL

## aspirin_production_users

| Поле | Назначение |
|---|---|
| `production_user_id` | Первичный ключ пользователя |
| `production_username` | Уникальное имя пользователя |
| `production_user_password` | Пароль пользователя |

## aspirin_stages

| Поле | Назначение |
|---|---|
| `aspirin_stage_id` | Первичный ключ этапа |
| `aspirin_stage_name` | Название |
| `aspirin_stage_description` | Описание |
| `aspirin_stage_status` | `draft / published / deleted` |
| `aspirin_stage_image_url` | Имя изображения в MinIO |
| `aspirin_stage_video_url` | Имя видео в MinIO |
| `synthesis_duration_minutes` | Длительность этапа |
| `pure_aspirin_yield_percent` | Выход чистого вещества |
| `aspirin_stage_created_at` | Дата создания |
| `aspirin_stage_formed_at` | Дата публикации |
| `production_user_id` | Создатель услуги |

## aspirin_stage_likes

| Поле | Назначение |
|---|---|
| `aspirin_stage_like_id` | Первичный ключ лайка |
| `production_user_id` | Пользователь |
| `aspirin_stage_id` | Этап синтеза |

Для пары:

```text
production_user_id + aspirin_stage_id
```

установлено ограничение уникальности.

# Хранение медиа

Изображения и видео хранятся в MinIO.

В PostgreSQL сохраняется только имя объекта.

Пример значения в БД:

```text
image_1760001234567_ab12cd34.png
```

При сериализации backend преобразует имя в URL:

```text
http://localhost:9000/aspirin-media/image_1760001234567_ab12cd34.png
```

# Статусы услуги

Допустимый основной переход:

```text
draft -> published
```

Также услуга может быть логически удалена:

```text
draft/published -> deleted
```

Возврат из `published` в `draft` не реализован.

Удалённые записи клиенту не передаются.

# SSR ЛР-2

Старые SSR-маршруты сохранены:

- GET `/aspirin-synthesis-stage`
- GET `/aspirin-synthesis-stage-draft`
- GET `/aspirin-synthesis-stages`
- POST `/aspirin-synthesis-stage-draft`
- POST `/aspirin-synthesis-stage-publication`
- POST `/aspirin-synthesis-stage-deletion`

Таким образом ЛР-3 добавляет API, не удаляя функциональность ЛР-2.