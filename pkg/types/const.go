// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package types

const (
	OrgName = "elf.io"

	ApiVersion = "v1beta1"
	ApiGroup   = "balancing." + OrgName

	// use this annotation to mark an ID in the annotation of each node
	// "bpfElf.org/nodeId": "32BitNumber"
	NodeAnnotaitonNodeIdKey = OrgName + "/nodeId"

	// the user could mark the ip in the annotation of each node
	// "bpfElf.org/nodeProxyIpv4": "192.168.1.1."
	NodeAnnotaitonNodeProxyIPv4 = OrgName + "/nodeProxyIpv4"
	NodeAnnotaitonNodeProxyIPv6 = OrgName + "/nodeProxyIpv6"

	HostProcMountDir = "/host"

	BpfFSPath    = "/sys/fs/bpf"
	MapsPinpath  = BpfFSPath + "/elf"
	CgroupV2Path = "/run/elf"

	LogLevelEbpfDebug = "verbose"
	LogLevelEbpfInfo  = "info"
	LogLevelEbpfErr   = "error"
)
