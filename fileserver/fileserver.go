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

var (
	host    = flag.String("host", "0.0.0.0:8100", "Host to serve on")
	workDir = flag.String("dir", ".", "The directory of file to host")
)

func main() {
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(*workDir)))
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/delete", deleteFile)
	http.HandleFunc("/download", downloadFile)

	log.Printf("Serving %s on HTTP %s\n", *workDir, *host)
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

		absPath := getFileAbsPath(handler.Filename)

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

		absPath := getFileAbsPath(file)

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
		absPath := getFileAbsPath(fileName)

		tmpPrefix := "_tmp_"
		absTmpPath := getFileAbsPath(tmpPrefix + fileName)

		if _, err := os.Stat(absPath); err == nil {
			w.Write([]byte("File '" + fileName + "' already exists"))
			return
		}

		if _, err := os.Stat(absTmpPath); err == nil {
			w.Write([]byte("File '" + fileName + "' is downloading"))
			return
		}

		w.Write([]byte("Downloading file " + fileName))

		go func() {
			log.Printf("Downloading file %s", fileName)

			fileOut, err := os.Create(absTmpPath)
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

			if _, err := os.Stat(absTmpPath); os.IsNotExist(err) {
				log.Printf("Downloaded file %s but removed", fileName)
				return
			}

			if os.Rename(absTmpPath, absPath); err != nil {
				log.Printf("Downloaded file %s but rename fail", fileName)
				return
			}

			log.Printf("Downloaded file %s", fileName)
		}()
	} else {
		w.Write([]byte("Invalid HTTP " + r.Method + " Method"))
	}
}

func getFileAbsPath(name string) string {
	absPath, _ := filepath.Abs(strings.TrimRight(*workDir, "/") + "/" + name)
	return absPath
}
