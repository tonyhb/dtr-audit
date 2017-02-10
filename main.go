package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var (
	host   string
	auth   string
	client *http.Client
)

func init() {
	// initialize our client with sane defaults.
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func main() {
	host := os.Getenv("HOST")
	user := os.Getenv("USER")
	pass := os.Getenv("PASS")

	auditor := NewAuditor(host, user, pass)
	err := auditor.Run()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(auditor, "", "  ")
	fmt.Printf("%s\n", data)
}
