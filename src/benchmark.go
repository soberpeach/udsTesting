package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var UnixDomain = flag.Bool("unixdomain", false, "Use Unix domain sockets")
var MsgSize = flag.Int("msgsize", 128, "Message size in each ping")
var NumPings = flag.Int("n", 1_000_000, "Number of pings to measure")

var UnixAddress = "/tmp/test.sock"

func domainAndAddress() (string, string) {
	return "unix", UnixAddress
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
			log.Fatalf("bad nread = %d", nread)
		}
		nwrite, err := conn.Write(buf)
		if err != nil {
			log.Fatal(err)
		}
		if nwrite != *MsgSize {
			log.Fatalf("bad nwrite = %d", nwrite)
		}
	}

	time.Sleep(50 * time.Millisecond)
}

func main() {
	flag.Parse()

	go server()
	time.Sleep(50 * time.Millisecond)

	// This is the client code in the main goroutine.
	domain, address := domainAndAddress()
	conn, err := net.Dial(domain, address)
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, *MsgSize)
	t1 := time.Now()
	for n := 0; n < *NumPings; n++ {
		nwrite, err := conn.Write(buf)
		if err != nil {
			log.Fatal(err)
		}
		if nwrite != *MsgSize {
			log.Fatalf("bad nwrite = %d", nwrite)
		}
		nread, err := conn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if nread != *MsgSize {
			log.Fatalf("bad nread = %d", nread)
		}
	}
	elapsed := time.Since(t1)

	totalpings := int64(*NumPings * 2)
	fmt.Println("Client done")
	fmt.Printf("%d pingpongs took %d ms; avg. latency %d ns\n",
		totalpings, elapsed.Milliseconds(),
		elapsed.Nanoseconds()/totalpings)

	time.Sleep(50 * time.Millisecond)
}
