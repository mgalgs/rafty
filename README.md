Rafty is a hacky DVD backup solution with the potential to scale
horizontally.  You can read more about the design in
[this blog post](http://mgalgs.github.io/2015/04/02/rafty-dvd-backups-using-systemd-docker-rabbitmq-and-go.html).

# Installation

Installation is currently completely manual.  The following steps should
roughly get you up and running:

    $ cd /path/to/rafty
    $ make deps all
    $ sudo make install

You also need to edit two configuration files and a systemd service
(updating them with your username, etc):

    /etc/conf.d/rafty-dd-one.conf
    /etc/conf.d/rafty-handbraked.conf
    /etc/systemd/system/rafty-handbraked.service

You'll also need [`rabbitmq`](https://www.rabbitmq.com/) running on
`localhost`.

Now start the daemon and reload your udev rules:

    $ sudo systemctl start rafty-handbraked.service
    $ sudo udevadm control --reload

And "that's it"!  Now the next time you insert a DVD the Rafty Rube
Goldberg machine should kick into action!

# Manual Invocation

If the `udev` rafting isn't working raftily enough for you, you can invoke
the rip+encode process manually with:

    $ sudo DEVNAME=/dev/sr0 rafty-dd-one.sh

(adjusting `/dev/sr0` as necessary).

I've found that some discs need to be "primed" by playing them for a few
minutes with `mplayer` before attempting to rip them.  You'll need to use
the above command to start the rip process after "priming" the disc in
those cases.

Happy Rafting!
