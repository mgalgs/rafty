Rafty is a hacky DVD backup solution.  You can read more about the design
in
[this blog post](http://mgalgs.github.io/2015/04/02/rafty-dvd-backups-using-systemd-docker-rabbitmq-and-go.html).

# Installation

Installation is currently completely manual.  The following steps should
roughly get you up and running:

    $ cd /path/to/rafty
    $ go build rafty-handbrakectl.go
    $ go build rafty-handbraked.go
    $ sudo ln -sv $PWD/98-rafty-dd-one-from-udev.rules /etc/udev/rules.d/
    $ sudo ln -sv $PWD/rafty-dd-dvd@.service /etc/systemd/system/
    $ sudo cp -v rafty-handbraked.service /etc/systemd/system/
    $ sudo ln -sv $PWD/rafty-dd-one.sh /usr/bin/
    $ sudo ln -sv $PWD/rafty-handbraked /usr/bin/
    $ sudo ln -sv $PWD/rafty-handbrakectl /usr/bin/
    $ sudo cp -v rafty-dd-one.conf /etc/conf.d/
    $ sudo cp -v rafty-handbraked.conf /etc/conf.d/
    $ sudo vim /etc/conf.d/rafty-dd-one.conf
    $ sudo vim /etc/conf.d/rafty-handbraked.conf
    $ sudo vim /etc/systemd/system/rafty-handbraked.service

You'll also need [`rabbitmq`](https://www.rabbitmq.com/) running.  One easy
way to get `rabbitmq` is with the
[`dockerfile/rabbitmq`](http://dockerfile.github.io/#/rabbitmq) `docker`
image and the provided `systemd` service file to start it:

    $ sudo docker pull dockerfile/rabbitmq
    $ sudo ln -sv $PWD/rafty-docker-rabbitmq.service /etc/systemd/system/

Everything should now be in place.  Now you can start all the moving parts:

    $ sudo systemctl start rafty-docker-rabbitmq.service
    $ sudo systemctl start rafty-handbraked.service
    $ sudo udevadm control --reload

And "that's it"!  Now the next time you insert a DVD the Rafty Rube
Goldberg machine should kick into action!

Happy Rafting!
