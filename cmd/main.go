package command

import (
	"bufio"

	"fmt"
	"log"
	"net"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/coreos/go-systemd/v22/activation"
	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/pkg/errors"

	"github.com/godbus/dbus/v5"
	"github.com/holoplot/go-avahi"
)

type BridgeCommand struct {
	GpsdAddress string
}

func ready() {
	if _, testSet := os.LookupEnv("TEST_SKIP_ENV"); testSet {
		log.Println("skip notify")
		return
	}
	ready, err := daemon.SdNotify(true, daemon.SdNotifyReady)
	if err != nil {
		log.Fatalln("sd-notify failed", errors.Wrap(err, "sd-notify failed"))
	}

	if !ready {
		log.Fatalln("not ready!")
	}
	log.Println("Reported ready...")
}

func advertiseAvahi() {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatalf("Cannot get system bus: %v", errors.Wrap(err, ""))
	}

	a, err := avahi.ServerNew(conn)
	if err != nil {
		log.Fatalf("Avahi new failed: %v", errors.Wrap(err, ""))
	}

	eg, err := a.EntryGroupNew()
	if err != nil {
		log.Fatalf("EntryGroupNew() failed: %v", errors.Wrap(err, ""))
	}

	hostname, err := a.GetHostName()
	if err != nil {
		log.Fatalf("GetHostName() failed: %v", errors.Wrap(err, ""))
	}

	fqdn, err := a.GetHostNameFqdn()
	if err != nil {
		log.Fatalf("GetHostNameFqdn() failed: %v", errors.Wrap(err, ""))
	}

	err = eg.AddService(
		avahi.InterfaceUnspec, avahi.ProtoUnspec, 0, hostname,
		"_nmea-0183._tcp", "local", fqdn, 10110, nil,
	)
	if err != nil {
		log.Fatalf("AddService() failed: %v", errors.Wrap(err, ""))
	}

	err = eg.Commit()
	if err != nil {
		log.Fatalf("Commit() failed: %v", errors.Wrap(err, ""))
	}
}

func (c *BridgeCommand) Execute(p *kingpin.ParseContext) error {
	listeners, err := activation.Listeners()
	if err != nil {
		log.Fatalln(errors.Wrap(err, ""))
	}

	if len(listeners) != 1 {
		log.Fatalln("Unexpected number of socket activation fds")
	}
	l := listeners[0]

	advertiseAvahi()

	ready()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("error", errors.Wrap(err, ""))
		}
		go handleConnection(conn, c.GpsdAddress)
	}
}

func handleConnection(conn net.Conn, address string) {
	// Close the connection when we're done
	defer conn.Close()

	gpsdconn, err := net.Dial("tcp", address)
	if err != nil {
		log.Println("connection failed", errors.Wrap(err, "connection failed"))
		return
	}
	defer gpsdconn.Close()

	gpsdconnReader := bufio.NewScanner(gpsdconn)

	gpsdconnReader.Scan()
	_, err = fmt.Fprintln(gpsdconn, "?WATCH={\"enable\":true,\"nmea\":true}")
	if err != nil {
		log.Println("error", errors.Wrap(err, "watch write error"))
		return
	}
	gpsdconnReader.Scan()
	gpsdconnReader.Scan()

	for {
		if gpsdconnReader.Scan() {
			line := gpsdconnReader.Text()
			fmt.Println("Got line:", line)
			_, err = fmt.Fprintln(conn, line)
			if err != nil {
				log.Println("error", errors.Wrap(err, ""))
				return
			}
		} else {
			err = gpsdconnReader.Err()
			if err != nil {
				log.Println("error", errors.Wrap(err, ""))
				return
			}
		}
	}
}
