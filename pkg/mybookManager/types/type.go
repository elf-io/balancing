// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package types

type MybookManager interface {
	RunWebhookServer(webhookPort int, tlsDir string)
	RunController(leaseName, leaseNameSpace string, leaseId string)
}
