package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	rebuildTimeout = 180
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
	start := time.Now()
	resp, err := client.Do(req)
	callDuration := time.Since(start)
	if err != nil {
		log.Printf("http call error with execution time: %v\n", callDuration)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("http call wrong status with execution time: %v\n", callDuration)
		b, _ := httputil.DumpResponse(resp, true)
		return fmt.Errorf("Wrong Status: %v", string(b))
	}
	log.Printf("http call success with execution time: %v\n", callDuration)
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

func readVaultData() (string, string, error) {
	jsonFile, err := os.Open("/volume/vault/secrets.json")
	if err != nil {
		log.Println(err)
		return "", "", err
	}
	log.Println("Successfully opened secrets.json")
	defer jsonFile.Close()
	b, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return "", "", err
	}
	var result map[string]interface{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return "", "", err
	}
	data := result["data"].(map[string]interface{})
	return data["pypicloud-user"].(string), data["pypicloud-pass"].(string), nil
}

func main() {
	username, password, err := readVaultData()
	if err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
	log.Printf("--%s--\n", username)
	log.Printf("--%s--\n", password)

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

	log.Printf("--%s--\n", username)
	log.Printf("--%s--\n", password)

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
