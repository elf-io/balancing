// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package types

const (
	OrgName = "elf.io"

	ApiVersion = "v1beta"
	ApiGroup   = "balancing." + OrgName

	// use this annotation to mark an ID in the annotation of each node
	// "bpfElf.org/nodeId": "32BitNumber"
	NodeAnnotaitonNodeIdKey = OrgName + "/nodeId"

	// the user could mark the ip in the annotation of each node
	// "bpfElf.org/nodeEntryIpv4": "192.168.1.1."
	NodeAnnotaitonNodeEntryIPv4 = OrgName + "/nodeEntryIpv4"
	NodeAnnotaitonNodeEntryIPv6 = OrgName + "/nodeEntryIpv6"

	HostProcMountDir = "/host"

	BpfFSPath    = "/sys/fs/bpf"
	MapsPinpath  = BpfFSPath + "/elf"
	CgroupV2Path = "/var/run/elf"
)
