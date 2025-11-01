# ADR-0005: Device Lifecycle & Provisioning

- **Status:** Accepted
- **Date:** 2025-10-30
- **Owners:** Architecture Team
- **Context:** arc-warmhouse / SmartHome platform

## 1) Context

Каждое IoT-устройство должно безопасно зарегистрироваться в системе, получить уникальный идентификатор и безопасные учетные данные (сертификат / токен) для связи через MQTT и REST.  
Платформа должна обеспечивать:
- контроль жизненного цикла устройства (от завода до утилизации);
- безопасную аутентификацию;
- управление версиями прошивок и статусом подключения;
- возможность массового ввода в эксплуатацию (bulk provisioning);
- полную прослеживаемость (audit trail).

---

## 2) Decision

### Архитектурная схема

```text
┌──────────────┐
│ Manufacturer │
│ (factory)    │
└──────┬───────┘
       │ initial key
       ▼
┌─────────────────────┐
│ Device Provisioning │ ←─ REST / API Gateway
│ Service (Go)        │
└──────┬──────────────┘
       │ issues cert / token
       ▼
┌──────────────────────┐
│ Device Twin Service  │
│ (state mgmt)         │
└────────┬─────────────┘
         │
         ▼
┌──────────────────────┐
│ MQTT Broker (NATS)   │
│ telemetry / twin      │
└──────────────────────┘