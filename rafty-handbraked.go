// -*- compile-command: "go build rafty-handbraked.go"; -*-
package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

// We assume that the main track is the longest one.  This might not be
// true all the time... :(
//
// We use lsdvd to figure out which track is longest.
func getMainTrack(isopath string) (error, string) {
	var err error

	// lsdvd ~/isos/MULAN_USA.iso| grep Longest | cut -d: -f2 | cut -c2
	cmd := exec.Command("lsdvd", isopath)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error retrieving disc listing: %v", err)
		return err, ""
	}
	for _, line := range strings.Split(string(output[:]), "\n") {
		re := regexp.MustCompile(`^Longest track: (\d+)$`)
		res := re.FindStringSubmatch(line)
		if res != nil {
			return nil, res[1]
		}
	}
	return errors.New("Couldn't find Longest track"), ""
}

func main() {
	log.Println("*** rafty-handbraked is starting up...")

	conn, err := amqp.Dial(os.Getenv("RAFTY_AMQP_URI"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"handbraked_ready_isos", // name
		false, // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
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

			err, maintrack := getMainTrack(isopath)
			if err != nil {
				continue
			}
			log.Printf("Found main track: %s", maintrack);

			outfile := fmt.Sprintf("%s/%s.mp4", outdir, title)
			cmd = exec.Command("HandBrakeCLI",
				"-i", isopath,
				"-o", outfile,
				"--preset=\"High Profile\"",
				"--title", maintrack)
			log.Printf("Now ripping with: %s", cmd)
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
