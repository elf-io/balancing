package podLabel

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
	GetIPWithLabelSelector(selector *metav1.LabelSelector) []IpInfo
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
		if oldPod == nil || !reflect.DeepEqual(oldPod.Labels, newPod.Labels) || oldPod.Status.PodIP != newPod.Status.PodIP {

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

			m.l.Sugar().Debugf("update pod %s/%s, ip4: %v, ip6: %v ", newPod.Namespace, newPod.Name, ipv4, ipv6)
			m.store.UpdatePod(newPod.Namespace, newPod.Name, newPod.Labels, ipv4, ipv6)
			return true
		}
	}

	return false
}

func (m *podLabelManager) GetIPWithLabelSelector(selector *metav1.LabelSelector) []IpInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.store.FindIPsByLabelSelector(selector)
}
