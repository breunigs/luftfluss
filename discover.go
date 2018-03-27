package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/godbus/dbus"
)

const (
	AIRPLAY_TYPE   = "_airplay._tcp"
	AIRPLAY_DOMAIN = "local"

	AVAHI_IF_UNSPEC    = int32(-1)
	AVAHI_PROTO_UNSPEC = int32(-1)
)

type txtRecord struct {
	key   string
	value string
}

type avahiItem struct {
	iface    int32
	protocol int32
	name     string
	atype    string
	domain   string

	host      string
	aprotocol int32
	address   string
	port      uint16
	txt       []txtRecord
	flags     uint32
}

func listenDiscoveries(conn *dbus.Conn) {
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal'")
	cs := make(chan *dbus.Signal, 10)
	conn.Signal(cs)

	for s := range cs {
		switch s.Name {
		case "org.freedesktop.Avahi.ServiceBrowser.ItemNew":
			item := avahiItem{
				iface:    s.Body[0].(int32),
				protocol: s.Body[1].(int32),
				name:     s.Body[2].(string),
				atype:    s.Body[3].(string),
				domain:   s.Body[4].(string),
				flags:    s.Body[5].(uint32),
			}
			if item.atype != AIRPLAY_TYPE || item.domain != AIRPLAY_DOMAIN {
				continue
			}
			avahi := conn.Object("org.freedesktop.Avahi", "/")
			call := avahi.Call("org.freedesktop.Avahi.Server.ServiceResolverNew", 0, item.iface, item.protocol, item.name, AIRPLAY_TYPE, AIRPLAY_DOMAIN, int32(-1), uint32(0))
			if call.Err != nil {
				panic(call.Err)
			}

		case "org.freedesktop.Avahi.ServiceResolver.Found":
			txt := s.Body[9].([][]uint8)
			item := avahiItem{
				iface:     s.Body[0].(int32),
				protocol:  s.Body[1].(int32),
				name:      s.Body[2].(string),
				atype:     s.Body[3].(string),
				domain:    s.Body[4].(string),
				host:      s.Body[5].(string),
				aprotocol: s.Body[6].(int32),
				address:   s.Body[7].(string),
				port:      s.Body[8].(uint16),
				txt:       make([]txtRecord, len(txt)),
				flags:     s.Body[10].(uint32),
			}
			for i := 0; i < len(txt); i++ {
				entry := strings.SplitN(string(txt[i]), "=", 2)
				item.txt[i] = txtRecord{key: entry[0], value: entry[1]}
			}
			fmt.Printf("%s:%d  %s\n", item.address, item.port, item.name)
			fmt.Printf("%+v\n", item)
		default:
			// fmt.Printf("Unsupported message: %s\n", s.Name)
		}
	}
}

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}
	avahi := conn.Object("org.freedesktop.Avahi", "/")

	go listenDiscoveries(conn)

	call := avahi.Call("org.freedesktop.Avahi.Server.ServiceBrowserNew", 0, AVAHI_IF_UNSPEC, AVAHI_PROTO_UNSPEC, AIRPLAY_TYPE, AIRPLAY_DOMAIN, uint32(0))
	if call.Err != nil {
		panic(call.Err)
	}
	time.Sleep(60 * time.Second)
}
