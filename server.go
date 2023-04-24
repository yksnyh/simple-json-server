package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sr := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	defer func() {
		log.Printf("%s - %s %s %d %v", r.RemoteAddr, r.Method, r.URL.Path, sr.statusCode, time.Since(start))
	}()

	if strings.HasPrefix(r.URL.Path, "/html/") {
		staticFilePath := filepath.Join(".", r.URL.Path)
		http.ServeFile(sr, r, staticFilePath)
	} else {
		sr.Header().Set("Content-Type", "application/json")
		handleJSONRequest(sr, r)
	}
}

func handleJSONRequest(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	method := r.Method

	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	if method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
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
