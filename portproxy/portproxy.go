package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	src = flag.String("src", "", "host:port")
	dst = flag.String("dst", "", "host:port")
)

func main() {
	flag.Parse()

	if *src == "" || *dst == "" {
		usageAndExit("")
	}

	server, err := net.Listen("tcp", *src)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Listening on '%s'\n", *dst)

	for {
		client, err := server.Accept()
		if err != nil {
			log.Fatalln(err)
		}

		go handleRequest(client)
	}
}

func handleRequest(client net.Conn) {
	defer client.Close()

	remote, err := net.Dial("tcp", *dst)
	if err != nil {
		log.Fatalln(err)
	}
	defer remote.Close()
	log.Printf("Connection to server '%v' from '%v'\n", remote.RemoteAddr(), client.RemoteAddr())

	go io.Copy(remote, client)
	io.Copy(client, remote)
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}
