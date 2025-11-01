# ADR-0001: API Styles (REST + Events + gRPC)

- **Status:** Accepted
- **Date:** 2025-10-30
- **Owners:** Architecture Team
- **Context:** arc-warmhouse / SmartHome platform

## 1) Context

Платформа разделена на микросервисы (Device Registry & Twin, Telemetry, Heating, Identity, Automation).
Требуется согласованный подход к взаимодействию:

- внешние клиенты (Web/Mobile),
- сервис↔сервис синхронные запросы,
- потоковая телеметрия с устройств,
- надёжная доставка команд/событий,
- оффлайн-устойчивость и идемпотентность.

У нас уже есть спецификации:
- **REST (OpenAPI 3.1):** `schemas/openapi-smarthome.yml` → `docs/images/openapi.html`
- **Events (AsyncAPI 2.6):** `schemas/asyncapi-smarthome.yml` → `docs/images/asyncapi/index.html`

## 2) Decision

Используем **полиглотный стиль API**:

1. **REST (HTTP/JSON, OpenAPI 3.1)** — для синхронных операций и чтений
    - Внешние клиенты и «медленные» S2S: CRUD устройств/домов/пользователей, чтение твина, агрегации телеметрии, постановка команд «по кнопке».
    - Контракты версионируем по URL (`/api/v1/...`) и схемам (semantic versioning).
    - Конфликты/идемпотентность — `ETag/If-Match` (версионирование Twin), `202 Accepted` для команд с асинхронной конвергенцией.

2. **Событийная шина (MQTT/NATS, AsyncAPI 2.6)** — для асинхронного обмена
    - Инжест телеметрии, публикация `desired/delta` твина, доменные события (`device.online`, `command.ack`, `rule.fired`).
    - Топики версионируем префиксом: `v1/devices/{id}/twin/...`, `v1/telemetry/{deviceId}`.
    - Идемпотентность по `messageId`/версии твина; QoS1/2 и retained-сообщения для устойчивости.

3. **gRPC (Protobuf, опционально для внутреннего S2S)** — для «горячих» путей
    - Низкая латентность между сервисами (напр., API Gateway ↔ Telemetry/Heating).
    - .proto в `apis/grpc/` (введём при необходимости).

## 3) Alternatives Considered

- **Только REST:** просто, но плохо для высокочастотной телеметрии и оффлайна; вырастет связность и latency.
- **Только события (event-first):** сложно выстроить UX-сценарии с немедленной обратной связью; повышается сложность запроса состояния.
- **GraphQL:** удобные выборки, но не решает ingestion/оффлайн, требует дополнительной инфраструктуры и дисциплины схем.
- **gRPC-only:** отличная производительность, но сложнее для внешних клиентов и не покрывает IoT-телеметрию/ретенции.

## 4) Consequences

**Плюсы**
- Масштабируемость: ingestion и командная доставка не блокируют REST.
- Устойчивость к оффлайну: retained сообщения, QoS, twin-версионирование.
- Чёткие контракты: OpenAPI/AsyncAPI/Protobuf, генерация SDK/стабов.
- Наблюдаемость: трейсинг REST/gRPC, метрики брокера, события конвергенции twin.

**Минусы**
- Сложнее DevEx: 2–3 технологии вместо одной.
- Нужно дисциплинированное версионирование и каталог схем.
- Единая авторизация и аудит для разных транспортов.

## 5) Security

- **REST:** OIDC/JWT (Keycloak), mTLS между сервисами, rate limit на API Gateway.
- **MQTT/NATS:** mTLS для устройств, пер-устройство учётные данные, ACL по топикам, LWT (offline).
- **gRPC:** mTLS, authz на уровне сервисов.
- Корреляция запросов/сообщений через `X-Correlation-Id` / message headers.

## 6) Versioning & Compatibility

- REST: `/api/v{n}` и семантические версии схем; deprecation headers.
- Events: префикс `v{n}/` в каналах; поле `schemaVersion` в payload.
- gRPC: пакетная версия в namespace (`smarthome.v1`), additive-only изменения.

## 7) Observability

- OpenTelemetry для REST/gRPC; трейс-ид в брокерные сообщения.
- Метрики: `twin_convergence_time_seconds`, `telemetry_ingest_rps`, `command_retry_total`, `broker_qos1_redelivery_total`.
- Логи с `correlationId`, `deviceId`, `twinVersion`.

## 8) Compliance / Docs

- Артефакты генерируем `make docs`:
    - OpenAPI → `docs/images/openapi.html` (Redocly)
    - AsyncAPI → `docs/images/asyncapi/index.html`
    - C4/ERD PNG → `docs/images/`
- Храним исходники в:
    - `docs/c4/`, `docs/erd/`
    - `schemas/openapi-smarthome.yml`, `schemas/asyncapi-smarthome.yml`

## 9) Adoption Plan

1. Зафиксировать контракты Device/Twin/Telemetry/Heating (OpenAPI/AsyncAPI).
2. Включить в CI проверку схем (`redocly openapi lint`, `asyncapi validate`).
3. Поднять брокер, включить ACL и LWT; перевести ingestion на события.
4. В API Gateway включить JWT, rate limit, трейсинг.
5. Мигрировать функции монолита в сервисы, сохраняя совместимость контрактов.

---