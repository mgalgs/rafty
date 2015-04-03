// -*- compile-command: "go build handbrakectl.go"; -*-

package main

import (
    "fmt"
    "log"
	"os"
	"strings"

    "github.com/streadway/amqp"
)

const AMQP_URI = "amqp://guest:guest@localhost:5672/"

func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
        panic(fmt.Sprintf("%s: %s", msg, err))
    }
}

func main() {
    conn, err := amqp.Dial(AMQP_URI)
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "handbraked_ready_isos", // name
        false,   // durable
        false,   // delete when usused
        false,   // exclusive
        false,   // no-wait
        nil,     // arguments
    )
    failOnError(err, "Failed to declare a queue")

	var body string
	cmd := os.Args[1]
	switch cmd {
	case "newiso":
		body = strings.Join(os.Args[2:], " ")
	default:
		log.Fatalf("Unknown command: %v", cmd)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
}
