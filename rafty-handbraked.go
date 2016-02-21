// -*- compile-command: "go build rafty-handbraked.go"; -*-
package main

import (
	"errors"
	"flag"
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

// We assume that the main title is the longest one.  This might not be
// true all the time... :(
//
// We use lsdvd to figure out which title is longest.
func getMainTitle(isopath string) (error, string) {
	var err error

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
	return errors.New("Couldn't find Longest title"), ""
}

func getDiscTitle(isopath string) (error, string) {
	cmd := exec.Command("lsdvd", isopath)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error retrieving disc title: %v", err)
		return err, ""
	}
	line := strings.Split(string(output[:]), "\n")[0]
	re := regexp.MustCompile(`^Disc Title: (.*)$`)
	res := re.FindStringSubmatch(line)
	if res != nil {
		return nil, res[1]
	}
	return errors.New("Couldn't find disc title"), ""
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
			parts := strings.Split(string(d.Body[:]), " ")
			flagset := flag.NewFlagSet(parts[0], flag.ContinueOnError)
			isopath := parts[0]
			log.Printf("Got iso from queue: %v", isopath)

			var cmd *exec.Cmd
			var err error

			disctitleopt := flagset.String("disctitle", "", "Movie title")
			titleopt := flagset.String("title", "", "Which title to rip (lsdvd for a listing)")
			flagset.Parse(parts[1:])

			disctitle := *disctitleopt
			title := *titleopt

			if disctitle == "" {
				err, disctitle = getDiscTitle(isopath)
				if err != nil {
					continue
				}
				log.Printf("Found disc title: %s", disctitle);
			} else {
				log.Printf("Using disc title from args: %s", disctitle)
			}
			outdir := path.Join(os.Getenv("RAFTY_OUTPUT_PATH"), disctitle)
			err = os.MkdirAll(outdir, 0777)
			if err != nil {
				log.Printf("Error making dir %s: %v", outdir, err)
				continue
			}

			if title == "" {
				err, title = getMainTitle(isopath)
				if err != nil {
					continue
				}
				log.Printf("Found main title: %s", title);
			} else {
				log.Printf("Using main title from args: %s", title)
			}

			outfile := fmt.Sprintf("%s/%s.mp4", outdir, disctitle)
			cmd = exec.Command("HandBrakeCLI",
				"-i", isopath,
				"-o", outfile,
				"--preset=\"High Profile\"",
				"--title", title)
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
