package gb32960

import "sync"

type vinRegistry struct {
	mu   sync.RWMutex
	vins map[string][]*Connection
}

func newVinRegistry() *vinRegistry {
	return &vinRegistry{
		vins: make(map[string][]*Connection),
	}
}

func (r *vinRegistry) add(vin string, c *Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.vins[vin] = append(r.vins[vin], c)
}

func (r *vinRegistry) remove(vin string, c *Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()

	conns, ok := r.vins[vin]
	if !ok {
		return
	}
	for i, conn := range conns {
		if conn == c {
			r.vins[vin] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	if len(r.vins[vin]) == 0 {
		delete(r.vins, vin)
	}
}

func (r *vinRegistry) get(vin string) *Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conns, ok := r.vins[vin]
	if !ok || len(conns) == 0 {
		return nil
	}
	return conns[0]
}

func (r *vinRegistry) getAll(vin string) []*Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Connection, 0)
	if conns, ok := r.vins[vin]; ok {
		result = append(result, conns...)
	}
	return result
}

func (r *vinRegistry) count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.vins)
}
