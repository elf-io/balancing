// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/debug"
	"github.com/elf-io/balancing/pkg/ebpf"
	"github.com/elf-io/balancing/pkg/ebpfEvent"
	"github.com/elf-io/balancing/pkg/ebpfWriter"
	"github.com/elf-io/balancing/pkg/nodeId"
	"github.com/elf-io/balancing/pkg/podId"
	"github.com/elf-io/balancing/pkg/podLabel"
	"github.com/elf-io/balancing/pkg/types"
	"github.com/elf-io/balancing/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/signal"
	runtimedebug "runtime/debug"
	"syscall"
)

func SetupUtility() {
	// run gops
	d := debug.New(rootLogger)
	if types.AgentConfig.GopsPort != 0 {
		d.RunGops(int(types.AgentConfig.GopsPort))
	}

	if types.AgentConfig.PyroscopeServerAddress != "" {
		d.RunPyroscope(types.AgentConfig.PyroscopeServerAddress, types.AgentConfig.PodName)
	}
}

func DaemonMain() {
	defer func() {
		if e := recover(); nil != e {
			rootLogger.Sugar().Errorf("Panic details: %v", e)
			runtimedebug.PrintStack()
			os.Exit(1)
		}
	}()
	rootLogger.Sugar().Infof("config: %+v", types.AgentConfig)
	SetupUtility()

	// ------------------------------------
	rootLogger.Sugar().Debugf("RunReconciles")
	// get clientset
	apiServerHostAddress := ""
	if len(types.AgentConfig.KubeconfigPath) > 0 {
		rootLogger.Sugar().Infof("out of cluster: set Kubebeconfig to %s", types.AgentConfig.KubeconfigPath)
	} else if len(types.AgentConfig.Configmap.ApiServerHost) > 0 && len(types.AgentConfig.Configmap.ApiServerPort) > 0 {
		apiServerHostAddress = fmt.Sprintf("%s:%s", types.AgentConfig.Configmap.ApiServerHost, types.AgentConfig.Configmap.ApiServerPort)
		rootLogger.Sugar().Infof("in cluster: replace the address of api Server to %s", apiServerHostAddress)
	}
	clientConfig, e1 := utils.AutoK8sConfig(types.AgentConfig.KubeconfigPath, apiServerHostAddress)
	if e1 != nil {
		rootLogger.Sugar().Fatalf("failed to find client-go config, make sure it is in a pod or ~/.kube/config exists: %v", e1)
	}
	rootLogger.Sugar().Debugf("clientConfig: %+v", clientConfig)

	Client, e2 := kubernetes.NewForConfig(clientConfig)
	if e2 != nil {
		rootLogger.Sugar().Fatalf("failed to NewForConfig: %v", e2)
	}

	// before informer and ebpf, build nodeId database
	nodeId.InitNodeIdManager(Client, rootLogger.Named("nodeId"))

	// before informer and ebpf, build pod ip database of local node
	podId.InitPodIdManager(Client, rootLogger.Named("podId"), types.AgentConfig.LocalNodeName)

	podLabel.InitPodLabelManager(rootLogger.Named("podLabel"))

	// setup ebpf and load
	bpfManager := ebpf.NewEbpfProgramMananger(rootLogger.Named("ebpf"))
	if err := bpfManager.LoadProgramp(); err != nil {
		rootLogger.Sugar().Fatalf("failed to Load ebpf Programp: %v \n", err)
	}
	rootLogger.Sugar().Infof("succeeded to Load ebpf Programp \n")
	// setup ebpf writer
	writer := ebpfWriter.NewEbpfWriter(Client, bpfManager, InformerListInvterval, rootLogger.Named("ebpfWriter"))
	// before informer, clean all map data to keep all data up to date
	writer.CleanEbpfMapData()

	// setup informer
	stopWatchCh := make(chan struct{})
	NewPodInformer(Client, stopWatchCh, writer)
	NewNodeInformer(Client, stopWatchCh, writer)

	NewServiceInformer(Client, stopWatchCh, writer)
	NewEndpointSliceInformer(Client, stopWatchCh, writer)

	// crd reconcile
	SetupController(clientConfig, writer)

	//
	ebpfEvent := ebpfEvent.NewEbpfEvent(rootLogger.Named("ebpfEvent"), bpfManager, writer)
	ebpfEvent.WatchEbpfEvent(stopWatchCh)

	rootLogger.Info("finish all setup ")

	finishlock.Lock()
	finishSetUp = true
	finishlock.Unlock()

	// ------------------------------------
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
		for sig := range sigCh {
			rootLogger.Sugar().Warnf("Received signal %+v ", sig)
			rootLogger.Info("unload ebpf program ")
			bpfManager.UnloadProgramp()
			os.Exit(1)
		}
	}()
	select {}

}
