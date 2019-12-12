#!/bin/sh -ex
BIN=/usr/local/bin/minivmm
USR=minivmm

if [ $(id -u) -eq 0 ]; then
  sudo=
else
  sudo=sudo
fi

default_dev=$(ip route | grep "^default" | sed -e 's|.*\(dev.*\)|\1|' | awk '{print $2}')
$sudo iptables -t nat -D POSTROUTING -o $default_dev -s 192.168.200.0/24 -j MASQUERADE

$sudo systemctl stop minivmm.service
$sudo systemctl disable minivmm.service
$sudo rm -f /etc/systemd/system/minivmm.service

$sudo $BIN -reset-nw
$sudo rm -f /etc/sudoers.d/$USR
$sudo userdel $USR
$sudo rm -f $BIN
