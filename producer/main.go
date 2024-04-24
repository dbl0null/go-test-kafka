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
		Id      int    `json:"id"`
		Class   string `json:"class"`
		Visible bool   `json:"visible"`
		Bbox    struct {
			Left   int `json:"left"`
			Right  int `json:"right"`
			Top    int `json:"top"`
			Bottom int `json:"bottom"`
		} `json:"bbox"`
	} `json:"objects"`
}

func producerHandler(kafkaWriter *kafka.Writer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var m Message

		err := json.NewDecoder(req.Body).Decode(&m)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		topic := m.Type
		value, err := json.Marshal(m)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		msg := kafka.Message{
			Key:   []byte(fmt.Sprintf("address-%s", req.RemoteAddr)),
			Topic: topic,
			Value: value,
		}

		err = kafkaWriter.WriteMessages(req.Context(), msg)
		if err != nil {
			fmt.Println(err.Error())
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

	http.HandleFunc("/api/v1/runner/events", producerHandler(kafkaWriter))

	fmt.Println("Listening on :8080/api/v1/runner/events...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
