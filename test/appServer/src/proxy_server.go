/*
This program implements a simple proxy server that can forward requests using either HTTP or UDP.

Main Features:
1. Forwards client requests to a specified backend URL using HTTP or UDP.
2. Controls the timeout for backend requests.
3. Returns the backend response to the client, including success status and data or error message.

Usage:
go run proxy_server.go -port=<port> -timeout=<seconds>

Options:
-h: Display help information
-port: Specify the TCP port for the server to listen on (default is 8090)
-timeout: Specify the default timeout for backend requests in seconds (default is 4)

Notes:
- The server listens on the specified port.

Testing with curl:
- To test the proxy server over IPv4, use:
  curl -X POST http://127.0.0.1:8090 -d '{"BackendUrl":"http://127.0.0.1:8080","Timeout":5,"ForwardType":"http"}'  | jq .

- To test the proxy server over IPv6, use:
  curl -X POST http://\[::1\]:8090 -d '{"BackendUrl":"http://[::1]:8080","Timeout":5,"ForwardType":"udp"}'  | jq .
*/

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"main/common"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

var requestCount int
var mutex sync.Mutex

func main() {
	// Define command-line flags
	help := flag.Bool("h", false, "Display help information")
	port := flag.String("port", "8090", "Specify the TCP port for the server to listen on")
	defaultTimeout := flag.Int("timeout", 4, "Specify the default timeout for backend requests in seconds")
	flag.Parse()

	// If the -h flag is set, display help information and exit
	if *help {
		flag.Usage()
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		requestCount++
		currentRequestCount := requestCount
		mutex.Unlock()

		serverIP, _, err := common.GetServerIPAndPort()
		if err != nil {
			http.Error(w, "Unable to determine server IP", http.StatusInternalServerError)
			return
		}

		var clientReq common.ProxyClientRequest
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sendProxyResponse(w, r, common.ProxyResponse{
				Success:         false,
				ErrorMessage:    "Unable to read request body. Ensure it is valid JSON.",
				BackendResponse: "",
				BackendUrl:      clientReq.BackendUrl,
				FrontUrl:        constructFullURL(r),
				FrontIP:         serverIP,
				FrontPort:       *port,
				RequestCounter:  currentRequestCount,
				ForwardType:     clientReq.ForwardType,
			})
			return
		}

		if err := json.Unmarshal(body, &clientReq); err != nil {
			sendProxyResponse(w, r, common.ProxyResponse{
				Success:         false,
				ErrorMessage:    "Invalid request format. Ensure it is a valid JSON with required fields.",
				BackendResponse: "",
				BackendUrl:      clientReq.BackendUrl,
				FrontUrl:        constructFullURL(r),
				FrontIP:         serverIP,
				FrontPort:       *port,
				RequestCounter:  currentRequestCount,
				ForwardType:     clientReq.ForwardType,
			})
			return
		}

		if clientReq.BackendUrl == "" {
			sendProxyResponse(w, r, common.ProxyResponse{
				Success:         false,
				ErrorMessage:    "BackendUrl is required. Please provide a valid URL for the backend server.",
				BackendResponse: "",
				BackendUrl:      clientReq.BackendUrl,
				FrontUrl:        constructFullURL(r),
				FrontIP:         serverIP,
				FrontPort:       *port,
				RequestCounter:  currentRequestCount,
				ForwardType:     clientReq.ForwardType,
			})
			return
		}

		// Validate BackendUrl based on ForwardType
		if clientReq.ForwardType == "http" {
			if !isValidHTTPURL(clientReq.BackendUrl) {
				sendProxyResponse(w, r, common.ProxyResponse{
					Success:         false,
					ErrorMessage:    "Invalid HTTP URL format for BackendUrl. Use a valid HTTP URL, e.g., 'http://example.com'.",
					BackendResponse: "",
					BackendUrl:      clientReq.BackendUrl,
					FrontUrl:        constructFullURL(r),
					FrontIP:         serverIP,
					FrontPort:       *port,
					RequestCounter:  currentRequestCount,
					ForwardType:     clientReq.ForwardType,
				})
				return
			}
		} else if clientReq.ForwardType == "udp" {
			if !isValidUDPAddress(clientReq.BackendUrl) {
				sendProxyResponse(w, r, common.ProxyResponse{
					Success:         false,
					ErrorMessage:    "Invalid UDP address format for BackendUrl. Use a valid UDP address, e.g., 'localhost:8080'.",
					BackendResponse: "",
					BackendUrl:      clientReq.BackendUrl,
					FrontUrl:        constructFullURL(r),
					FrontIP:         serverIP,
					FrontPort:       *port,
					RequestCounter:  currentRequestCount,
					ForwardType:     clientReq.ForwardType,
				})
				return
			}
		} else {
			sendProxyResponse(w, r, common.ProxyResponse{
				Success:         false,
				ErrorMessage:    "Unsupported ForwardType. Supported values are 'http' and 'udp'.",
				BackendResponse: "",
				BackendUrl:      clientReq.BackendUrl,
				FrontUrl:        constructFullURL(r),
				FrontIP:         serverIP,
				FrontPort:       *port,
				RequestCounter:  currentRequestCount,
				ForwardType:     clientReq.ForwardType,
			})
			return
		}

		timeout := time.Duration(clientReq.Timeout) * time.Second
		if clientReq.Timeout == 0 {
			timeout = time.Duration(*defaultTimeout) * time.Second
		}

		switch clientReq.ForwardType {
		case "http":
			handleHTTPForwarding(w, r, clientReq, serverIP, *port, currentRequestCount, timeout)
		case "udp":
			handleUDPForwarding(w, r, clientReq, serverIP, *port, currentRequestCount, timeout)
		}
	})

	// Start the HTTP server
	address := fmt.Sprintf(":%s", *port)
	fmt.Printf("Proxy server is listening on port %s\n", *port)
	if err := http.ListenAndServe(address, nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}

// isValidHTTPURL checks if the given URL is a valid HTTP URL
func isValidHTTPURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// isValidUDPAddress checks if the given address is a valid UDP address
func isValidUDPAddress(address string) bool {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	if host == "" || port == "" {
		return false
	}
	return true
}

// handleHTTPForwarding handles HTTP forwarding to the backend server
func handleHTTPForwarding(w http.ResponseWriter, r *http.Request, clientReq common.ProxyClientRequest, serverIP, port string, requestCounter int, timeout time.Duration) {
	if clientReq.BackendUrl == "" {
		sendProxyResponse(w, r, common.ProxyResponse{
			Success:         false,
			ErrorMessage:    "BackendUrl is required",
			BackendResponse: "",
			BackendUrl:      clientReq.BackendUrl,
			FrontUrl:        constructFullURL(r),
			FrontIP:         serverIP,
			FrontPort:       port,
			RequestCounter:  requestCounter,
			ForwardType:     clientReq.ForwardType,
		})
		return
	}

	client := &http.Client{Timeout: timeout}

	// Parse the backend URL to extract the host and port
	parsedURL, err := url.Parse(clientReq.BackendUrl)
	if err != nil {
		sendProxyResponse(w, r, common.ProxyResponse{
			Success:         false,
			ErrorMessage:    fmt.Sprintf("Invalid BackendUrl: %v", err),
			BackendResponse: "",
			BackendUrl:      clientReq.BackendUrl,
			FrontUrl:        constructFullURL(r),
			FrontIP:         serverIP,
			FrontPort:       port,
			RequestCounter:  requestCounter,
			ForwardType:     clientReq.ForwardType,
		})
		return
	}

	backendHost, backendPort, err := net.SplitHostPort(parsedURL.Host)
	if err != nil {
		backendHost = parsedURL.Host
		backendPort = "80" // Default to port 80 if not specified
	}

	// Resolve the backend IP address
	backendIPs, err := net.LookupIP(backendHost)
	if err != nil || len(backendIPs) == 0 {
		sendProxyResponse(w, r, common.ProxyResponse{
			Success:         false,
			ErrorMessage:    fmt.Sprintf("Failed to resolve backend IP: %v", err),
			BackendResponse: "",
			BackendUrl:      clientReq.BackendUrl,
			FrontUrl:        constructFullURL(r),
			FrontIP:         serverIP,
			FrontPort:       port,
			RequestCounter:  requestCounter,
			ForwardType:     clientReq.ForwardType,
		})
		return
	}
	backendIP := backendIPs[0].String()

	// Send EchoData as the request body
	resp, err := client.Post(clientReq.BackendUrl, "application/json", bytes.NewBuffer([]byte(clientReq.EchoData)))
	if err != nil {
		sendProxyResponse(w, r, common.ProxyResponse{
			Success:         false,
			ErrorMessage:    fmt.Sprintf("Failed to access backend: %v", err),
			BackendResponse: "",
			BackendUrl:      clientReq.BackendUrl,
			BackendIP:       backendIP,
			BackendPort:     backendPort,
			FrontUrl:        constructFullURL(r),
			FrontIP:         serverIP,
			FrontPort:       port,
			RequestCounter:  requestCounter,
			ForwardType:     clientReq.ForwardType,
		})
		return
	}
	defer resp.Body.Close()

	backendData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sendProxyResponse(w, r, common.ProxyResponse{
			Success:         false,
			ErrorMessage:    fmt.Sprintf("Failed to read backend response: %v", err),
			BackendResponse: "",
			BackendUrl:      clientReq.BackendUrl,
			BackendIP:       backendIP,
			BackendPort:     backendPort,
			FrontUrl:        constructFullURL(r),
			FrontIP:         serverIP,
			FrontPort:       port,
			RequestCounter:  requestCounter,
			ForwardType:     clientReq.ForwardType,
		})
		return
	}

	sendProxyResponse(w, r, common.ProxyResponse{
		Success:         true,
		BackendResponse: string(backendData),
		ErrorMessage:    "",
		BackendUrl:      clientReq.BackendUrl,
		BackendIP:       backendIP,
		BackendPort:     backendPort,
		FrontUrl:        constructFullURL(r),
		FrontIP:         serverIP,
		FrontPort:       port,
		RequestCounter:  requestCounter,
		ForwardType:     clientReq.ForwardType,
	})
}

// handleUDPForwarding handles UDP forwarding to the backend server
func handleUDPForwarding(w http.ResponseWriter, r *http.Request, clientReq common.ProxyClientRequest, serverIP, port string, requestCounter int, timeout time.Duration) {
	backendAddr, err := net.ResolveUDPAddr("udp", clientReq.BackendUrl)
	if err != nil {
		sendProxyResponse(w, r, common.ProxyResponse{
			Success:         false,
			ErrorMessage:    "Failed to resolve backend address. Ensure BackendUrl is a valid UDP address.",
			BackendResponse: "",
			BackendUrl:      clientReq.BackendUrl,
			FrontUrl:        constructFullURL(r),
			FrontIP:         serverIP,
			FrontPort:       port,
			RequestCounter:  requestCounter,
			ForwardType:     clientReq.ForwardType,
		})
		return
	}

	// Forward the EchoData to the backend server
	backendConn, err := net.DialUDP("udp", nil, backendAddr)
	if err != nil {
		sendProxyResponse(w, r, common.ProxyResponse{
			Success:         false,
			ErrorMessage:    "Failed to connect to backend server. Ensure the backend server is reachable via UDP.",
			BackendResponse: "",
			BackendUrl:      clientReq.BackendUrl,
			BackendIP:       backendAddr.IP.String(),
			BackendPort:     fmt.Sprintf("%d", backendAddr.Port),
			FrontUrl:        constructFullURL(r),
			FrontIP:         serverIP,
			FrontPort:       port,
			RequestCounter:  requestCounter,
			ForwardType:     clientReq.ForwardType,
		})
		return
	}
	defer backendConn.Close()

	_, err = backendConn.Write([]byte(clientReq.EchoData))
	if err != nil {
		sendProxyResponse(w, r, common.ProxyResponse{
			Success:         false,
			ErrorMessage:    "Failed to send data to backend server. Ensure the data can be sent to the backend server.",
			BackendResponse: "",
			BackendUrl:      clientReq.BackendUrl,
			BackendIP:       backendAddr.IP.String(),
			BackendPort:     fmt.Sprintf("%d", backendAddr.Port),
			FrontUrl:        constructFullURL(r),
			FrontIP:         serverIP,
			FrontPort:       port,
			RequestCounter:  requestCounter,
			ForwardType:     clientReq.ForwardType,
		})
		return
	}

	// Set a read deadline for the response
	backendConn.SetReadDeadline(time.Now().Add(timeout))

	// Read the response from the backend server
	buffer := make([]byte, 1024)
	n, _, err := backendConn.ReadFromUDP(buffer)
	if err != nil {
		sendProxyResponse(w, r, common.ProxyResponse{
			Success:         false,
			ErrorMessage:    "Failed to read response from backend server. Ensure the backend server sends a valid response.",
			BackendResponse: "",
			BackendUrl:      clientReq.BackendUrl,
			BackendIP:       backendAddr.IP.String(),
			BackendPort:     fmt.Sprintf("%d", backendAddr.Port),
			FrontUrl:        constructFullURL(r),
			FrontIP:         serverIP,
			FrontPort:       port,
			RequestCounter:  requestCounter,
			ForwardType:     clientReq.ForwardType,
		})
		return
	}

	sendProxyResponse(w, r, common.ProxyResponse{
		Success:         true,
		BackendResponse: string(buffer[:n]),
		ErrorMessage:    "",
		BackendUrl:      clientReq.BackendUrl,
		BackendIP:       backendAddr.IP.String(),
		BackendPort:     fmt.Sprintf("%d", backendAddr.Port),
		FrontUrl:        constructFullURL(r),
		FrontIP:         serverIP,
		FrontPort:       port,
		RequestCounter:  requestCounter,
		ForwardType:     clientReq.ForwardType,
	})
}

// constructFullURL constructs the full URL from the request
func constructFullURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.Path)
}

// sendProxyResponse marshals the response data to JSON and writes it to the response writer
func sendProxyResponse(w http.ResponseWriter, r *http.Request, response common.ProxyResponse) {
	hostname, _ := os.Hostname()
	clientIP, clientPort, _ := net.SplitHostPort(r.RemoteAddr)
	_, ipVersion := common.GetServerIPAndVersion(r)

	response.ProxyHostName = hostname
	response.ClientIP = clientIP
	response.ClientPort = clientPort
	response.IPVersion = ipVersion

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Unable to marshal response data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)

	log.Printf("Sent response: %s", responseJSON)
}
