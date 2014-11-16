package base

import (
	"sync"
)

// This type can be used as the backing storage for VariableAttributes.
type PackedVariableStorageGroup struct {
	m sync.Mutex // Sync
	v [][]byte   // Storage
}

func NewPackedVariableStorageGroup() *PackedVariableStorageGroup {
	return &PackedVariableStorageGroup{
		sync.Mutex{},
		make([][]byte, 0),
	}
}

// Allocate allocates and/or returns size bytes for permanent storage.
func (p *PackedVariableStorageGroup) Allocate(size int) (int, []byte) {
	p.m.Lock()
	defer p.m.Unlock()

	ret := make([]byte, size)
	p.v = append(p.v, ret)
	return len(p.v) - 1, ret
}

//
func (p *PackedVariableStorageGroup) Retrieve(off int) []byte {
	return p.v[off]
}
