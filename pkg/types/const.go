// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package types

const (
	OrgName = "elf.io"

	ApiVersion = "v1beta1"
	ApiGroup   = "balancing." + OrgName

	// use this annotation to mark an ID in the annotation of each node
	// "balancing.elf.org/nodeId": "32BitNumber"
	NodeAnnotationNodeIdKey = ApiGroup + "/nodeId"

	// use this annotation to mark an ID in the annotation of each node
	// "balancing.elf.org/serviceId": "32BitNumber"
	AnnotationServiceID = ApiGroup + "/serviceId"

	// the user could mark the ip in the annotation of each node
	// "bpfElf.org/nodeProxyIpv4": "192.168.1.1."
	NodeAnnotaitonNodeProxyIPv4 = ApiGroup + "/nodeProxyIpv4"
	NodeAnnotaitonNodeProxyIPv6 = ApiGroup + "/nodeProxyIpv6"

	HostProcMountDir = "/host"

	BpfFSPath    = "/sys/fs/bpf"
	MapsPinpath  = BpfFSPath + "/balancing"
	CgroupV2Path = "/run/balancing/cgroupv2"

	LogLevelEbpfDebug = "verbose"
	LogLevelEbpfInfo  = "info"
	LogLevelEbpfErr   = "error"

	NodeNameIgnore  = "ignoreNode"
	NamespaceIgnore = "ignoreNamespace"
)
