INSTALL_PREFIX = /usr/bin
CONF_PREFIX = /etc/conf.d

all: rafty-handbrakectl rafty-handbraked

rafty-handbrakectl: rafty-handbrakectl.go
	go build rafty-handbrakectl.go

rafty-handbraked: rafty-handbraked.go
	go build rafty-handbraked.go

install:
	install -m 644 98-rafty-dd-one-from-udev.rules /etc/udev/rules.d/
	install -m 644 rafty-dd-dvd@.service /etc/systemd/system/
	test -e /etc/systemd/system/rafty-handbraked.service || install -m 644 rafty-handbraked.service /etc/systemd/system/
	install rafty-dd-one.sh $(INSTALL_PREFIX)/
	install rafty-handbraked $(INSTALL_PREFIX)/
	install rafty-handbrakectl $(INSTALL_PREFIX)/
	test -e /etc/conf.d/rafty-dd-one.conf || install -m 644 rafty-dd-one.conf /etc/conf.d/
	test -e /etc/conf.d/rafty-handbraked.conf || install -m 644 rafty-handbraked.conf /etc/conf.d/

deps:
	go get github.com/streadway/amqp
