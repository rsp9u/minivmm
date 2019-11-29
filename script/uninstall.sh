#!/bin/sh -ex
BIN=/usr/local/bin/minivmm
USR=minivmm

sudo systemctl stop minivmm.service
sudo systemctl disable minivmm.service
sudo rm -f /etc/systemd/system/minivmm.service

sudo -u minivmm $BIN -reset-nw
sudo rm -f /etc/sudoers.d/$USR
sudo userdel $USR
sudo rm -f $BIN
