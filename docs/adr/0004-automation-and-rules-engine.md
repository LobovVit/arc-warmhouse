# ADR-0004: Automation & Rules Engine Integration

- **Status:** Accepted
- **Date:** 2025-10-30
- **Owners:** Architecture Team
- **Context:** arc-warmhouse / SmartHome platform

## 1) Context

Пользователи хотят автоматизировать управление устройствами по принципу:
> «Если температура упала ниже 20°C — включи отопление»  
> «Если никто не дома — выключи все устройства»  
> «Каждое утро в 7:00 — установить setpoint 22°C»

Для этого требуется компонент **Automation & Rules Engine**,  
который реагирует на события из Telemetry, Twin и Scheduler,  
выполняет условия и формирует команды/desired-патчи в Device Twin.

---

## 2) Decision

### Общая архитектура

```text
┌────────────────────┐
│  MQTT / NATS Bus   │
│  (telemetry.*      │
│   twin.changed.*)  │
└──────────┬─────────┘
           │
           ▼
┌────────────────────┐
│ Automation Service │
│ (Rules Engine)     │
└──────────┬─────────┘
           │
           ▼
┌───────────────────────────┐
│ Device Twin Service       │
│ (applies desired changes) │
└───────────────────────────┘