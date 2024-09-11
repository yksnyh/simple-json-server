package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DEFAULT_CONTENTTYPE = "application/json;charset=UTF-8"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(statusCode int) {
	sr.statusCode = statusCode
	sr.ResponseWriter.WriteHeader(statusCode)
}

var responseConf []ResponseConfig

func main() {
	conf, err := ReadResponseConfig()
	if err != nil {
		log.Println("Not found config file", err)
	} else {
		responseConf = conf
	}

	http.HandleFunc("/", handler)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8888"
	}
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
	// log.Fatal(http.ListenAndServeTLS(":"+port, "debug.crt", "debug.key", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	body := new(bytes.Buffer)
	body.ReadFrom(r.Body)
	sr := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	defer func() {
		logLine := fmt.Sprintf("%s - %s %s %d %v", r.RemoteAddr, r.Method, r.URL.String(), sr.statusCode, time.Since(start))

		if os.Getenv("DETAILED_LOGGING") == "1" {
			headers := make([]string, 0, len(r.Header))
			for name, header := range r.Header {
				headers = append(headers, fmt.Sprintf("%v: %v", name, header))
			}
			logLine += ", " + strings.Join(headers, ", ")
			logLine += ", Body: " + body.String()
		}
		log.Print(logLine)
	}()

	conf := MatchResponseConfig(r, responseConf)
	if conf.Delay > 0 {
		time.Sleep(time.Duration(conf.Delay) * time.Millisecond)
	}

	staticContentDirs := os.Getenv("STATIC_CONTENT_DIRS")
	if staticContentDirs == "" {
		staticContentDirs = "html"
	}
	for _, dir := range strings.Split(staticContentDirs, ",") {
		if strings.HasPrefix(r.URL.Path, "/"+dir+"/") {
			staticFilePath := filepath.Join(".", r.URL.Path)
			http.ServeFile(sr, r, staticFilePath)
			return
		}
	}

	if os.Getenv("CORS") == "1" {
		origin := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	handleApiRequest(sr, r, conf)
}

func handleApiRequest(w http.ResponseWriter, r *http.Request, conf ResponseConfig) {
	if len(conf.ContentType) > 0 {
		w.Header().Set("Content-Type", conf.ContentType)
	} else {
		w.Header().Set("Content-Type", DEFAULT_CONTENTTYPE)
	}

	if conf.HttpStatus > 0 && conf.HttpStatus != 200 {
		w.WriteHeader(conf.HttpStatus)
	}

	jsonFilePath := filepath.Join("data", strings.ToLower(r.Method))
	jsonFilePath += r.URL.Path
	jsonFilePath += ".json"

	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "{\"error\": \"File not found\"}")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\": \"Internal Server Error\"}")
		}
		return
	}
	w.Write(jsonData)
}
