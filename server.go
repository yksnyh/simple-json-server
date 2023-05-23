package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(statusCode int) {
	sr.statusCode = statusCode
	sr.ResponseWriter.WriteHeader(statusCode)
}

func main() {
	http.HandleFunc("/", handler)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8888"
	}
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	body := new(bytes.Buffer)
	body.ReadFrom(r.Body)
	sr := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	defer func() {
		logLine := fmt.Sprintf("%s - %s %s %d %v", r.RemoteAddr, r.Method, r.URL.Path, sr.statusCode, time.Since(start))

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

	if strings.HasPrefix(r.URL.Path, "/html/") {
		staticFilePath := filepath.Join(".", r.URL.Path)
		http.ServeFile(sr, r, staticFilePath)
	} else {
		delay := os.Getenv("API_REQUEST_DELAY_MS")
		delayInMs, err := strconv.Atoi(delay)
		if err == nil {
			time.Sleep(time.Duration(delayInMs) * time.Millisecond)
		}
		handleJSONRequest(sr, r)
	}
}

func handleJSONRequest(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	if os.Getenv("CORS") == "1" {
		origin := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	jsonFilePath := filepath.Join("data", strings.ToLower(method))
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

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"error\": \"Internal Server Error\"}")
		return
	}

	response, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"error\": \"Internal Server Error\"}")
		return
	}

	w.Write(response)
}
