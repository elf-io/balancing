/*
This program implements a simple HTTP server.

Main Features:
1. Returns the server's hostname when an HTTP request is received.
2. Returns the client's source IP address.
3. Echoes any data from the client's request.

Usage:
go run http_server.go -port=<port>

Options:
-h: Display help information
-port: Specify the TCP port for the server to listen on (default is 8080)

Notes:
- The server listens on the specified port.

Testing with curl:
- To test the server over IPv4, use:
  curl http://127.0.0.1:8080
- To test the server over IPv6, use:
  curl http://[::1]:8080
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"main/common"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var requestCount int
var mutex sync.Mutex

func main() {
	// Define command-line flags
	help := flag.Bool("h", false, "Display help information")
	port := flag.String("port", "8080", "Specify the TCP port for the server to listen on")
	flag.Parse()

	// If the -h flag is set, display help information and exit
	if *help {
		flag.Usage()
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, *port)
	})

	// 添加 /healthy 路由
	http.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start the HTTP server
	address := fmt.Sprintf(":%s", *port)
	fmt.Printf("Server is listening on port %s\n", *port)
	if err := http.ListenAndServe(address, nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}

// handleRequest processes incoming HTTP requests
func handleRequest(w http.ResponseWriter, r *http.Request, serverPort string) {
	mutex.Lock()
	requestCount++
	currentRequestCount := requestCount
	mutex.Unlock()

	serverHostName, clientIP, clientPort, serverIP, ipVersion, echoData, requestHttpHeaders, err := processRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	envList := common.GetEnvironmentVariables("ENV_")

	response := common.HttpServerResponse{
		ServerHostName:     serverHostName,
		ClientIP:           clientIP,
		ClientPort:         clientPort,
		ServerIP:           serverIP,
		ServerPort:         serverPort, // Use the specified server port
		IPVersion:          ipVersion,
		ClientEchoData:     echoData,
		RequestHttpHeaders: requestHttpHeaders,
		RequestTimestamp:   time.Now().Format(time.RFC3339),
		URL:                r.URL.String(),
		RequestCounter:     currentRequestCount,
		ServerType:         "http",  // Set server type to http
		EnvList:            envList, // Add environment variables to the response
	}

	if err := sendResponse(w, response); err != nil {
		http.Error(w, "Unable to send response", http.StatusInternalServerError)
	}
}

// processRequest extracts and logs request data
func processRequest(r *http.Request) (string, string, string, string, string, string, map[string]string, error) {
	serverHostName, err := os.Hostname()
	if err != nil {
		return "", "", "", "", "", "", nil, fmt.Errorf("unable to get hostname: %v", err)
	}

	clientIP, clientPort, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", "", "", "", "", "", nil, fmt.Errorf("unable to parse client IP address: %v", err)
	}

	serverIP, ipVersion := common.GetServerIPAndVersion(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", "", "", "", "", "", nil, fmt.Errorf("unable to read request body: %v", err)
	}

	requestHttpHeaders := make(map[string]string)
	for name, values := range r.Header {
		requestHttpHeaders[name] = values[0] // Assuming single value for simplicity
	}

	echoData := string(body)
	log.Printf("Received request from %s:%s with data: %s", clientIP, clientPort, echoData)

	return serverHostName, clientIP, clientPort, serverIP, ipVersion, echoData, requestHttpHeaders, nil
}

// sendResponse marshals the response data to JSON and writes it to the response writer
func sendResponse(w http.ResponseWriter, response common.HttpServerResponse) error {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("unable to marshal response data: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)

	log.Printf("Sent response: %s", responseJSON)
	return nil
}
