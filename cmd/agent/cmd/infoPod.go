package cmd

import (
	"github.com/elf-io/balancing/pkg/ebpfWriter"
	"github.com/elf-io/balancing/pkg/podId"
	"github.com/elf-io/balancing/pkg/podLabel"
	"github.com/elf-io/balancing/pkg/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"reflect"
)

// -----------------------------------
type PodReconciler struct {
	log    *zap.Logger
	writer ebpfWriter.EbpfWriter
}

func (s *PodReconciler) HandlerAdd(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		s.log.Sugar().Warnf("HandlerAdd failed to get pod obj: %v")
		return
	}
	name := pod.Namespace + "/" + pod.Name
	logger := s.log.With(
		zap.String("pod", name),
	)

	logger.Sugar().Debugf("process HandlerAdd for pod %s/%s", pod.Namespace, pod.Name)

	// for ebpf event
	if pod.Spec.NodeName == types.AgentConfig.LocalNodeName {
		podId.PodIdHander.Update(nil, pod)
	}

	if changed := podLabel.PodLabelHandle.UpdatePodInfo(nil, pod); changed {
		// inform the balancing policy
		s.writer.UpdateBalancingByPod(logger, pod)

		// inform the localRedirect poliy
		if pod.Spec.NodeName == types.AgentConfig.LocalNodeName {
			// data changed, try to update the ebpf data
			s.writer.UpdateRedirectByPod(logger, pod)
		}
	}

	return
}

func (s *PodReconciler) HandlerUpdate(oldObj, newObj interface{}) {
	oldPod, ok1 := oldObj.(*corev1.Pod)
	if !ok1 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get old pod obj %v")
		return
	}
	newPod, ok2 := newObj.(*corev1.Pod)
	if !ok2 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get new pod obj %v")
		return
	}
	name := newPod.Namespace + "/" + newPod.Name
	logger := s.log.With(
		zap.String("pod", name),
	)

	logger.Sugar().Debugf("process HandlerUpdate for pod %s/%s", newPod.Namespace, newPod.Name)

	// for ebpf event
	if newPod.Spec.NodeName == types.AgentConfig.LocalNodeName {
		if !reflect.DeepEqual(oldPod.Status.ContainerStatuses, newPod.Status.ContainerStatuses) {
			logger.Sugar().Debugf("update podId for pod %s/%s", newPod.Namespace, newPod.Name)
			podId.PodIdHander.Update(oldPod, newPod)
		}
	}

	if changed := podLabel.PodLabelHandle.UpdatePodInfo(oldPod, newPod); changed {
		// inform the balancing policy
		s.writer.UpdateBalancingByPod(logger, newPod)

		// inform the localRedirect poliy
		if newPod.Spec.NodeName == types.AgentConfig.LocalNodeName {
			// data changed, try to update the ebpf data
			s.writer.UpdateRedirectByPod(logger, newPod)
		}
	}

	return
}

func (s *PodReconciler) HandlerDelete(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		s.log.Sugar().Warnf("HandlerDelete failed to get pod obj: %v")
		return
	}
	name := pod.Namespace + "/" + pod.Name
	logger := s.log.With(
		zap.String("pod", name),
	)

	logger.Sugar().Debugf("process HandlerDelete for pod %s/%s", pod.Namespace, pod.Name)

	if pod.Spec.NodeName == types.AgentConfig.LocalNodeName {
		podId.PodIdHander.Update(pod, nil)
	}

	if changed := podLabel.PodLabelHandle.UpdatePodInfo(pod, nil); changed {
		// inform the balancing policy
		s.writer.DeleteBalancingByPod(logger, pod)

		// inform the localRedirect poliy
		if pod.Spec.NodeName == types.AgentConfig.LocalNodeName {
			// data changed, try to update the ebpf data
			s.writer.DeleteRedirectByPod(logger, pod)
		}
	}

	return
}

func NewPodInformer(Client *kubernetes.Clientset, stopWatchCh chan struct{}, writer ebpfWriter.EbpfWriter) {

	// call HandlerUpdate at an interval of 60s
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(Client, InformerListInvterval)
	res := corev1.SchemeGroupVersion.WithResource("pods")
	info, e3 := kubeInformerFactory.ForResource(res)
	if e3 != nil {
		rootLogger.Sugar().Fatalf("failed to create pod informer %v", e3)
	}

	r := PodReconciler{
		log:    rootLogger.Named("PodReconciler"),
		writer: writer,
	}
	info.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.HandlerAdd,
		UpdateFunc: r.HandlerUpdate,
		DeleteFunc: r.HandlerDelete,
	})

	// notice that there is no need to run Start methods in a separate goroutine.
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopWatchCh)

	if !cache.WaitForCacheSync(stopWatchCh, info.Informer().HasSynced) {
		rootLogger.Sugar().Fatalf("failed to WaitForCacheSync for pod ")
	}

	rootLogger.Sugar().Infof("succeeded to NewPodInformer, begin to only watch local pod ")
}
