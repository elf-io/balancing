// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package nodeId

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"math/rand"
	"net"
	"strconv"
	"time"
)

func stringToUint32(str string) (uint32, error) {
	num, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(num), nil
}

func Uint32ToString(num uint32) string {
	return strconv.FormatUint(uint64(num), 10)
}

func generateRandomUint32() uint32 {
	src := rand.New(rand.NewSource(time.Now().UnixNano()))
	return src.Uint32()
}

func CheckInterfaceAddresses(interfaceName string) (string, string, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", "", fmt.Errorf("interface %s not found: %v", interfaceName, err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", "", fmt.Errorf("failed to get addresses for interface %s: %v", interfaceName, err)
	}

	var ipv4Address, ipv6Address string

	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPNet:
			ip := v.IP
			if ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil && ipv4Address == "" {
				ipv4Address = ip.String()
			} else if ip.To16() != nil && ipv6Address == "" {
				ipv6Address = ip.String()
			}
		}
	}

	if ipv4Address == "" && ipv6Address == "" {
		return "", "", fmt.Errorf("interface %s does not have a valid IPv4 or IPv6 address", interfaceName)
	}

	return ipv4Address, ipv6Address, nil
}

func getNodeIpv4(node *corev1.Node) string {

	// loop node
	for _, v := range node.Status.Addresses {
		t := net.ParseIP(v.Address)
		if t == nil {
			continue
		}
		if t.To4() != nil {
			return t.To4().String()
		}
	}
	return ""
}

func getNodeIpv6(node *corev1.Node) string {
	// loop node
	for _, v := range node.Status.Addresses {
		t := net.ParseIP(v.Address)
		if t == nil {
			continue
		}
		if t.To4() == nil {
			return t.To16().String()
		}
	}
	return ""
}
