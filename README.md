# Aspirin Production

Учебное веб-приложение для управления этапами синтеза аспирина: создание черновиков, загрузка медиа, публикация, фильтрация по длительности и лайки. Доступны REST API и серверные HTML-страницы.

**Стек:** Go 1.25+, Gin, GORM, PostgreSQL 15, MinIO, Docker Compose.

## Локальный запуск

1. Установите Go и Docker Compose, затем запустите инфраструктуру:

   ```sh
   docker compose up -d
   ```

2. Создайте `.env` в корне проекта с параметрами локальных сервисов:

   ```dotenv
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=student
   DB_PASS=root
   DB_NAME=aspirin_db
   MINIO_ENDPOINT=localhost:9000
   MINIO_ACCESS_KEY=root
   MINIO_SECRET_KEY=rootpassword
   MINIO_BUCKET_NAME=aspirin-media
   ```

3. Подготовьте таблицы `aspirin_production_users`, `aspirin_stages` и `aspirin_stage_likes` по моделям в [repository.go](internal/app/repository/repository.go). Миграции и автоматическое создание таблиц в проекте отсутствуют. Для лайков требуется уникальность пары `(production_user_id, aspirin_stage_id)`.
4. Для отображения загруженных медиа разрешите публичное чтение объектов в бакете MinIO. Бакет создаётся приложением при запуске, но политика доступа автоматически не задаётся.
5. Дождитесь готовности PostgreSQL и MinIO и запустите приложение из корня проекта:

   ```sh
   go run ./cmd/aspirin
   ```

Веб-интерфейс: [localhost:8080/aspirin-synthesis-stages](http://localhost:8080/aspirin-synthesis-stages). MinIO Console: [localhost:9001](http://localhost:9001). Adminer: [localhost:8081](http://localhost:8081).

## REST API

Все маршруты используют префикс `/api`.

| Метод | Маршрут | Назначение |
| --- | --- | --- |
| GET | `/aspirin-stages` | Опубликованные этапы; фильтр `max_synthesis_duration_minutes` |
| GET | `/aspirin-stages/draft` | Черновик текущего пользователя или `data: null` |
| POST | `/aspirin-stages` | Создание черновика с медиа (`multipart/form-data`) |
| PUT | `/aspirin-stages/draft/publication` | Публикация черновика |
| GET | `/aspirin-stages/feed` | Первый опубликованный этап |
| GET | `/aspirin-stages/feed/:id` | Этап ленты; `?next=true` возвращает следующий |
| POST | `/aspirin-stages/:id/like` | Установка или снятие лайка: `{"like": 1}` / `{"like": 0}` |
| DELETE | `/aspirin-stages/:id` | Логическое удаление своего этапа |
| POST | `/users` | Регистрация: `username`, `password` |
| POST | `/authentication`, `/deauthentication` | Заглушки аутентификации и выхода |

При создании передаётся `aspirin_synthesis_stage_name`, опционально — файлы `aspirin_synthesis_stage_image` (до 10 МБ) и `aspirin_synthesis_stage_video` (до 50 МБ). Для публикации используются JSON-поля `aspirin_stage_description`, `synthesis_duration_minutes` и `pure_aspirin_yield_percent`.

Этапы переходят из `draft` в `published`; удаление устанавливает статус `deleted`. Медиа хранятся в MinIO, имена объектов — в PostgreSQL.

## Текущие ограничения

Текущий пользователь фиксирован (`ID = 1`), а аутентификация ещё не реализована. Пароли сохраняются без хеширования. SSR-маршрут удаления не проверяет владельца этапа. Проект предназначен для учебного локального использования.
