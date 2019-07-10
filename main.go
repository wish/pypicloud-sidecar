package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	username       = ""
	password       = ""
	rebuildTimeout = 60
	adminURL       = ""

	// TODO(tvi): Add metrics for last fetch.
)

const refreshPath = "/admin/rebuild"

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func reload() error {
	req, err := http.NewRequest("GET", adminURL, nil)
	req.Header.Add("Authorization", "Basic "+basicAuth(username, password))

	timeout := time.Duration(rebuildTimeout) * time.Second
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := httputil.DumpResponse(resp, true)
		return fmt.Errorf("Wrong Status: %v", string(b))
	}
	return nil
}

// TODO(tvi): Write this.
func ready(w http.ResponseWriter, _ *http.Request) {
	io.WriteString(w, "OK\n")
}

func health(w http.ResponseWriter, _ *http.Request) {
	err := reload()
	if err != nil {
		log.Printf("Got error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, "OK\n")
}

func main() {
	username = os.Getenv("USERNAME")
	if username == "" {
		log.Printf("USERNAME environment variable not set\n")
		os.Exit(1)
	}

	password = os.Getenv("PASSWORD")
	if password == "" {
		log.Printf("PASSWORD environment variable not set\n")
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		log.Printf("Usage: sidecar <pypicloud-url> <rebuild-timeout>\n")
		os.Exit(1)
	}

	u, err := url.Parse(os.Args[1])
	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	rebuildTimeout, err = strconv.Atoi(os.Args[2])
	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	u.Path = refreshPath
	adminURL = u.String()

	ok := func(w http.ResponseWriter, _ *http.Request) { io.WriteString(w, "OK\n") }

	http.HandleFunc("/", ok)
	http.HandleFunc("/ready", ready) // TODO(tvi): Swap over.
	http.HandleFunc("/health", health)

	// TODO(tvi): Make bindAddr configurable.
	log.Printf("Listening on :8100\n")
	log.Fatal(http.ListenAndServe(":8100", nil))
}
