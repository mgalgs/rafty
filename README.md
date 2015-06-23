Rafty is a hacky DVD backup solution.  You can read more about the design
in
[this blog post](http://mgalgs.github.io/2015/04/02/rafty-dvd-backups-using-systemd-docker-rabbitmq-and-go.html).

# Installation

Installation is currently completely manual.  The following steps should
roughly get you up and running:

    $ cd /path/to/rafty
    $ go build handbrakectl.go
    $ go build handbraked.go
    $ sudo ln -sv $PWD/98-dd-one-from-udev.rules /etc/udev/rules.d/98-dd-one-from-udev.rules
    $ sudo ln -sv $PWD/dd-dvd@.service /etc/systemd/system/
    $ sudo ln -sv $PWD/handbraked.service /etc/systemd/system/
    $ sudo ln -sv $PWD/dd-one.sh /usr/bin/
    $ sudo ln -sv $PWD/handbraked /usr/bin/
    $ sudo ln -sv $PWD/handbrakectl /usr/bin/
    $ sudo cp -v rafty.conf /etc/conf.d/
    $ sudo vim /etc/conf.d/rafty.conf

You'll also need [`rabbitmq`](https://www.rabbitmq.com/) running.  One easy
way to get `rabbitmq` is with `docker`:

    $ sudo docker pull dockerfile/rabbitmq
    $ sudo docker run -d -p 5672:5672 -p 15672:15672 -v /path/to/rabbitmq.persist/log:/data/log -v /path/to/rabbitmq.persist/mnesia:/data/mnesia dockerfile/rabbitmq

More info on running `rabbitmq` with the `dockerfile/rabbitmq` `docker`
image is available [here](http://dockerfile.github.io/#/rabbitmq).  You
might also want to start the `rabbitmq` container on boot.  You can do so
with the included `docker-rabbitmq.service` `systemd` service file:

    $ vim handbraked.service # see comments in this file
    $ sudo ln -sv $PWD/handbraked.service /etc/systemd/system/
