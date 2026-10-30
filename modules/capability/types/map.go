package types

import "sync"

// CapabilityMap is the in-memory index -> capability map shared by the keeper
// and every scoped keeper. It lives outside the branched store, so it is safe
// for concurrent use and keeps one pointer per index across store branches.
type CapabilityMap struct {
	sync.RWMutex
	caps map[uint64]*Capability
}

func NewCapabilityMap() *CapabilityMap {
	return &CapabilityMap{caps: make(map[uint64]*Capability)}
}

func (c *CapabilityMap) Get(index uint64) (*Capability, bool) {
	c.RLock()
	defer c.RUnlock()
	cap, ok := c.caps[index]
	return cap, ok
}

// GetOrSet returns the capability already stored at index, storing and
// returning cap only if the index is empty. Capabilities are identified by
// pointer, so every store branch that creates the same index must observe
// the same pointer.
func (c *CapabilityMap) GetOrSet(index uint64, cap *Capability) *Capability {
	c.Lock()
	defer c.Unlock()
	if existing, ok := c.caps[index]; ok {
		return existing
	}
	c.caps[index] = cap
	return cap
}

func (c *CapabilityMap) Set(index uint64, cap *Capability) {
	c.Lock()
	defer c.Unlock()
	c.caps[index] = cap
}

func (c *CapabilityMap) Delete(index uint64) {
	c.Lock()
	defer c.Unlock()
	delete(c.caps, index)
}
