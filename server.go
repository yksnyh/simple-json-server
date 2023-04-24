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

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	method := r.Method
	start := time.Now()
	statusCode := -1

	defer func() {
		log.Printf("%s - %s %s %d %v", r.RemoteAddr, method, r.URL.Path, statusCode, time.Since(start))
	}()

	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	if method == "OPTIONS" {
		statusCode = http.StatusOK
		w.WriteHeader(statusCode)
		return
	}

	jsonFilePath := filepath.Join("data", strings.ToLower(method))
	jsonFilePath += r.URL.Path
	jsonFilePath += ".json"

	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			statusCode = http.StatusNotFound
			fmt.Fprintf(w, "{\"error\": \"File not found\"}")
		} else {
			statusCode = http.StatusInternalServerError
			fmt.Fprintf(w, "{\"error\": \"Internal Server Error\"}")
		}
		w.WriteHeader(statusCode)
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		statusCode = http.StatusInternalServerError
		fmt.Fprintf(w, "{\"error\": \"Internal Server Error\"}")
		w.WriteHeader(statusCode)
		return
	}

	response, err := json.Marshal(result)
	if err != nil {
		statusCode = http.StatusInternalServerError
		fmt.Fprintf(w, "{\"error\": \"Internal Server Error\"}")
		w.WriteHeader(statusCode)
		return
	}

	w.Write(response)
	statusCode = http.StatusOK
}
