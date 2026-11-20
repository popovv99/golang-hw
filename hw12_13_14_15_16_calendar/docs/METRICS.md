# Метрики сервиса Calendar

Для мониторинга сервиса Calendar реализован endpoint `/metrics`, который возвращает метрики в формате Prometheus.

## Endpoint

```
GET /metrics
Content-Type: text/plain; version=0.0.4
```

## Запуск сбора метрик

В `deployments/docker-compose.yaml` добавлен сервис `prometheus`. После запуска стека:

```bash
docker compose -f deployments/docker-compose.yaml up -d
```

Интерфейс Prometheus доступен по адресу:

```
http://localhost:9090
```

Prometheus автоматически собирает метрики со следующих endpoint'ов каждые 15 секунд:

| Сервис | Endpoint внутри Docker-сети | Внешний URL (с `ports` из compose) |
|---|---|---|
| `calendar` | `calendar:8080/metrics` | `http://localhost:8888/metrics` |
| `scheduler` | `scheduler:8081/metrics` | `http://localhost:8081/metrics` |
| `storer` | `storer:8082/metrics` | `http://localhost:8082/metrics` |

## Список метрик

### HTTP API

| Метрика | Тип | Лейблы | Описание |
|---------|-----|--------|----------|
| `calendar_http_requests_total` | Counter | `method`, `path`, `status` | Количество HTTP-запросов к API |
| `calendar_http_request_duration_seconds` | Histogram | `method`, `path` | Время обработки HTTP-запросов в секундах |


### Бизнес-операции

| Метрика | Тип | Описание |
|---------|-----|----------|
| `calendar_events_created_total` | Counter | Количество успешно созданных событий |
| `calendar_events_updated_total` | Counter | Количество успешно обновленных событий |
| `calendar_events_deleted_total` | Counter | Количество успешно удалённых событий |
| `calendar_notifications_sent_total` | Counter | `status` (`success` / `error`) | Количество отправленных уведомлений шедулером |
| `calendar_notifications_saved_total` | Counter | `status` (`success` / `error`) | Количество сохранённых уведомлений сторером |

### Фоновые задачи

| Метрика | Тип | Лейблы / примечание | Описание |
|---------|-----|---------------------|----------|
| `calendar_scheduler_runs_total` | Counter | — | Количество итераций планировщика |
| `calendar_scheduler_errors_total` | Counter | — | Количество ошибок при выполнении итерации планировщика |
| `calendar_scheduler_last_success_timestamp_seconds` | Gauge | Unix timestamp | Время последней успешной итерации планировщика |
| `calendar_storer_runs_total` | Counter | — | Количество обработанных сообщений сторером |
| `calendar_storer_errors_total` | Counter | — | Количество ошибок при обработке сообщений сторером |
| `calendar_storer_last_success_timestamp_seconds` | Gauge | Unix timestamp | Время последнего успешного сохранения уведомления сторером |

## PromQL-руководство

В веб-интерфейсе Prometheus (`http://localhost:9090/graph`) вводите запрос, выбирайте вкладку **Graph** (график) или **Table** (таблица), задавайте интервал в правом верхнем углу и нажимайте **Execute**.

### HTTP API — количество и скорость запросов

Метрика: `calendar_http_requests_total` (Counter), лейблы: `method`, `path`, `status`.

| Что увидеть | PromQL | Пояснение |
|---|---|---|
| Текущее количество запросов | `calendar_http_requests_total` | Таблица: по строкам `method`/`path`/`status`, по столбцам значения |
| RPS по всем endpoint'ам | `rate(calendar_http_requests_total[5m])` | Скорость роста счётчика, запросов в секунду |
| RPS по конкретному пути | `rate(calendar_http_requests_total{path="/events"}[5m])` | Фильтр по `path` |
| RPS с разбивкой по статусам | `sum by (status) (rate(calendar_http_requests_total[5m]))` | Сумма по `status` |
| Доля ошибок (4xx/5xx) | `sum(rate(calendar_http_requests_total{status=~"4..\|5.."}[5m])) / sum(rate(calendar_http_requests_total[5m]))` | `0..1`, где `1` — 100% ошибок |
| Успешные POST /events | `calendar_http_requests_total{method="POST", path="/events", status="201"}` | Счётчик успешных созданий событий |

### HTTP API — длительность запросов

Метрика: `calendar_http_request_duration_seconds` (Histogram), лейблы: `method`, `path`.

Histogram в Prometheus разбивается на три суффикса:

- `calendar_http_request_duration_seconds_bucket{le="..."}` — количество попаданий в buckets
- `calendar_http_request_duration_seconds_sum` — сумма всех наблюдений
- `calendar_http_request_duration_seconds_count` — общее количество наблюдений

| Что увидеть | PromQL | Пояснение |
|---|---|---|
| Общее количество измерений | `calendar_http_request_duration_seconds_count` | Сколько раз мы измерили latency |
| Средняя latency | `sum(rate(calendar_http_request_duration_seconds_sum[5m])) / sum(rate(calendar_http_request_duration_seconds_count[5m]))` | Среднее время ответа в секундах |
| 95-й перцентиль по всем запросам | `histogram_quantile(0.95, sum(rate(calendar_http_request_duration_seconds_bucket[5m])) by (le))` | 95% запросов быстрее этого значения |
| 95-й перцентиль по пути | `histogram_quantile(0.95, sum(rate(calendar_http_request_duration_seconds_bucket[5m])) by (le, path))` | p95 с разбивкой по `path` |
| p95 для POST /events | `histogram_quantile(0.95, sum(rate(calendar_http_request_duration_seconds_bucket{method="POST",path="/events"}[5m])) by (le))` | p95 только для создания событий |

### Бизнес-операции

| Что увидеть | PromQL | Пояснение |
|---|---|---|
| Сколько событий создано | `calendar_events_created_total` | Текущее значение |
| Сколько событий обновлено | `calendar_events_updated_total` | Текущее значение |
| Сколько событий удалено | `calendar_events_deleted_total` | Текущее значение |
| Скорость создания событий | `rate(calendar_events_created_total[5m])` | Событий в секунду |
| Скорость обновлений/удалений | `rate(calendar_events_updated_total[5m])` и `rate(calendar_events_deleted_total[5m])` | Аналогично |

### Уведомления

Метрики: `calendar_notifications_sent_total` (scheduler) и `calendar_notifications_saved_total` (storer), лейбл `status`.

| Что увидеть | PromQL | Пояснение |
|---|---|---|
| Всего успешных/ошибочных | `calendar_notifications_sent_total{status="success"}` и `{status="error"}` | Счётчики |
| RPS уведомлений по статусу | `sum by (status) (rate(calendar_notifications_sent_total[5m]))` | График успехов и ошибок |
| Доля ошибок при отправке | `sum(rate(calendar_notifications_sent_total{status="error"}[5m])) / sum(rate(calendar_notifications_sent_total[5m]))` | Alert: доля > 0.01 |
| Всего сохранено уведомлений | `calendar_notifications_saved_total{status="success"}` | Сторер сохранил в БД |
| Доля ошибок сохранения | `sum(rate(calendar_notifications_saved_total{status="error"}[5m])) / sum(rate(calendar_notifications_saved_total[5m]))` | Ошибки storer'а |

### Фоновые сервисы

Метрики: `calendar_scheduler_runs_total`, `calendar_scheduler_errors_total`, `calendar_scheduler_last_success_timestamp_seconds`, `calendar_storer_runs_total`, `calendar_storer_errors_total`, `calendar_storer_last_success_timestamp_seconds`.

| Что увидеть | PromQL | Пояснение |
|---|---|---|
| Сколько итераций выполнено | `calendar_scheduler_runs_total` и `calendar_storer_runs_total` | Счётчики |
| Скорость итераций | `rate(calendar_scheduler_runs_total[5m])` | Итераций в секунду |
| Доля ошибок планировщика | `sum(rate(calendar_scheduler_errors_total[5m])) / sum(rate(calendar_scheduler_runs_total[5m]))` | Процент ошибок |
| Доля ошибок сторера | `sum(rate(calendar_storer_errors_total[5m])) / sum(rate(calendar_storer_runs_total[5m]))` | Процент ошибок |
| "Пропал" ли планировщик | `time() - calendar_scheduler_last_success_timestamp_seconds` | Если значение > `SCHEDULER_INTERVAL` в несколько раз — сервис завис |
| "Пропал" ли сторер | `time() - calendar_storer_last_success_timestamp_seconds` | Аналогично |

### Health scraping

Prometheus сам добавляет служебные метрики:

| Что увидеть | PromQL | Пояснение |
|---|---|---|
| Доступен ли target | `up{job=~"calendar\|scheduler\|storer"}` | `1` — доступен, `0` — нет. Показывает состояние всех трёх target'ов сразу |
| Все target'ы живы | `min(up{job=~"calendar\|scheduler\|storer"})` | `1` — все доступны, `0` — хотя бы один упал (удобно для alert) |
| Длительность сбора | `scrape_duration_seconds{job="calendar"}` | Сколько длится scrape |
| Размер ответа | `scrape_samples_scraped{job="calendar"}` | Сколько sample'ов пришло |

### Как читать результат

- **Graph**: линия по времени. Подходит для `rate(...)` и `histogram_quantile(...)`.
- **Table**: мгновенное значение в точке времени. Удобно для счётчиков (`calendar_events_created_total`) и для `up`.
- **Range** (правый верхний угол): выбирайте `Last 15 minutes` или больше, чтобы `rate(...[5m])` имел данные.

## Как использовать метрики

- **Обнаружение узких мест**: высокое время обработки запросов к `ListEventsMonth` или большой размер histogram-бакетов указывает на медленные запросы к хранилищу.
- **Мониторинг ошибок**: рост `calendar_notifications_sent_total{status="error"}` позволяет заметить проблемы с Kafka или сетью.
- **Проверка жизни фоновых сервисов**: `calendar_scheduler_last_success_timestamp_seconds` и `calendar_storer_last_success_timestamp_seconds` показывают, что `scheduler` и `storer` реально работают и обрабатывают данные.
- **Capacity planning**: RPS и длительность запросов помогают решить, когда увеличивать ресурсы или оптимизировать запросы к БД.

## Сбор и визуализация

Сбор метрик выполняется Prometheus, развёрнутым через `docker compose -f deployments/docker-compose.yaml up`. Веб-интерфейс Prometheus (`localhost:9090/graph`) позволяет выполнять ad-hoc PromQL-запросы и строить простые графики. Для более удобной визуализации можно подключить Grafana к тому же Prometheus.
