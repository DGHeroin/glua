package core

import (
    "sync"
    "sync/atomic"
)

type Registry struct {
    mu       sync.RWMutex
    registry map[uint32]interface{}
    index    uint32
}

func NewRegistry() *Registry {
    r := &Registry{
        registry: map[uint32]interface{}{},
    }
    return r
}

func (r *Registry) Get(id uint32) interface{} {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.registry[id]
}

func (r *Registry) Put(p interface{}) uint32 {
    r.mu.Lock()
    defer r.mu.Unlock()
    id := uint32(0)
    for {
        id = atomic.AddUint32(&r.index, 1)
        if id == 0 {
            continue
        }
        if _, ok := r.registry[id]; !ok {
            break
        }
    }
    r.registry[id] = p
    return id
}

func (r *Registry) Del(id uint32) {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.registry, id)
}
