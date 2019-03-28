package main

import (
	"flag"
	"io"
	"log"
	"net"
)

var (
	srchost = flag.String("srchost", "0.0.0.0:81", "")
	dsthost = flag.String("dsthost", "0.0.0.0:80", "")
)

func main() {
	flag.Parse()

	server, err := net.Listen("tcp", *srchost)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Listening on '%s'\n", *srchost)

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

	remote, err := net.Dial("tcp", *dsthost)
	if err != nil {
		log.Fatalln(err)
	}
	defer remote.Close()
	log.Printf("Connection to server '%v' from '%v'\n", remote.RemoteAddr(), client.RemoteAddr())

	go io.Copy(remote, client)
	io.Copy(client, remote)
}
