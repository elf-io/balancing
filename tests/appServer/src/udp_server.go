// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
/*
This program implements a simple UDP server.

Main Features:
1. Returns the server's hostname when a UDP packet is received.
2. Returns the client's source IP address.
3. Echoes any data from the client's request.

Usage:
go run udp_server.go -port=<port>

Options:
-h: Display help information
-port: Specify the UDP port for the server to listen on (default is 8080)

Notes:
- The server listens on the specified port.

Testing with netcat (nc) on Linux:
- To test the server, you can use the following netcat commands:
  1. Send a message to the server:
     echo "your data here" | nc -u -w1 localhost 8080
  2. Listen for responses from the server:
     nc -u -l 8080
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"main/common"
	"net"
	"os"
	"sync"
	"time"
)

var requestCount int
var mutex sync.Mutex

func main() {
	// Define command-line flags
	help := flag.Bool("h", false, "Display help information")
	port := flag.String("port", "8080", "Specify the UDP port for the server to listen on")
	flag.Parse()

	// If the -h flag is set, display help information and exit
	if *help {
		flag.Usage()
		return
	}

	// Start the UDP server
	address := fmt.Sprintf(":%s", *port)
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP port %s: %v", *port, err)
	}
	defer conn.Close()

	fmt.Printf("UDP server is listening on port %s\n", *port)

	buffer := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		go handleUDPRequest(conn, addr, buffer[:n], *port)
	}
}

// handleUDPRequest processes incoming UDP requests
func handleUDPRequest(conn *net.UDPConn, addr *net.UDPAddr, data []byte, port string) {
	mutex.Lock()
	requestCount++
	currentRequestCount := requestCount
	mutex.Unlock()

	serverHostName, err := os.Hostname()
	if err != nil {
		log.Printf("Unable to get hostname: %v", err)
		return
	}

	clientIP := addr.IP.String()
	clientPort := fmt.Sprintf("%d", addr.Port)
	serverIP, ipVersion := getServerIPAndVersion(addr)

	echoData := string(data)
	log.Printf("Received request from %s:%s with data: %s", clientIP, clientPort, echoData)

	envList := common.GetEnvironmentVariables("ENV_")

	response := common.UdpServerResponse{
		ServerHostName:   serverHostName,
		ClientIP:         clientIP,
		ClientPort:       clientPort,
		ServerIP:         serverIP,
		ServerPort:       port,
		IPVersion:        ipVersion,
		ClientEchoData:   echoData,
		RequestTimestamp: time.Now().Format(time.RFC3339),
		RequestCounter:   currentRequestCount,
		ServerType:       "udp",   // Set server type to udp
		EnvList:          envList, // Add environment variables to the response
	}

	if err := sendUDPResponse(conn, addr, response); err != nil {
		log.Printf("Unable to send response: %v", err)
	}
}

// getServerIPAndVersion determines the server IP and whether the request is IPv4 or IPv6
func getServerIPAndVersion(addr *net.UDPAddr) (string, string) {
	ip := addr.IP
	if ip.To4() != nil {
		return ip.String(), "IPv4"
	}
	return ip.String(), "IPv6"
}

// sendUDPResponse marshals the response data to JSON and sends it back to the client
func sendUDPResponse(conn *net.UDPConn, addr *net.UDPAddr, response common.UdpServerResponse) error {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("unable to marshal response data: %v", err)
	}

	_, err = conn.WriteToUDP(responseJSON, addr)
	if err != nil {
		return fmt.Errorf("unable to send response: %v", err)
	}

	log.Printf("Response JSON length: %d", len(responseJSON))
	log.Printf("Sent response to %s: %s", addr.String(), responseJSON)
	return nil
}
