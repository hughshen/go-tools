package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	help := flag.Bool("help", false, "Help message")
	host := flag.String("host", "0.0.0.0:8100", "Host to serve on")
	directory := flag.String("dir", ".", "The directory of file to host")

	flag.Parse()

	if *help == true {
		flag.Usage()
		return
	}

	http.Handle("/", http.FileServer(http.Dir(*directory)))

	log.Printf("Serving %s on HTTP %s\n", *directory, *host)
	log.Fatal(http.ListenAndServe(*host, nil))
}
