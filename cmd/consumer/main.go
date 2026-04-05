package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
)

type CDCRecord struct {
	Schema  json.RawMessage `json:"schema"`
	Payload json.RawMessage `json:"payload"`
}

type Payload struct {
	Before interface{}     `json:"before"`
	After  json.RawMessage `json:"after"`
	Source Source          `json:"source"`
	Op     string          `json:"op"`
}

type Source struct {
	Table string `json:"table"`
}

func main() {
	brokers := []string{"localhost:9092", "localhost:9094", "localhost:9096"}
	topics := []string{"debezium.public.users", "debezium.public.orders"}

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, "cdc-consumer-group", config)
	if err != nil {
		log.Fatalf("Error creating consumer group: %v", err)
	}
	defer consumerGroup.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, topics, &consumerGroupHandler{}); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-sigchan
	fmt.Println("\nShutting down consumer...")
	cancel()
}

type consumerGroupHandler struct{}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			if len(message.Value) == 0 {
				session.MarkMessage(message, "")
				continue
			}

			var record CDCRecord
			if err := json.Unmarshal(message.Value, &record); err != nil {
				log.Printf("Error unmarshaling record: %v, value: %s", err, string(message.Value))
				session.MarkMessage(message, "")
				continue
			}

			if len(record.Payload) == 0 {
				fmt.Printf("\n=== CDC Event ===\n")
				fmt.Printf("Topic: %s\n", message.Topic)
				fmt.Printf("Operation: DELETE (no payload)\n")
				session.MarkMessage(message, "")
				continue
			}

			var payload Payload
			if err := json.Unmarshal(record.Payload, &payload); err != nil {
				log.Printf("Error unmarshaling payload: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			fmt.Printf("\n=== CDC Event ===\n")
			fmt.Printf("Topic: %s\n", message.Topic)
			fmt.Printf("Table: %s\n", payload.Source.Table)
			fmt.Printf("Operation: %s\n", payload.Op)
			fmt.Printf("Data: %s\n", string(record.Payload))

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
