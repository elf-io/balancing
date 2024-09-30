// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of elf-io

package utils_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/elf-io/balancing/pkg/utils"
	"net"
)

var _ = Describe("net", func() {

	It("ListHostAllInterfaces", func() {
		v, e := utils.ListHostAllInterfaces()
		Expect(e).NotTo(HaveOccurred())

		GinkgoWriter.Printf("interface list %v \n", v)
		fmt.Printf("interface list %v \n", v)

		result, _ := net.InterfaceByName("en0")
		t, _ := result.Addrs()
		fmt.Printf("ip %+v \n", t)

		ipv4List, ipv6List, e := utils.GetAllInterfaceUnicastAddrWithoutMask()
		Expect(e).NotTo(HaveOccurred(), "failed to get ip, %v", e)
		fmt.Printf("ipv4List %+v \n", ipv4List)
		fmt.Printf("ipv6List %+v \n", ipv6List)

	})

})
