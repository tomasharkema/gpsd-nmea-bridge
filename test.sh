#!/usr/bin/env sh

set -x

rm -rf ./.build
go build -o ./.build/gpsd-nmea-bridge

systemd-socket-activate -E TEST_SKIP_ENV=1 -E GPSD_ADDR=timepi.local:2947 -l 10110 ./.build/gpsd-nmea-bridge