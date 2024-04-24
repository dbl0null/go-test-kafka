package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type Message struct {
	SourceName string    `json:"source_name"`
	Hostname   string    `json:"hostname"`
	Type       string    `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
	Objects    []struct {
		Id      string `json:"id"`
		Class   string `json:"class"`
		Visible bool   `json:"visible"`
		Bbox    struct {
			Left   string `json:"left"`
			Right  string `json:"right"`
			Top    string `json:"top"`
			Bottom string `json:"bottom"`
		} `json:"bbox"`
	} `json:"objects"`
}

func producerHandler(kafkaWriter *kafka.Writer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var m Message

		err := json.NewDecoder(req.Body).Decode(&m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		topic := m.Type
		value, err := json.Marshal(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		msg := kafka.Message{
			Key:   []byte(fmt.Sprintf("address-%s", req.RemoteAddr)),
			Topic: topic,
			Value: value,
		}

		err = kafkaWriter.WriteMessages(req.Context(), msg)
		if err != nil {
			w.Write([]byte(err.Error()))
			log.Fatalln(err)
		}
	}
}

func getKafkaWriter(kafkaURL string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(kafkaURL),
		Balancer: &kafka.LeastBytes{},
	}
}

func main() {
	kafkaURL := os.Getenv("KAFKA_URL")
	kafkaWriter := getKafkaWriter(kafkaURL)

	defer kafkaWriter.Close()

	http.HandleFunc(":8080/api/v1/runner/events", producerHandler(kafkaWriter))

	fmt.Println("Listening on :8080/api/v1/runner/events...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
