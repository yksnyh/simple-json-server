package main

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
)

type ResponseConfig struct {
	Path        string `json:"path"`
	Delay       int    `json:"delay"`
	HttpStatus  int    `json:"httpStatus"`
	ContentType string `json:"contentType"`
}

func ReadResponseConfig() (config []ResponseConfig, err error) {
	filepath := "responseConfig.json"
	raw, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	err = json.Unmarshal(raw, &config)
	return
}

func MatchResponseConfig(r *http.Request, config []ResponseConfig) ResponseConfig {
	conf := ResponseConfig{r.URL.Path, 0, 0, ""}
	for _, c := range config {
		m, _ := regexp.MatchString(c.Path, r.URL.Path)
		if m {
			if conf.Delay == 0 && c.Delay > 0 {
				conf.Delay = c.Delay
			}
			if conf.HttpStatus == 0 && c.HttpStatus > 0 {
				conf.HttpStatus = c.HttpStatus
			}
			if len(conf.ContentType) == 0 && len(c.ContentType) > 0 {
				conf.ContentType = c.ContentType
			}
		}
	}
	return conf
}
