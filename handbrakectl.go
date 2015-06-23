// -*- compile-command: "go build handbrakectl.go"; -*-

package main

import (
    "fmt"
    "log"
	"os"
	"strings"
    "bufio"

    "github.com/streadway/amqp"
)

const CONFIG_PATH = "/etc/conf.d/rafty.conf"

func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
        panic(fmt.Sprintf("%s: %s", msg, err))
    }
}

func parseConfig(configPath string) map[string]string {
    file, err := os.Open(configPath)
    failOnError(err, fmt.Sprintf("Couldn't open %s", configPath))
    defer file.Close()

    cfg := make(map[string]string)

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if line[0] == '#' { continue }
        if !strings.Contains(line, "=") { continue }
        arr := strings.Split(line, "=")
        cfg[arr[0]] = arr[1]
    }

    err = scanner.Err()
    failOnError(err, fmt.Sprintf("Error scanning config file: ", configPath))
    fmt.Println(cfg)
    return cfg
}

func main() {
    cfg := parseConfig(CONFIG_PATH)
    conn, err := amqp.Dial(cfg["RAFTY_AMQP_URI"])
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
