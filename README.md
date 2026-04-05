# CDC with Debezium, PostgreSQL and Kafka Connect

Настройка Debezium Connector для передачи данных из PostgreSQL в Apache Kafka с использованием CDC (Change Data Capture).

## Компоненты

| Компонент | Описание | Порт |
|-----------|----------|------|
| PostgreSQL | База данных с таблицами users и orders | 5432 |
| Kafka Broker (x3) | Kafka кластер для обработки сообщений | 9092, 9094, 9096 |
| Kafka Connect | Платформа для интеграции с Debezium | 8083 |
| JMX Exporter | Экспорт JMX метрик в Prometheus | 9404 |
| Prometheus | Сбор метрик | 9090 |
| Grafana | Визуализация метрик | 3000 |

## Архитектура

```
┌─────────────┐    CDC     ┌──────────────┐         ┌─────────────┐
│ PostgreSQL  │ ─────────► │Kafka Connect │ ──────► │   Kafka     │
│  (source)   │            │  (Debezium)  │         │  (topics)   │
└─────────────┘            └──────────────┘         └─────────────┘
                                       │
                              ┌────────┴────────┐
                              │   JMX Exporter  │
                              │  (port 9999)    │
                              └────────┬────────┘
                                       ▼
                               ┌─────────────┐
                               │ Prometheus  │
                               └─────────────┘
                                      │
                                      ▼
                               ┌─────────────┐
                               │   Grafana   │
                               └─────────────┘
```

## Запуск

### 1. Сборка JMX Exporter

```bash
docker compose build jmx-exporter
```

### 2. Запуск инфраструктуры

```bash
docker compose up -d
```

Дождитесь полного запуска всех сервисов (~30 секунд).

### 3. Регистрация Debezium Connector

```bash
curl -X POST -H "Content-Type:application/json" \
  http://localhost:8083/connectors/ \
  -d @connector-config.json
```

### 4. Проверка статуса коннектора

```bash
curl http://localhost:8083/connectors/postgres-source-connector/status
```

Ожидаемый ответ:
```json
{
  "connector": {
    "state": "RUNNING"
  },
  "tasks": [
    {
      "state": "RUNNING"
    }
  ]
}
```

### 5. Проверка созданных топиков

```bash
docker exec -it kafka1 /opt/kafka/bin/kafka-topics.sh \
  --list --bootstrap-server localhost:9092
```

Должны появиться топики:
- `debezium.public.users`
- `debezium.public.orders`


### 6. Запуск Go Consumer

```bash
go run cmd/consumer/main.go
```

## Мониторинг

### Prometheus

Откройте http://localhost:9090

Targets должны быть "up":
- prometheus
- kafka-connect (jmx-exporter)

### Grafana

Откройте http://localhost:3000

Логин: `admin`
Пароль: `admin`

Dashboard: **Kafka Connect CDC Monitoring**

Доступные графики:
- Active Records in Source
- Producer Outgoing Byte Rate
- Producer Request Rate
- Producer Request Size Avg
- Producer Error Rate

## Настройки Debezium Connector

```json
{
  "name": "postgres-source-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "postgres",
    "database.port": "5432",
    "database.user": "postgres",
    "database.password": "postgres",
    "database.dbname": "mydb",
    "table.include.list": "public.users,public.orders",
    "topic.prefix": "debezium",
    "plugin.name": "pgoutput",
    "snapshot.mode": "initial"
  }
}
```

## JMX Exporter

JMX Exporter собирает метрики с Kafka Connect через JMX порт 9999.

Конфигурация: `jmx-exporter/config.yml`

Метрики собираются по всем паттернам JMX MBeans Kafka Connect.

## Тестирование CDC

### Вставка данных

```bash
docker exec -it postgres psql -U postgres -d mydb -c \
  "INSERT INTO users (name, email) VALUES ('Test User', 'test@example.com');"
```

### Обновление данных

```bash
docker exec -it postgres psql -U postgres -d mydb -c \
  "UPDATE users SET email='new@example.com' WHERE name='Test User';"
```

### Удаление данных

```bash
docker exec -it postgres psql -U postgres -d mydb -c \
  "DELETE FROM users WHERE name='Test User';"
```

### Просмотр сообщений в топике

```bash
docker exec -it kafka1 /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic debezium.public.users \
  --from-beginning
```

## Структура проекта

```
.
├── docker-compose.yml       # Конфигурация всех сервисов
├── init.sql                 # SQL для создания таблиц
├── connector-config.json    # Конфигурация Debezium
├── prometheus.yml           # Конфигурация Prometheus
├── jmx-exporter/
│   ├── Dockerfile           # Dockerfile для JMX Exporter
│   └── config.yml           # Конфигурация JMX Exporter
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/     # Автонастройка источников
│   │   └── dashboards/      # Автонастройка dashboard
│   └── dashboards/          # JSON dashboards
└── README.md
```

## Остановка

```bash
docker compose down -v
```
