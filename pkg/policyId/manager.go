// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package policyId

import (
	"fmt"
	"github.com/elf-io/balancing/pkg/lock"
	"github.com/elf-io/balancing/pkg/utils"
)

type PolicyIdManager interface {
	GeneratePolicyId(string) (string, error)
	SavePolicyId(string, string) error
}

type policyIdManager struct {
	// node id
	dataLock *lock.Mutex
	idData   map[string]uint32
}

var _ PolicyIdManager = (*policyIdManager)(nil)

var PolicyIdManagerHandler PolicyIdManager

func init() {
	PolicyIdManagerHandler = &policyIdManager{
		dataLock: &lock.Mutex{},
		idData:   make(map[string]uint32),
	}
}

func (s *policyIdManager) GeneratePolicyId(policyName string) (string, error) {

	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	if id, ok := s.idData[policyName]; ok {
		return fmt.Sprintf("%d", id), fmt.Errorf("policy already has a id=%d", id)
	}

	id := uint32(0)
	for {
		id = utils.RandomUint32()
		conflict := false
		for _, v := range s.idData {
			if v == id {
				conflict = true
				break
			}
		}
		if conflict {
			continue
		}
		break
	}
	s.idData[policyName] = id

	return fmt.Sprintf("%d", id), nil
}

func (s *policyIdManager) SavePolicyId(policyName string, idStr string) error {

	newId, e := utils.StringToUint32(idStr)
	if e != nil {
		return fmt.Errorf("id is invalid, it is must be a string for uint32 ")
	}

	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	if id, ok := s.idData[policyName]; ok {
		if newId != id {
			return fmt.Errorf("policy %s already has another id=%d", policyName, id)
		}
	} else {
		s.idData[policyName] = newId
	}

	return nil
}
