package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	help := flag.Bool("h", false, "Help message")
	port := flag.String("p", "8100", "Port to serve on")
	directory := flag.String("d", ".", "The directory of file to host")

	flag.Parse()

	if *help == true {
		flag.Usage()
		return
	}

	http.Handle("/", http.FileServer(http.Dir(*directory)))

	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
