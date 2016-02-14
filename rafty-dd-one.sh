#!/bin/bash

LOGFILE=/var/log/dd-one-from-udev
DDRESCUE_LOGFILE=/var/log/dd-one-from-udev-ddrescue
DDRESCUE_OUTPUT_LOGFILE=/var/log/dd-one-from-udev-ddrescue-output
CONFFILE=/etc/conf.d/rafty-dd-one.conf
PROGGIE=$(basename $0)

log()
{
    echo "$*"
    date +"%F %T $PROGGIE $*" >> $LOGFILE
}

log "hello from $PROGGIE"

# $DEVNAME should come from udev or systemd
log "hello from $0"
[[ -z "$DEVNAME" ]] && { log "DEVNAME env var not set\! bailing..."; exit 1; }
[[ -r $CONFFILE ]] || { log "Couldn't read $CONFFILE"; exit 1; }
source $CONFFILE
[[ -z "$ISOOUTDIR" ]] && { log "Bogus config. Missing ISOOUTDIR."; exit 1; }
[[ -z "$ISOOWNER" ]] && { log "Bogus config. Missing ISOOWNER."; exit 1; }
[[ -z "$ISOGROUP" ]] && { log "Bogus config. Missing ISOGROUP."; exit 1; }
[[ -n "$DISABLE" ]] && { log "rafty-dd-one.sh is disabled. Bailing."; exit 0; }
EJECT=eject
[[ "$NOEJECT" = 1 ]] && EJECT=true
[[ -z "$ISOOUTDIR_REMOTE" ]] && ISOOUTDIR_REMOTE="$ISOOUTDIR"

[[ $EUID -eq 0 ]] || { log "please run as root"; exit 1; }

waitfordisc()
{
    log "waiting for disc and dancing around..."
    sleep 3
    # make sure the disc is unmounted first:
    while read line; do
	sleep 3
        umount $(awk '{print $1}' <<<$line)
    done < <(mount | grep $DEVNAME)

    # bonus mount seems to help get things going... don't ask me...
    tmpmount=$(mktemp -d)
    sleep 3
    mount -t iso9660 -o ro $DEVNAME $tmpmount || { log "Couldn't mount $DEVNAME to $tmpmount... bailing"; errorout; }
    local longest=$(lsdvd /dev/sr0 | grep "Longest track:" | cut -d: -f2 | cut -c2-)
    log "playing a few seconds of the longest title ($longest) to /dev/null to grease the wheels..."
    mplayer -dvd-device $DEVNAME -ao null -vo null -endpos 20 dvd://$longest
    # sometimes umount takes some convincing... I don't know...
    success=no
    for i in $(seq 5); do
        umount $tmpmount && { success=yes; break; }
        log "Couldn't unmount... Sleeping for 3 then trying again..."
        sleep 3
    done
    [[ $success = no ]] && { log "Couldn't umount $DEVNAME"; errorout; }
    rm -r $tmpmount
}

errorout()
{
    $EJECT $DEVNAME
    exit 1
}

[[ -e "$DEVNAME" ]] || { log "Can't find $DEVNAME"; exit 1; }
waitfordisc

isobase=$(blkid $DEVNAME -o value -s LABEL)
if [[ -z "$isobase" ]]; then
    # couldn't get disc title with blkid :( let's try lsdvd.
    log "Couldn't get disc title with blkid... Trying lsdvd..."
    isobase=$(lsdvd $DEVNAME | grep "Disc Title:" | cut -d: -f2 | cut -c2-)
fi
[[ -z "$isobase" ]] && { log "Couldn't get disc title. bailing."; errorout; }
isoname="${isobase}.iso"
IMGNAME=$ISOOUTDIR/$isoname
log "ripping $isoname to $IMGNAME"
mkdir -pv $ISOOUTDIR

success=no
for blocksize in 64k 8k 4k 1024; do
    echo "trying dd with blocksize=$blocksize"
    dd if=$DEVNAME of=$IMGNAME bs=$blocksize || {
        echo "blocksize=$blocksize failed..."
        sleep 2
        continue
    }
    success=yes
    break
done

[[ $success = yes ]] || {
    log "dd failed. making one last attempt with ddrescue..."

    # for all: block size=2048 (which is what the manual suggests for
    # cdroms)

    # need a loop for the `break'...
    while :; do
        log "first trying with direct access"
        ddrescue -d -r 1 -b2048 $DEVNAME $IMGNAME $DDRESCUE_LOGFILE &>$DDRESCUE_OUTPUT_LOGFILE && break
        log "still no dice. Trying with retrim."
        ddrescue -d -R -r 1 -b2048 $DEVNAME $IMGNAME $DDRESCUE_LOGFILE &>$DDRESCUE_OUTPUT_LOGFILE && break
        log "still no... Just try one more time..."
        ddrescue    -r 1 -b2048 $DEVNAME $IMGNAME $DDRESCUE_LOGFILE &>$DDRESCUE_OUTPUT_LOGFILE && break
        log "Nothing worked. Relenting..."
        errorout
    done
}

chown $ISOOWNER:$ISOGROUP $ISOOUTDIR/$isoname
log 'done!'
$EJECT $DEVNAME
cd $(dirname $0)
./rafty-handbrakectl newiso $ISOOUTDIR_REMOTE/$isoname
