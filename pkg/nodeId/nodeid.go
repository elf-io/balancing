package nodeId

import (
	"context"
	"fmt"
	"github.com/elf-io/balancing/pkg/lock"
	"github.com/elf-io/balancing/pkg/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

/*
   each node has a persistent ID
*/

*/
type NodeIdManager interface {
	GetNodeId(string) (uint32, error)
	BuildNodeId(*corev1.Node) error
	DeleteNodeId(string)
}

type nodeIdManager struct {
	nodeIdData map[string]uint32
	client     *kubernetes.Clientset
	dataLock   *lock.Mutex
	log        *zap.Logger
}

// var NodeIdManagerHander NodeIdManager = (*nodeIdManager)(nil)

var _ NodeIdManager = (*nodeIdManager)(nil)

var NodeIdManagerHander NodeIdManager

// used to generate and store nodeIp for each node
// when ebpf applies some endpoints data, they need to use nodeId, but the node resource possibly has not been synchronized,
// so it introduce an abstraction layer to store and search dynamically from api-server
func InitNodeIdManager(c *kubernetes.Clientset, log *zap.Logger) {
	if _, ok := NodeIdManagerHander.(*nodeIdManager); !ok {
		t := &nodeIdManager{
			client:     c,
			nodeIdData: make(map[string]uint32),
			dataLock:   &lock.Mutex{},
			log:        log,
		}
		t.initNodeId()
		NodeIdManagerHander = t
		log.Sugar().Info("finish initialize NodeIdManagerHander")
	} else {
		log.Sugar().Errorf("secondary calling for InitNodeIdManager")
	}
}

func (s *nodeIdManager) applyNewNodeIP(oldNode corev1.Node) (uint32, error) {
	node := &oldNode
	if v, ok := node.ObjectMeta.Annotations[types.NodeAnnotationNodeIdKey]; ok {
		return stringToUint32(v)
	}

	for count := 1; count < 1000; count++ {
		nodeId := generateRandomUint32()
		t := Uint32ToString(nodeId)
		node.ObjectMeta.Annotations[types.NodeAnnotationNodeIdKey] = t

		s.log.Sugar().Info("node %s lacks nodeId, try to apply a new nodeId %s for ", node.Name, t)
		if _, err := s.client.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{}); err != nil {
			if apierrors.IsConflict(err) {
				// look whether pod has been updated with the nodeId by other pod
				node, err = s.client.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
				if err != nil {
					s.log.Sugar().Errorf("failed to get node %s: %v", node.Name, err)
				} else {
					if nodeIdStr, ok := node.ObjectMeta.Annotations[types.NodeAnnotationNodeIdKey]; ok {
						r, err1 := stringToUint32(nodeIdStr)
						if err1 != nil {
							s.log.Sugar().Errorf("resourceVersion conflicted,  node %s got an invalid nodeId %s: %v", node.Name, nodeIdStr, err1)
							// try to generate an new valid one
						} else {
							s.log.Sugar().Info("resourceVersion conflicted,  node %s got another nodeId %s", node.Name, nodeIdStr)
							return r, nil
						}
					} else {
						s.log.Sugar().Info("resourceVersion conflicted for node %s ", node.Name)
					}
				}
			} else {
				s.log.Sugar().Errorf("failed to set nodeIp to node %s: %v", node.Name, err)
			}
		} else {
			s.log.Sugar().Infof("succeeded to apply a new nodeId %s for node %s ", t, node.Name)
			return nodeId, nil
		}
		// sleep and retry
		time.Sleep(time.Duration(count*100) * time.Millisecond)
	}
	return 0, fmt.Errorf("failed to apply nodeID for node %s", node.Name)
}

// generate nodeId and set to the node's annotation , and build all local database
func (s *nodeIdManager) initNodeId() {

	s.log.Sugar().Infof("initial nodeId")

	nodeList, err := s.client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.log.Sugar().Fatalf("failed to list node: %v", err)
	}

	nodeIdMap := map[string]int{}
	missNodeIdList := []corev1.Node{}

	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	// build all nodeId first
	for _, node := range nodeList.Items {
		if nodeIdStr, ok := node.ObjectMeta.Annotations[types.NodeAnnotationNodeIdKey]; ok {
			if t, err1 := stringToUint32(nodeIdStr); err1 != nil {
				s.log.Sugar().Errorf("found an invalid nodeId %s for node %s: %v", nodeIdStr, node.Name, err1)
				missNodeIdList = append(missNodeIdList, node)
			} else {
				s.nodeIdData[node.Name] = t
				nodeIdMap[nodeIdStr] = 0
			}
		} else {
			missNodeIdList = append(missNodeIdList, node)
		}
	}

	// process the node missing nodeId
	for _, node := range missNodeIdList {
		nodeId, err := s.applyNewNodeIP(node)
		if err != nil {
			s.log.Sugar().Fatalf("failed to set nodeId for node %s : %v", node.Name, err)
		}
		s.nodeIdData[node.Name] = nodeId
	}

	s.log.Sugar().Infof("succeeded to get all nodeId: %+v", s.nodeIdData)
	return
}

func (s *nodeIdManager) BuildNodeId(node *corev1.Node) error {
	if node == nil {
		return fmt.Errorf("empty node obj ")
	}

	s.dataLock.Lock()
	if _, ok := s.nodeIdData[node.Name]; ok {
		s.dataLock.Unlock()
		return nil
	}
	s.dataLock.Unlock()

	nodeId, err := s.applyNewNodeIP(*node)
	if err != nil {
		return fmt.Errorf("failed to applyNewNodeIP for node %s : %v", node.Name, err)
	}

	s.dataLock.Lock()
	s.nodeIdData[node.Name] = nodeId
	s.dataLock.Unlock()

	s.log.Sugar().Infof("succeeded to build nodeId %d for node %s", nodeId, node.Name)

	return nil
}

func (s *nodeIdManager) GetNodeId(nodeName string) (uint32, error) {
	if len(nodeName) == 0 {
		s.log.Sugar().Errorf("empty nodeName ")
		return 0, fmt.Errorf("empty nodeName ")
	}

	// for redirect and balancing polity, it does not care about nodeId
	if nodeName == types.NodeNameIgnore {
		return 0xffffffff, nil
	}

	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	if nodeId, ok := s.nodeIdData[nodeName]; ok {
		return nodeId, nil
	}
	s.log.Sugar().Errorf("no node Id for node %s ", nodeName)

	return 0, fmt.Errorf("no node Id")
}

func (s *nodeIdManager) DeleteNodeId(nodeName string) {
	if len(nodeName) == 0 {
		return
	}
	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	delete(s.nodeIdData, nodeName)

	s.log.Sugar().Infof("succeeded to delete nodeId for node %s", nodeName)

	return
}
