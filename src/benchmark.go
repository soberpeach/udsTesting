package main

import (
	"flag"
	"log"
	"net"
	"os"
)

var UnixDomain = flag.Bool("unixdomain", false, "Use Unix domain sockets")
var MsgSize = flag.Int("msgsize", 128, "Message size in each ping")
var NumPings = flag.Int("n", 50000, "Number of pings to measure")

var UnixAddress = "/tmp/test.sock"

func domainAndAddress() (string, string) {
	if *UnixDomain {
		return "unix", UnixAddress
	}
}

func server() {
	if *UnixDomain {
		if err := os.RemoveAll(UnixAddress); err != nil {
			panic(err)
		}
	}

	domain, address := domainAndAddress()

	l, err := net.Listen(domain, address)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, *MsgSize)
	for n := 0; n < *NumPings; n++ {
		nread, err := conn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if nread != *MsgSize {
			log.Fatal("bad nread = %d", nread)
		}
	}
}
