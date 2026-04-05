# CDC Consumer

Потребляет CDC события из PostgreSQL через Debezium Connector.

## Запуск

### 1. Запуск инфраструктуры

```bash
docker-compose up -d
```

### 2. Регистрация Debezium Connector

```bash
curl -i -X POST -H "Accept:application/json" \
  -H "Content-Type:application/json" \
  http://localhost:8083/connectors/ \
  -d @connector-config.json
```

### 3. Проверка статуса коннектора

```bash
curl http://localhost:8083/connectors/postgres-source-connector/status
```

### 4. Запуск Go consumer

```bash
go run cmd/consumer/main.go
```

### 5. Генерация изменений в БД

В отдельном терминале:

```bash
docker exec -it postgres psql -U postgres -d mydb -c "INSERT INTO users (name, email) VALUES ('Test', 'test@test.com');"
docker exec -it postgres psql -U postgres -d mydb -c "UPDATE users SET email='new@test.com' WHERE name='Test';"
docker exec -it postgres psql -U postgres -d mydb -c "DELETE FROM users WHERE name='Test';"
```

## 6. Проверка топиков

```bash
docker exec -it kafka1 /opt/kafka/bin/kafka-topics.sh --list --bootstrap-server localhost:9092
```

```bash
docker exec -it kafka1 /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic debezium.dbserver1.users \
  --from-beginning
```

## 7. Остановка

```bash
docker-compose down -v
```
