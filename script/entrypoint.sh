#!/bin/bash
cleanup() {
  /usr/bin/minivmm -reset-nw
}

trap cleanup EXIT

minivmm -init-nw
minivmm
