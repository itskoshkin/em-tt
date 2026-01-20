# subscription-aggregator-service

REST-сервис для агрегации данных об онлайн-подписках пользователей

<details>
<summary><h2>ТЗ</h2></summary>

Задача: спроектировать и реализовать REST-сервис для агрегации данных об
онлайн-подписках пользователей.
Требования:
1. Выставить HTTP-ручки для CRUDL-операций над записями о подписках. Каждая запись
   содержит:
    1. Название сервиса, предоставляющего подписку
    2. Стоимость месячной подписки в рублях
    3. ID пользователя в формате UUID
    4. Дата начала подписки (месяц и год)
    5. Опционально дата окончания подписки
2. Выставить HTTP-ручку для подсчета суммарной стоимости всех подписок за выбранный
   период с фильтрацией по id пользователя и названию подписки
3. СУБД – PostgreSQL. Должны быть миграции для инициализации базы данных
4. Покрыть код логами
5. Вынести конфигурационные данные в .env/.yaml-файл
6. Предоставить swagger-документацию к реализованному API
7. Запуск сервиса с помощью docker compose

Примечания:
1. Проверка существования пользователя не требуется. Управление пользователями вне
   зоны ответственности вашего сервиса
2. Стоимость любой подписки – целое число рублей, копейки не учитываются

Пример тела запроса на создание записи о подписке:
```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025"
}
```

</details>

<details>
<summary><h2>Запуск</h2></summary>

### Docker Compost

1.  Клонируйте репозиторий
    ```bash
    git clone https://github.com/itskoshkin/em-tt.git && cd em-tt
    ```

2.  Запустите проект
    ```bash
    docker compose up --build
    ```
    Эта команда:
    *   Поднимет контейнер с базой данных
    *   Запустит контейнер-мигратор (`goose`), который дождётся готовности БД, накатит схему и выйдет
    *   Запустит контейнер с приложением на порту **8080**

3.  Откройте Swagger UI в браузере
    [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

</details>

<details>
<summary><h2>API</h2></summary>

### Список эндпоинтов

- `POST /api/v1/subscriptions` - Создать подписку
- `GET /api/v1/subscriptions/{id}` - Получить подписку по ID
- `PUT /api/v1/subscriptions/{id}` - Обновить подписку
- `DELETE /api/v1/subscriptions/{id}` - Удалить подписку
- `GET /api/v1/subscriptions` - Список подписок (+ фильтры `user_id` и `service_name`)
- `GET /api/v1/subscriptions/total` - Расчет стоимости за период

<details>
<summary><h3>Примеры запросов (cURL)</h3></summary>

#### Создать подписку
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 299,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "start_date": "01-2024"
  }'
```

#### Получить общую стоимость
```bash
curl "http://localhost:8080/api/v1/subscriptions/total?user_id=550e8400-e29b-41d4-a716-446655440000&start_date=01-2024&end_date=12-2024"
```

</details>

Полная документация и отправка запросов доступна в [Swagger UI](http://localhost:8080/swagger/index.html)

</details>

<details>
<summary><h2>Тесты</h2></summary>

<details>
<summary><h3>Unit</h3></summary>

#### Покрытие

| Пакет                      | Что тестируется                                                                                |
|----------------------------|------------------------------------------------------------------------------------------------|
| `internal/api/controllers` | HTTP handlers: статус-коды, формат ответов, обработка ошибок                                   |
| `internal/api/models`      | Валидация request-моделей, парсинг дат                                                         |
| `internal/service`         | Бизнес-логика: CRUD операции, валидация, расчёт стоимости подписок                             |
| `internal/utils/dates`     | Парсинг дат из строки → `time.Time`                                                            |

<details>
<summary><h4>Команды для запуска</h4></summary>

```bash
# Все тесты
go test ./...
```

```bash
# С подробным выводом
go test ./... -v
```

```bash
# Конкретный пакет
go test ./internal/service/...
```

```bash
# Конкретный тест
go test ./internal/service -run TestCalculateSubscriptionCost
```

```bash
# С покрытием
go test ./... -cover 
```

```bash
# Генерация HTML-отчёта покрытия
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
```

```bash
# С race detector
go test ./... -race
```

</details>

</details>

<details>
<summary><h3>Integration</h3></summary>

Тесты storage-слоя с реальной PostgreSQL в Docker (testcontainers, требует Docker)

| Тест                       | Описание                            |
|----------------------------|-------------------------------------|
| `TestCreateSubscription`   | Создание подписки в БД              |
| `TestGetSubscriptionByID`  | Получение по ID, проверка NotFound  |
| `TestUpdateSubscription`   | Обновление полей                    |
| `TestDeleteSubscription`   | Soft-delete                         |
| `TestListSubscriptions`    | Фильтрация по user_id, service_name |
| `TestConcurrentOperations` | Конкурентные операции               |

```bash
go test ./tests/integration/... -tags=integration -v
```

</details>

<details>
<summary><h3>E2E</h3></summary>

Полный HTTP flow: запрос → роутер → контроллер → сервис → БД → ответ (требует Docker)

| Тест                       | Описание                                            |
|----------------------------|-----------------------------------------------------|
| `TestFullCRUDFlow`         | Полный цикл: Create → Read → Update → List → Delete |
| `TestTotalCostCalculation` | Расчёт стоимости за период                          |
| `TestValidationErrors`     | Проверка 400 на невалидные данные                   |
| `TestNotFound`             | Проверка 404                                        |
| `TestListWithFilters`      | Фильтрация списка                                   |
| `TestResponseTimes`        | Время ответа < 500ms                                |

```bash
go test ./tests/e2e/... -tags=e2e -v
```

</details>

<details>
<summary><h3>Smoke</h3></summary>

Базовая проверка что сервис запускается и отвечает

#### Запуск

```bash
go test ./tests/e2e/... -tags=e2e -run TestSmoke -v
```

</details>

<details>
<summary><h3>Load/Perf</h3></summary>

Нагрузочные тесты с метриками: RPS, latency (p50/p95/p99), error rate (требует Docker)

| Тест                           | Описание                               |
|--------------------------------|----------------------------------------|
| `TestLoad_CreateSubscriptions` | 10 воркеров × 100 запросов на создание |
| `TestLoad_ListSubscriptions`   | 20 воркеров × 50 запросов на список    |
| `TestLoad_MixedOperations`     | Смешанная нагрузка 10 сек              |

**Пример вывода:**
```
Load Test Results:
  Total Requests:    1000
  Successful:        998
  Errors:            2
  Requests/sec:      312.50
  Avg Latency:       28ms
  P95 Latency:       85ms
  P99 Latency:       120ms
```

```bash
go test ./tests/load/... -tags=load -v
```

</details>

#### Запустить все тесты

```bash
go test ./... -tags=integration,e2e,load
```

</details>

<details>
<summary><h2>TODO</h2></summary>

- [x] Добавить пагинацию для списка подписок
- [x] Выполнять агрегацию/фильтрацию в БД, а не в коде
- [ ] Добавить gracefull shutdown
- [x] Покрыть код юнит-тестами
- [ ] Настроить пайплайн CI/CD и задеплоить на сервер
- [ ] Добавить Request ID (~связать HTTP запрос с операцией в БД)
- [ ] Makefile?
- [ ] Minors
  - [ ] Перенести Postgres из `pkg/` в `internal/`
  - [ ] Возвращать `204 NC` вместо `200 OK` при `DELETE`

</details>
