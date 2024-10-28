// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0
package podId

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

/*
  存储数据 podName -> pod containerId
  可根据进程的 pid  -》 查询到 containerId -》 查询到对应 哪个 pod
*/

type PodIdManager interface {
	Update(*corev1.Pod, *corev1.Pod)
	LookupPodByPid(uint32) (string, string, string, string, bool, error)
}

type podIdManager struct {
	client   *kubernetes.Clientset
	log      *zap.Logger
	NodeName string
	// key : podId and containerID
	// values: pod name and namespace
	podInfo *PodRegistry
}

var _ PodIdManager = (*podIdManager)(nil)

var PodIdHander PodIdManager

func InitPodIdManager(c *kubernetes.Clientset, log *zap.Logger, nodeName string) {
	if _, ok := PodIdHander.(*podIdManager); !ok {
		t := &podIdManager{
			client:   c,
			log:      log,
			NodeName: nodeName,
			// each node running pod with a total of max 1000
			podInfo: NewPodRegistry(1000),
		}
		t.initPodId()
		PodIdHander = t
		log.Sugar().Info("finish initialize PodBankHander")
	} else {
		log.Sugar().Errorf("secondary calling for PodBankHander")
	}
}

// -----------------------------------

func (s *podIdManager) updatePodInfo(pod *corev1.Pod) error {
	getContaineridFunc := func(line string) string {
		// ContainerID is the ID of the container in the format '<type>://<container_id>'.
		index := strings.Index(line, "//")
		if index == -1 {
			return ""
		}
		return strings.TrimSpace(line[index+2:])
	}

	if len(pod.Status.ContainerStatuses) > 0 {
		l := len(pod.Status.ContainerStatuses)
		containerId := getContaineridFunc(pod.Status.ContainerStatuses[l-1].ContainerID)
		if len(containerId) == 0 {
			return fmt.Errorf("failed to get container id")
		}
		key := PodName{
			Podname:   pod.Name,
			Namespace: pod.Namespace,
		}
		value := PodID{
			PodUuid:     string(pod.ObjectMeta.UID),
			ContainerId: containerId,
		}
		s.podInfo.Set(key, value)
		return nil
	}
	return fmt.Errorf("no ContainerStatuses")
}

func (s *podIdManager) deletePodInfo(pod *corev1.Pod) {
	key := PodName{
		Podname:   pod.Name,
		Namespace: pod.Namespace,
	}
	s.podInfo.Delete(key)
}

// -----------------------------------

// before pod informer, build local database firstly for serving ebpf in case of missing event
func (s *podIdManager) initPodId() {

	s.log.Sugar().Infof("initPodBank")

	namespaces, err := s.client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		s.log.Sugar().Fatalf("failed to list namespaces: %v", err)
	}
	for _, ns := range namespaces.Items {
		// only get the pod of local node
		pods, err := s.client.CoreV1().Pods(ns.Name).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", s.NodeName),
		})
		if err != nil {
			s.log.Sugar().Fatalf("Error listing pods in namespace %s: %v\n", ns.Name, err)
		}
		for _, pod := range pods.Items {
			s.log.Sugar().Debugf("save uid information of pod %s/%s", pod.Namespace, pod.Name)
			if err := s.updatePodInfo(&pod); err != nil {
				s.log.Sugar().Errorf("error: %s", err)
			}
		}
	}

	s.log.Sugar().Infof("succeeded to get all pod uid information, total %d: %+v", s.podInfo.Count(), s.podInfo.GetAll())
}

func (s *podIdManager) Update(oldPod, newPod *corev1.Pod) {
	if newPod == nil && oldPod == nil {
		return

	} else if newPod == nil && oldPod != nil {
		// delete
		s.log.Sugar().Debugf("delete pod id for pod %s/%s", oldPod.Namespace, oldPod.Name)
		s.deletePodInfo(oldPod)
		s.log.Sugar().Debugf("pod uid information, total %d: %+v", s.podInfo.Count(), s.podInfo.GetAll())
	} else {
		// add
		if len(newPod.Status.ContainerStatuses) > 0 {
			s.log.Sugar().Debugf("add pod id for pod %s/%s", newPod.Namespace, newPod.Name)
			if err := s.updatePodInfo(newPod); err != nil {
				s.log.Sugar().Errorf("error: %s", err)
			}
			s.log.Sugar().Debugf("pod uid information, total %d: %+v", s.podInfo.Count(), s.podInfo.GetAll())
		}
	}
}

// pid 用于查询 关联 pod name
// 如果是 k8s pod，则 podName, namespace, containerdId 有值
// 如果只是个 容器 但不是 pod，则   containerdId 有值
// 如果只是个 主机上的应用，则 bool 有值
func (s *podIdManager) LookupPodByPid(pid uint32) (podName, namespace, containerdId, podUuid string, hostApp bool, err error) {
	hostApp = false
	if pid == 0 {
		return "", "", "", "", false, fmt.Errorf("empty input")
	}

	// get pod information from cgroup
	podId, containerId, host, e := getPodAndContainerID(pid)
	if e != nil {
		err = fmt.Errorf("failed to getPodAndContainerID for pid %d: %v ", pid, e)
		return
	}
	s.log.Sugar().Debugf("process pid %d got: podUuid=%s, containerId=%s, host=%v", pid, podId, containerId, host)
	if host {
		// this is a host application
		return "", "", "", "", true, nil
	}

	if len(podId) == 0 && len(containerId) > 0 {
		return "", "", containerId, podId, false, nil
	}

	// for a k8s pod, get its pod name
	value := PodID{
		PodUuid:     podId,
		ContainerId: containerId,
	}
	if k, ok := s.podInfo.GetKeyByValue(value); ok {
		return k.Podname, k.Namespace, containerId, podId, false, nil
	}

	err = fmt.Errorf("no data of PodName for pid %d", pid)
	return
}
