# Aspirin Production

Учебный backend для управления этапами синтеза аспирина: черновики, публикации, медиа, фильтрация и лайки. ЛР-4 добавляет JWT-аутентификацию, Redis blacklist и Swagger.

**Стек:** Go 1.25+, Gin, GORM, PostgreSQL 15, MinIO, Redis 7, Swagger.

## Запуск

1. Создайте `.env` по `.env.example` (в существующий файл добавьте недостающие параметры), заполните `JWT_SECRET` случайной строкой не менее 32 байт (`openssl rand -hex 32`).
2. Запустите хранилища: `docker compose up -d`.
3. Подготовьте БД: `go run ./cmd/migrate`. Команда создаёт таблицы и индексы и переводит старые пароли ЛР-3 в bcrypt, сохраняя сами пароли для входа. При нескольких черновиках одного пользователя сначала устраните дубли через Adminer.
4. Запустите сервер из корня проекта: `go run ./cmd/aspirin`.

[Swagger](http://localhost:8080/swagger/index.html) · [Плитка SSR](http://localhost:8080/aspirin-synthesis-stages) · [MinIO](http://localhost:9001) · [Adminer](http://localhost:8081)

Бакет MinIO создаётся при запуске. Для отображения медиа разрешите публичное чтение объектов через MinIO. Настройки подключения находятся в `.env.example`; срок JWT задаётся `JWT_TTL` (по умолчанию `1h`).

## API

Префикс маршрутов — `/api`. В Swagger описаны параметры и ответы всех методов.

| Метод | Маршрут | Назначение | Доступ |
| --- | --- | --- | --- |
| GET | `/aspirin-stages` | Публикации; фильтр `max_synthesis_duration_minutes` | Все |
| GET | `/aspirin-stages/feed` | Первый этап ленты | Все |
| GET | `/aspirin-stages/feed/:id` | Этап или следующий (`?next=true`) | Все |
| GET | `/aspirin-stages/draft` | Свой черновик или `data: null` | JWT |
| POST | `/aspirin-stages` | Черновик с медиа (`multipart/form-data`) | JWT |
| PUT | `/aspirin-stages/draft/publication` | Публикация своего черновика | JWT |
| DELETE | `/aspirin-stages/:id` | Логическое удаление своего этапа | JWT |
| POST | `/aspirin-stages/:id/like` | Лайк: `{"like": 1}` / `{"like": 0}` | JWT |
| POST | `/users` | Регистрация: `username`, `password` | Все |
| POST | `/authentication` | Вход: `username`, `password`; возвращает JWT | Все |
| POST | `/deauthentication` | Выход и отзыв текущего JWT | JWT |

После входа скопируйте `access_token`. В Swagger нажмите **Authorize** и введите `Bearer <access_token>`; в Postman/Insomnia используйте такой же заголовок `Authorization`. Для персональных признаков в публичных GET заполните необязательное поле `Authorization` в Swagger. `expires_in` указан в секундах. После выхода JWT отклоняется; новая аутентификация выдаёт новый токен.

В плитке и ленте возвращаются `is_creator` и `is_liked` (0/1) для текущего пользователя; для гостя оба равны 0. Публичное чтение доступно также с недействительным токеном. Защищённые запросы без действительного JWT возвращают 401, при недоступности Redis — 503.

Создание: `aspirin_synthesis_stage_name` и необязательные файлы `aspirin_synthesis_stage_image` (10 МБ), `aspirin_synthesis_stage_video` (50 МБ). Публикация: JSON с `aspirin_stage_description`, `synthesis_duration_minutes` (> 0), `pure_aspirin_yield_percent` (0–100). Создатель, статусы и даты определяются сервером.

## Данные

Медиа хранятся в MinIO; PostgreSQL содержит пользователей, этапы и лайки. Черновик один на пользователя, лайк один на пару пользователь–этап. Переходы: `draft → published`, `draft/published → deleted`.

Redis хранит отозванные JWT по ключам `aspirin.production.jwt.blacklist.<jti>` до истечения токена. Для просмотра: `docker compose exec redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli --scan --pattern "aspirin.production.jwt.blacklist.*"'`.

Обновление Swagger после изменения аннотаций: `go tool swag init -g cmd/aspirin/main.go --parseInternal`.

SSR-маршруты чтения сохранены; черновик и изменения также требуют заголовок Bearer JWT. Браузерные HTML-формы самостоятельно этот заголовок не добавляют; сценарий ЛР-4 выполняется через Swagger или Postman/Insomnia.
