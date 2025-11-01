# ADR-0002: Twin Versioning & Idempotency

- **Status:** Accepted  
- **Date:** 2025-10-30  
- **Owners:** Architecture Team  
- **Context:** arc-warmhouse / SmartHome platform

## 1) Context

Устройства и облако обмениваются состоянием через механизм **Device Twin** (двойник устройства).  
Twin состоит из двух частей:
- **desired** — желаемое состояние, формируется облаком (через API/UI/правила);
- **reported** — фактическое состояние, публикуется устройством.

Требуется обеспечить:
- согласованность между desired/reported;
- идемпотентность команд и обновлений;
- корректную обработку повторных доставок (QoS1/2, retry);
- безопасное версионирование twin и событий.

## 2) Decision

Используется модель **оптимистичного версионирования** и **merge-patch идемпотентности**:

| Компонент | Версия | Передача | Принцип |
|------------|--------|-----------|----------|
| REST API | ETag / If-Match | HTTP | оптимистичный merge-patch |
| MQTT/NATS | `version` в payload | JSON | compare-and-apply |
| Storage | `twin_version` BIGINT | DB | автоинкремент при изменении desired/reported |

### REST API

- `GET /devices/{id}/twin` возвращает `ETag: "v123"`.
- `PATCH /devices/{id}/twin` принимает тело-патч и заголовок `If-Match: "v123"`.
- Если версия изменилась → `409 Conflict`, клиент перечитывает twin.
- Все операции **идемпотентны** — повтор того же PATCH c тем же `If-Match` не меняет состояние.

### MQTT / AsyncAPI

- Каналы:
  - `devices/{deviceId}/twin/desired`
  - `devices/{deviceId}/twin/reported`
  - `devices/{deviceId}/twin/delta`
- Каждое сообщение содержит:
  ```json
  {
    "version": 124,
    "patch": { "heating": { "setpoint": 22.0 } },
    "messageId": "uuid",
    "timestamp": "2025-10-30T13:15:00Z"
  }