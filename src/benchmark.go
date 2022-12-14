package main

import (
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"os"
	"time"
)

var UnixDomain = flag.Bool("unixdomain", true, "Use Unix domain sockets")
var MsgSize = flag.Int("msgsize", 128, "Message size in each ping")
var NumPings = flag.Int("n", 1_000_000, "Number of pings to measure")
var UseAncillaryData = flag.Bool("useAncillaryData", false, "Use ancillary data")

var UnixAddress = "/tmp/test.sock"

func domainAndAddress() (string, string) {
	return "unixgram", UnixAddress
}

func server() {
	if *UnixDomain {
		if err := os.RemoveAll(UnixAddress); err != nil {
			panic(err)
		}
	}

	domain, address := domainAndAddress()
	unixAddress, err := net.ResolveUnixAddr("unixgram", address)
	if err != nil {
		log.Panicln("could not resolve unix gram")
	}
	conn, err := net.ListenUnixgram(domain, unixAddress)
	if err != nil {
		log.Panicln(err)
	}

	if *UseAncillaryData {
		enableUDSPassCred(conn)
		if err != nil {
			log.Panicln(err)
		}
	}
	defer conn.Close()

	buf := make([]byte, *MsgSize)
	oob := make([]byte, *MsgSize)
	for n := 0; n < *NumPings; n++ {
		if *UseAncillaryData {
			nread, nOob, _, _, err := conn.ReadMsgUnix(buf, oob)
			if err != nil {
				log.Panicln(err)
			}
			if nread != *MsgSize {
				log.Fatalf("bad nread = %d", nread)
			}
			if nOob == 0 {
				log.Fatalf("bad nOob = %d", nread)
			}
		} else {
			nread, _, _, _, err := conn.ReadMsgUnix(buf, oob)
			if err != nil {
				log.Panicln(err)
			}
			if nread != *MsgSize {
				log.Fatalf("bad nread = %d", nread)
			}
		}
		// fmt.Println(string(buf))
		// nwrite, err := conn.Write(buf)
		// if err != nil {
		// 	log.Panicln(err)
		// }
		// if nwrite != *MsgSize {
		// 	log.Fatalf("bad nwrite = %d", nwrite)
		// }
	}

	time.Sleep(50 * time.Millisecond)
}

func enableUDSPassCred(conn *net.UnixConn) error {
	rawconn, err := conn.SyscallConn()
	if err != nil {
		return err
	}

	return rawconn.Control(func(fd uintptr) {
		unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_PASSCRED, 1) //nolint:errcheck
	})
}

func main() {
	flag.Parse()

	go server()
	time.Sleep(50 * time.Millisecond)

	// This is the client code in the main goroutine.
	domain, address := domainAndAddress()
	conn, err := net.Dial(domain, address)
	if err != nil {
		log.Panicln(err)
	}

	buf := make([]byte, *MsgSize)
	copy(buf, "some data :)")
	t1 := time.Now()
	for n := 0; n < *NumPings; n++ {
		nwrite, err := conn.Write(buf)
		if err != nil {
			log.Panicln(err)
		}
		if nwrite != *MsgSize {
			log.Fatalf("bad nwrite = %d", nwrite)
		}
		// nread, err := conn.Read(buf)
		// if err != nil {
		// 	log.Panicln(err)
		// }
		// if nread != *MsgSize {
		// 	log.Fatalf("bad nread = %d", nread)
		// }
	}
	elapsed := time.Since(t1)

	totalpings := int64(*NumPings * 2)
	fmt.Println("Client done")
	fmt.Printf("%d pingpongs took %d ms; avg. latency %d ns\n",
		totalpings, elapsed.Milliseconds(),
		elapsed.Nanoseconds()/totalpings)

	time.Sleep(50 * time.Millisecond)
}
