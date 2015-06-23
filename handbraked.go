// -*- compile-command: "go build handbraked.go"; -*-
package main

import (
    "fmt"
    "log"
	"os"
	"os/exec"
	"bytes"
	"strings"
	"path"

    "github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
        panic(fmt.Sprintf("%s: %s", msg, err))
    }
}

func main() {
    conn, err := amqp.Dial(os.Getenv("RAFTY_AMQP_URI"))
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

    msgs, err := ch.Consume(
        q.Name, // queue
        "",     // consumer
        true,   // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // args
    )
    failOnError(err, "Failed to register a consumer")

    forever := make(chan bool)

    go func() {
        for d := range msgs {
			isopath := string(d.Body[:])
            log.Printf("Got iso from queue: %v", isopath)

			var cmd *exec.Cmd
			var err error

			cmd = exec.Command("blkid",
				"-o", "value",
				"-s", "LABEL",
				isopath)
			var out bytes.Buffer
			cmd.Stdout = &out
			err = cmd.Run()
			if err != nil {
				log.Printf("Error retrieving disc title: %v", err)
				continue
			}
			title := strings.Trim(out.String(), "\n")
			outdir := path.Join(os.Getenv("RAFTY_OUTPUT_PATH"), title)
			err = os.MkdirAll(outdir, 0777)
			if err != nil {
				log.Printf("Error making dir %s: %v", outdir, err)
				continue
			}

			cmd = exec.Command("HandBrakeCLI",
				"-i", isopath,
				"-o", fmt.Sprintf("%s/%s.mp4", outdir, title),
				"--preset=\"High Profile\"")
			log.Printf("Now ripping...")
			err = cmd.Run()
			if err != nil {
				log.Printf("Got an error: %v", err)
			}
			log.Printf("Handbrake rip finished!  Waiting for more work...")
        }
    }()

    log.Printf(" [*] Waiting for isos. To exit press CTRL+C")
    <-forever
}
