package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/delete", deleteFile)
	http.HandleFunc("/download", downloadFile)

	log.Printf("Serving %s on HTTP %s\n", *directory, *host)
	log.Fatal(http.ListenAndServe(*host, nil))
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		file, handler, err := r.FormFile("file")
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		defer file.Close()

		absPath, _ := filepath.Abs("./" + handler.Filename)

		if _, err = os.Stat(absPath); err == nil {
			w.Write([]byte("File '" + handler.Filename + "' already exists"))
			return
		}

		f, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		defer f.Close()

		io.Copy(f, file)

		log.Printf("Upload file %s", handler.Filename)
		w.Write([]byte("Upload success"))
	} else {
		w.Write([]byte("Invalid HTTP " + r.Method + " Method"))
	}
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		file := r.URL.Query().Get("file")
		if len(file) < 1 {
			w.Write([]byte("Url param 'file' is missing"))
			return
		}

		absPath, _ := filepath.Abs("./" + file)

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			w.Write([]byte("File '" + file + "' doesn't exists"))
			return
		}

		err := os.Remove(absPath)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		log.Printf("Delete file %s", file)
		w.Write([]byte("Delete success"))
	} else {
		w.Write([]byte("Invalid HTTP " + r.Method + " Method"))
	}
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		url := r.URL.Query().Get("url")
		if len(url) < 1 {
			w.Write([]byte("Url param 'url' is missing"))
			return
		}

		_, file := path.Split(url)
		fileName := strings.Split(file, "?")[0]
		absPath, _ := filepath.Abs("./" + fileName)

		if _, err := os.Stat(absPath); err == nil {
			w.Write([]byte("File '" + fileName + "' already exists"))
			return
		}

		w.Write([]byte("Downloading file " + fileName))

		go func() {
			log.Printf("Downloading file %s", fileName)

			fileOut, err := os.Create(absPath)
			if err != nil {
				log.Printf("Download file error %s", err.Error())
				return
			}
			defer fileOut.Close()

			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Download file error %s", err.Error())
				return
			}
			defer resp.Body.Close()

			io.Copy(fileOut, resp.Body)

			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				log.Printf("Downloaded file %s but removed", fileName)
				return
			}

			log.Printf("Downloaded file %s", fileName)
		}()
	} else {
		w.Write([]byte("Invalid HTTP " + r.Method + " Method"))
	}
}
