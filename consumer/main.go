package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

func getKafkaReader(kafkaURL, topic, groupID string) *kafka.Reader {
	brokers := strings.Split(kafkaURL, ",")
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		//GroupTopics: []string{"fire", "human", "vehicle", "unknown"},
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}

func readKafkaTopic(ctx context.Context, kafkaURL, groupID, topic string) {
	reader := getKafkaReader(kafkaURL, topic, groupID)

	defer reader.Close()

	fmt.Printf("Consuming topic %s..\n", topic)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Done topic %s\n", topic)
			return
		default:
			message, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Fatalln(err)
			}
			// TODO: log to file
			fmt.Printf("Message in topic %s: partition:%v offset:%v	%s = %s\n", message.Topic, message.Partition, message.Offset, string(message.Key), string(message.Value))
		}
	}
}

func main() {
	kafkaURL := os.Getenv("KAFKA_URL")
	groupID := os.Getenv("KAFKA_GROUP_ID")

	ctx, cancel := context.WithCancel(context.Background())

	go readKafkaTopic(ctx, kafkaURL, groupID, "fire")
	go readKafkaTopic(ctx, kafkaURL, groupID, "human")
	go readKafkaTopic(ctx, kafkaURL, groupID, "Vehicle")
	go readKafkaTopic(ctx, kafkaURL, groupID, "Unknown")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-c:
			fmt.Println("Stopping...")
			cancel()
			time.Sleep(time.Duration(10) * time.Second)
			return
		default:
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}
