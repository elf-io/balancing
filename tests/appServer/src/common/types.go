// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package common

//--------------------------------- for udp server

// UdpServerResponse represents the structure of the UDP server response data
type UdpServerResponse struct {
	ServerHostName   string `json:"ServerHostName"`   // The hostname of the server
	ClientIP         string `json:"ClientIP"`         // The IP address of the client
	ClientPort       string `json:"ClientPort"`       // The port of the client
	ServerIP         string `json:"ServerIP"`         // The IP address of the server
	ServerPort       string `json:"ServerPort"`       // The port on which the server is listening
	IPVersion        string `json:"IPVersion"`        // The IP version (IPv4 or IPv6)
	ClientEchoData   string `json:"ClientEchoData"`   // The data echoed from the client's request
	RequestTimestamp string `json:"RequestTimestamp"` // The timestamp of the request
	RequestCounter   int    `json:"RequestCounter"`   // The count of requests since the server started
	ServerType       string `json:"ServerType"`       // The type of server (udp)
	EnvList          map[string]string `json:"EnvList"` // The list of environment variables
}

//--------------------------------- for http server

// HttpServerResponse represents the structure of the HTTP server response data
type HttpServerResponse struct {
	ServerHostName     string            `json:"ServerHostName"`     // The hostname of the server
	ClientIP           string            `json:"ClientIP"`           // The IP address of the client
	ClientPort         string            `json:"ClientPort"`         // The port of the client
	ServerIP           string            `json:"ServerIP"`           // The IP address of the server
	ServerPort         string            `json:"ServerPort"`         // The port on which the server is listening
	IPVersion          string            `json:"IPVersion"`          // The IP version (IPv4 or IPv6)
	ClientEchoData     string            `json:"ClientEchoData"`     // The data echoed from the client's request
	RequestHttpHeaders map[string]string `json:"RequestHttpHeaders"` // The HTTP headers from the client's request
	RequestTimestamp   string            `json:"RequestTimestamp"`   // The timestamp of the request
	URL                string            `json:"URL"`                // The URL of the request
	RequestCounter     int               `json:"RequestCounter"`     // The count of requests since the server started
	ServerType         string            `json:"ServerType"`         // The type of server (http)
	EnvList            map[string]string `json:"EnvList"`             // The list of environment variables
}

//--------------------------------- for proxy server

// ProxyResponse represents the structure of the proxy server response data
type ProxyResponse struct {
	Success         bool   `json:"Success"`         // Indicates if the request was successful
	BackendResponse string `json:"BackendResponse"` // The response data from the backend server
	ErrorMessage    string `json:"ErrorMessage"`    // Error message, if any
	ProxyHostName   string `json:"ProxyHostName"`   // The hostname of the proxy server
	ClientIP        string `json:"ClientIP"`        // The IP address of the client
	ClientPort      string `json:"ClientPort"`      // The port of the client
	IPVersion       string `json:"IPVersion"`       // The IP version (IPv4 or IPv6)
	BackendUrl      string `json:"BackendUrl"`      // The URL of the backend server
	BackendIP       string `json:"BackendIP"`       // The IP address of the backend server
	BackendPort     string `json:"BackendPort"`     // The port of the backend server
	FrontUrl        string `json:"FrontUrl"`        // The URL of the front-end request
	FrontIP         string `json:"FrontIP"`         // The IP address of the proxy server
	FrontPort       string `json:"FrontPort"`       // The port of the proxy server
	RequestCounter  int    `json:"RequestCounter"`  // The count of requests since the proxy server started
	ForwardType     string `json:"ForwardType"`     // The type of forwarding (http or udp)
}

// ProxyClientRequest represents the structure of the client's request body
type ProxyClientRequest struct {
	BackendUrl  string `json:"BackendUrl"`  // The backend URL requested by the client
	Timeout     int    `json:"Timeout"`     // The timeout for the request in seconds
	ForwardType string `json:"ForwardType"` // The type of forwarding (http or udp)
	EchoData    string `json:"EchoData"`    // The data to be echoed back by the server
}
