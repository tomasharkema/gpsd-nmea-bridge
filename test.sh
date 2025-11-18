#!/usr/bin/env sh

set -x
set -e

rm -rf ./.build
go build -o ./.build/gpsd-nmea-bridge

exec systemd-socket-activate -E TEST_SKIP_ENV=1 -E GPSD_ADDR=timepi.local:2947 -l 10110 --now ./.build/gpsd-nmea-bridge