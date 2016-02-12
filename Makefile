all:
	go build rafty-handbrakectl.go
	go build rafty-handbraked.go

deps:
	go get github.com/streadway/amqp
