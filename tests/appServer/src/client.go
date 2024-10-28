package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"main/common"
	"net"
	"net/http"
	"time"
)

func main() {
	// Test HTTP server
	testHTTPServer()

	// Test UDP server
	testUDPServer()

	// Test Proxy server with HTTP forwarding
	testProxyServer("http", "http://localhost:8080")

	// Test Proxy server with UDP forwarding
	testProxyServer("udp", "localhost:8080")
}

func testHTTPServer() {
	fmt.Println("Testing HTTP Server...")

	// Create a request to the HTTP server
	requestData := common.ProxyClientRequest{
		BackendUrl: "http://localhost:8080", // Ensure BackendUrl is set
		EchoData:   "Hello, HTTP!",
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		log.Fatalf("Error marshalling request body: %v", err)
	}

	fmt.Printf("HTTP Request: %+v\n", requestData)

	resp, err := http.Post("http://localhost:8080", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read and print the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	var response common.HttpServerResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatalf("Error unmarshalling response: %v", err)
	}

	fmt.Printf("HTTP Server Response: %+v\n\n", response)
}

func testUDPServer() {
	fmt.Println("Testing UDP Server...")

	// Create a UDP connection
	conn, err := net.Dial("udp", "localhost:8080")
	if err != nil {
		log.Fatalf("Error connecting to UDP server: %v", err)
	}
	defer conn.Close()

	// Send data to the UDP server
	requestData := common.ProxyClientRequest{
		BackendUrl: "localhost:8080", // Ensure BackendUrl is set correctly for UDP
		EchoData:   "Hello, UDP!",
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		log.Fatalf("Error marshalling request body: %v", err)
	}

	fmt.Printf("UDP Request: %+v\n", requestData)

	_, err = conn.Write(requestBody)
	if err != nil {
		log.Fatalf("Error sending data to UDP server: %v", err)
	}

	// Set a read deadline
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Read the response from the UDP server
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatalf("Error reading response from UDP server: %v", err)
	}

	var response common.UdpServerResponse
	if err := json.Unmarshal(buffer[:n], &response); err != nil {
		log.Fatalf("Error unmarshalling response: %v", err)
	}

	fmt.Printf("UDP Server Response: %+v\n\n", response)
}

func testProxyServer(forwardType, backendUrl string) {
	fmt.Printf("Testing Proxy Server with %s forwarding...\n", forwardType)

	// Construct the request body
	clientRequest := common.ProxyClientRequest{
		BackendUrl:  backendUrl, // Use the provided BackendUrl
		Timeout:     5,
		ForwardType: forwardType,
		EchoData:    fmt.Sprintf("Hello, %s!", forwardType),
	}

	requestBody, err := json.Marshal(clientRequest)
	if err != nil {
		log.Fatalf("Error marshalling request body: %v", err)
	}

	fmt.Printf("Proxy Request (%s): %+v\n", forwardType, clientRequest)

	// Create a request to the Proxy server
	resp, err := http.Post("http://localhost:8090", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Error making HTTP request to proxy server: %v", err)
	}
	defer resp.Body.Close()

	// Read and print the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body from proxy server: %v", err)
	}

	var response common.ProxyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatalf("Error unmarshalling response: %v", err)
	}

	fmt.Printf("Proxy Server Response (%s): %+v\n\n", forwardType, response)
}
