package opennebula

import (
	"fmt"
	"log"
	"sync"
)

type ResourceKey struct {
	Type string
	ID   int
}

func (k *ResourceKey) String() string {
	return fmt.Sprintf("%s_%d", k.Type, k.ID)
}

type SubResourceKey struct {
	Type    string
	ID      int
	SubType string
}

func (k *SubResourceKey) String() string {
	return fmt.Sprintf("%s_%d_%s", k.Type, k.ID, k.SubType)
}

// MutexKV is a simple key/value store for arbitrary mutexes. It can be used to
// serialize changes across arbitrary collaborators that share knowledge of the
// keys they must serialize on.
//
// The initial use case is to let aws_security_group_rule resources serialize
// their access to individual security groups based on SG ID.
type MutexKV struct {
	lock  sync.Mutex
	store map[string]*sync.RWMutex
}

// Locks the mutex for the given key. Caller is responsible for calling Unlock
// for the same key
func (m *MutexKV) Lock(key fmt.Stringer) {
	keyStr := key.String()
	log.Printf("[DEBUG] Locking %q", key)
	m.get(keyStr).Lock()
	log.Printf("[DEBUG] Locked %q", key)
}

// Unlock the mutex for the given key. Caller must have called Lock for the same key first
func (m *MutexKV) Unlock(key fmt.Stringer) {
	keyStr := key.String()
	log.Printf("[DEBUG] Unlocking %q", key)
	m.get(keyStr).Unlock()
	log.Printf("[DEBUG] Unlocked %q", key)
}

// Rlock
func (m *MutexKV) RLock(key fmt.Stringer) {
	keyStr := key.String()
	log.Printf("[DEBUG] Unlocking %q", key)
	m.get(keyStr).RLock()
	log.Printf("[DEBUG] Unlocked %q", key)
}

// RUnlock
func (m *MutexKV) RUnlock(key fmt.Stringer) {
	keyStr := key.String()
	log.Printf("[DEBUG] Unlocking %q", key)
	m.get(keyStr).RUnlock()
	log.Printf("[DEBUG] Unlocked %q", key)
}

// Returns a mutex for the given key, no guarantee of its lock status
func (m *MutexKV) get(key string) *sync.RWMutex {
	m.lock.Lock()
	defer m.lock.Unlock()
	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.RWMutex{}
		m.store[key] = mutex
	}
	return mutex
}

// Returns a properly initialized MutexKV
func NewMutexKV() *MutexKV {
	return &MutexKV{
		store: make(map[string]*sync.RWMutex),
	}
}
