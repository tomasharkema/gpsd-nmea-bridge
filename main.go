package main

import (
	"log"
	"os"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/pkg/errors"
	command "github.com/tomasharkema/gpsd-nmea-bridge/cmd"
)

func main() {

	app := kingpin.New("systemd-alert", "monitoring around systemd")

	c := &command.BridgeCommand{}
	cmd := app.Command("run", "run").Default()
	cmd.Flag("gpsd-address", "gpsd address").Short('s').Default("localhost:2947").Envar("GPSD_ADDR").StringVar(&c.GpsdAddress)
	cmd.Action(c.Execute)

	// signals := make(chan os.Signal, 1)
	// signal.Notify(signals, os.Kill, os.Interrupt, syscall.SIGUSR2)

	if pcmd, err := app.Parse(os.Args[1:]); err != nil {
		log.Fatalln(pcmd, errors.Wrap(err, "failed to parse commandline"))
	}

}
