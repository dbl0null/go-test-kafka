package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().Unix()))

type Bbox struct {
	Left   int `json:"left"`
	Right  int `json:"right"`
	Top    int `json:"top"`
	Bottom int `json:"bottom"`
}

type Object struct {
	Id      int    `json:"id"`
	Class   string `json:"class"`
	Visible bool   `json:"visible"`
	Bbox    Bbox   `json:"bbox"`
}

type Message struct {
	SourceName string    `json:"source_name"`
	Hostname   string    `json:"hostname"`
	Type       string    `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
	Objects    []Object  `json:"objects"`
}

func makeObjects(topic string) []Object {
	size := r.Intn(9) + 1
	o := make([]Object, size)
	for i := 0; i < size; i++ {
		o[i] = Object{
			Id:      r.Intn(4234242),
			Class:   topic,
			Visible: true,
			Bbox:    makeBbox(),
		}
	}
	return o
}

func makeBbox() Bbox {
	return Bbox{
		Left:   r.Intn(999),
		Right:  r.Intn(999),
		Top:    r.Intn(999),
		Bottom: r.Intn(999),
	}
}

func makeTopic() string {
	topics := []string{"human", "fire", "Vehicle", "Unknown"}
	return topics[r.Intn(len(topics))]
}

func createNewMessage() Message {
	topic := makeTopic()
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		hostname = "localhost"
	}

	return Message{
		SourceName: fmt.Sprintf("simulator-%d", r.Intn(10)),
		Hostname:   hostname,
		Type:       topic,
		Timestamp:  time.Now(),
		Objects:    makeObjects(topic),
	}
}

func main() {
	producerURL := fmt.Sprintf("http://%s/api/v1/runner/events", os.Getenv("PRODUCER_URL"))
	//producerURL := "https://webhook.site/5067df51-dfd6-44c0-8447-e06e2d8c307b"
	fmt.Printf("Producer URL: %s\n", producerURL)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-c:
			fmt.Println("Done")
			return
		default:
			message := createNewMessage()
			buf, err := json.Marshal(message)
			//fmt.Println(string(buf))
			if err != nil {
				fmt.Printf("Marshal %s\n", err.Error())
			}
			_, err = http.Post(producerURL, "application/json", bytes.NewBuffer(buf))
			if err != nil {
				fmt.Printf("Post %s\n", err.Error())
			}
			time.Sleep(time.Duration(2) * time.Second)
		}
	}
}
