package podLabel

/*
本文件定义了 PodLabelManager 接口及其实现，用于管理 Kubernetes Pod 的标签和 IP 信息。

主要功能和原理：

1. 接口定义：
   - PodLabelManager 接口提供了管理 Pod 信息的方法，包括更新 Pod 信息和根据标签选择器获取 IP 地址。

2. 接口方法：
   - UpdatePodInfo：更新或删除 Pod 信息，确保存储中的数据与当前 Pod 状态一致。
   - GetLocalIPWithLabelSelector：根据标签选择器获取匹配的本地 IP 地址。
   - GetGlobalIPWithLabelSelector：根据标签选择器获取匹配的全局 IP 地址。

3. 实现细节：
   - 使用 podLabelManager 结构体实现 PodLabelManager 接口。
   - 使用 PodStore 存储 Pod 信息，并提供线程安全的操作。
   - 通过反射检查 Pod 标签和 IP 地址的变化，确保数据的准确性。

4. 使用场景：
   - 适用于需要动态管理和查询 Kubernetes Pod 标签和 IP 信息的场景。
   - 可用于网络管理、监控和调试等场景。

注意事项：
- 所有公共方法都是并发安全的。
- 确保在初始化时调用 InitPodLabelManager 以创建 PodLabelManager 实例。
*/

import (
	"reflect"

	"github.com/elf-io/balancing/pkg/lock"
	"github.com/elf-io/balancing/pkg/utils"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodLabelManager interface {
	UpdatePodInfo(oldPod *corev1.Pod, newPod *corev1.Pod) bool
	GetLocalIPWithLabelSelector(selector *metav1.LabelSelector) []IpInfo
	GetGlobalIPWithLabelSelector(selector *metav1.LabelSelector) []IpInfo
}

type podLabelManager struct {
	store *PodStore
	mu    lock.RWMutex
	l     *zap.Logger
}

var PodLabelHandle PodLabelManager

func InitPodLabelManager(l *zap.Logger) {
	if _, ok := PodLabelHandle.(*podLabelManager); !ok {
		PodLabelHandle = &podLabelManager{
			store: NewPodStore(),
			l:     l,
		}
		l.Sugar().Infof("InitPodLabelManager")
	}
}

func (m *podLabelManager) UpdatePodInfo(oldPod *corev1.Pod, newPod *corev1.Pod) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if newPod == nil && oldPod == nil {
		return false
	}

	if newPod == nil && oldPod != nil {
		// 删除数据
		m.l.Sugar().Debugf("delete pod %s/%s ", oldPod.Namespace, oldPod.Name)
		m.store.DeletePod(oldPod.Namespace, oldPod.Name)
		return true
	}

	if newPod != nil {
		// 检查 Pod 是否处于 Running 状态且没有被标记为删除
		if newPod.Status.Phase != corev1.PodRunning || newPod.DeletionTimestamp != nil {
			// 仅当 oldPod 的状态与 newPod 不同时才删除
			if oldPod != nil && (oldPod.Status.Phase != newPod.Status.Phase || oldPod.DeletionTimestamp == nil) {
				m.l.Sugar().Debugf("delete pod %s/%s due to non-running state or deletion timestamp", oldPod.Namespace, oldPod.Name)
				m.store.DeletePod(oldPod.Namespace, oldPod.Name)
				return true
			}
			return false
		}

		// 检查数据是否变化
		if oldPod == nil || !reflect.DeepEqual(oldPod.Labels, newPod.Labels) || oldPod.Status.PodIP != newPod.Status.PodIP || oldPod.Spec.NodeName != newPod.Spec.NodeName {

			// 提取新 Pod 的 IP 地址
			var ipv4, ipv6 string
			for _, podIP := range newPod.Status.PodIPs {
				if ipv4 == "" && utils.CheckIPv4Format(podIP.IP) {
					ipv4 = podIP.IP
				} else if ipv6 == "" && !utils.CheckIPv4Format(podIP.IP) {
					ipv6 = podIP.IP
				}
				if ipv4 != "" && ipv6 != "" {
					break
				}
			}

			m.l.Sugar().Debugf("update pod %s/%s, ip4: %v, ip6: %v, node: %s ", newPod.Namespace, newPod.Name, ipv4, ipv6, newPod.Spec.NodeName)
			m.store.UpdatePod(newPod.Namespace, newPod.Name, newPod.Labels, ipv4, ipv6, newPod.Spec.NodeName)
			return true
		}
	}

	return false
}

func (m *podLabelManager) GetGlobalIPWithLabelSelector(selector *metav1.LabelSelector) []IpInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.store.FindGlobalIPsByLabelSelector(selector)
}

func (m *podLabelManager) GetLocalIPWithLabelSelector(selector *metav1.LabelSelector) []IpInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.store.FindLocalIPsByLabelSelector(selector)
}
