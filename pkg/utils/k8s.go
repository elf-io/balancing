// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func DeepCopy(src, dst interface{}) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	if err := enc.Encode(src); err != nil {
		return err
	}
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

var (
	DefaultKubeConfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	ScInPodPath           = "/var/run/secrets/kubernetes.io/serviceaccount"
)

// KubeConfigPath is for agent on hosts out of the cluster
// apiServerHostAddress is for agent and controller pod in the cluster when kube-proxy is not running
func AutoK8sConfig(KubeConfigPath, apiServerHostAddress string) (*rest.Config, error) {
	var config *rest.Config
	var err error

	if len(KubeConfigPath) != 0 {
		config, err = clientcmd.BuildConfigFromFlags("", KubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get config from kube config=%v , info=%v", KubeConfigPath, err)
		}
		return config, nil
	}

	if ExistFile(DefaultKubeConfigPath) {
		config, err = clientcmd.BuildConfigFromFlags("", DefaultKubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get config from kube config=%v , info=%v", DefaultKubeConfigPath, err)
		}

	} else if ExistDir(ScInPodPath) {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get config from serviceaccount=%v , info=%v", ScInPodPath, err)
		}
		if len(apiServerHostAddress) > 0 {
			config.Host = apiServerHostAddress
		}
	} else {
		return nil, fmt.Errorf("failed to get config ")
	}

	return config, nil
}

// KubeConfigPath is for agent on hosts out of the cluster
// apiServerHostAddress is for agent and controller pod in the cluster when kube-proxy is not running
func AutoCrdConfig(KubeConfigPath, apiServerHostAddress string) (*rest.Config, error) {
	var config *rest.Config
	var err error

	if len(KubeConfigPath) != 0 {
		config, err = clientcmd.BuildConfigFromFlags("", KubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get config from kube config=%v , info=%v", KubeConfigPath, err)
		}
		return config, nil
	}

	if ExistFile(DefaultKubeConfigPath) {
		config, err = clientcmd.BuildConfigFromFlags("", DefaultKubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get config from kube config=%v , info=%v", DefaultKubeConfigPath, err)
		}

	} else if ExistDir(ScInPodPath) {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get config from serviceaccount=%v , info=%v", ScInPodPath, err)
		}
		if len(apiServerHostAddress) > 0 {
			config.Host = apiServerHostAddress
		}
	} else {
		return nil, fmt.Errorf("failed to get config ")
	}

	return config, nil
}
