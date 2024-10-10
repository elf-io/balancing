package podId

import (
	"github.com/elf-io/balancing/pkg/lock"
)

// PodName 封装 Podname 和 Namespace
type PodName struct {
	Podname   string
	Namespace string
}

// PodID 封装 PodUuid 和 ContainerId
type PodID struct {
	PodUuid     string
	ContainerId string
}

// PodRegistry 是一个存储结构，用于存储和检索 Pod 相关信息
type PodRegistry struct {
	mutex      lock.RWMutex
	keyToValue map[PodName]PodID
	valueToKey map[PodID]PodName
	keyOrder   []PodName // 用于维护键的插入顺序
	capacity   int       // 存储的最大容量
}

// NewPodRegistry 创建并返回一个新的 PodRegistry 实例
func NewPodRegistry(capacity int) *PodRegistry {
	return &PodRegistry{
		keyToValue: make(map[PodName]PodID),
		valueToKey: make(map[PodID]PodName),
		keyOrder:   make([]PodName, 0, capacity),
		capacity:   capacity,
	}
}

// Set 设置 PodName 对应的 PodID 值
func (pr *PodRegistry) Set(key PodName, value PodID) {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	_, exists := pr.keyToValue[key]
	if exists {
		// 如果键已存在，直接更新值
		oldValue := pr.keyToValue[key]
		delete(pr.valueToKey, oldValue) // 删除旧的 value-key 映射
		pr.keyToValue[key] = value
		pr.valueToKey[value] = key
		// 更新键在 keyOrder 中的位置
		pr.removeFromKeyOrder(key)
		pr.keyOrder = append(pr.keyOrder, key)
	} else {
		// 如果是新键，检查是否达到容量上限
		if len(pr.keyToValue) >= pr.capacity {
			// 删除最旧的键值对
			oldestKey := pr.keyOrder[0]
			pr.deleteInternal(oldestKey)
		}
		// 添加新的键值对
		pr.keyToValue[key] = value
		pr.valueToKey[value] = key
		pr.keyOrder = append(pr.keyOrder, key)
	}
}

// Delete 删除与 PodName 对应的条目
func (pr *PodRegistry) Delete(key PodName) {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	pr.deleteInternal(key)
}

// deleteInternal 内部使用的删除方法，不加锁
func (pr *PodRegistry) deleteInternal(key PodName) {
	value, exists := pr.keyToValue[key]
	if !exists {
		// 如果键不存在，直接返回，不做任何操作
		return
	}

	// 删除 keyToValue 中的条目
	delete(pr.keyToValue, key)

	// 删除 valueToKey 中的条目
	delete(pr.valueToKey, value)

	// 从 keyOrder 中移除键
	pr.removeFromKeyOrder(key)
}

// removeFromKeyOrder 从 keyOrder 切片中移除指定的键
func (pr *PodRegistry) removeFromKeyOrder(key PodName) {
	for i, k := range pr.keyOrder {
		if k == key {
			// 使用 copy 来移除元素，避免内存泄漏
			copy(pr.keyOrder[i:], pr.keyOrder[i+1:])
			pr.keyOrder = pr.keyOrder[:len(pr.keyOrder)-1]
			break
		}
	}
}

// GetValueByKey 根据 PodName 查询 PodID
func (pr *PodRegistry) GetValueByKey(key PodName) (PodID, bool) {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	value, exists := pr.keyToValue[key]
	return value, exists
}

// GetKeyByValue 根据 PodID 查询 PodName
func (pr *PodRegistry) GetKeyByValue(value PodID) (PodName, bool) {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	key, exists := pr.valueToKey[value]
	return key, exists
}

// Count 返回存储的键值对数量
func (pr *PodRegistry) Count() int {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	return len(pr.keyToValue)
}

// GetAll 返回所有存储的键值对
func (pr *PodRegistry) GetAll() map[PodName]PodID {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()

	result := make(map[PodName]PodID, len(pr.keyToValue))
	for k, v := range pr.keyToValue {
		result[k] = v
	}
	return result
}
