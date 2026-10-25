package consistenthash

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
)

type point struct {
	node string
	hash uint64
}

type Ring struct {
	virtualNodes int
	points       []point
}

func New(virtualNodes int) *Ring {
	if virtualNodes < 1 {
		panic("virtualNodes must be positive")
	}
	return &Ring{virtualNodes: virtualNodes}
}

func (r *Ring) Add(name string) {
	if name == "" {
		panic("node name must not be empty")
	}
	for _, p := range r.points {
		if p.node == name {
			return
		}
	}
	for i := 0; i < r.virtualNodes; i++ {
		id := fmt.Sprintf("%s#%d", name, i)
		r.points = append(r.points, point{name, hashKey(id)})
	}
	sort.Slice(r.points, func(i, j int) bool {
		if r.points[i].hash == r.points[j].hash {
			return r.points[i].node < r.points[j].node
		}
		return r.points[i].hash < r.points[j].hash
	})
}

func (r *Ring) Remove(name string) {
	points := r.points[:0]
	for _, p := range r.points {
		if p.node != name {
			points = append(points, p)
		}
	}
	r.points = points
}

func (r *Ring) Lookup(key string) string {
	if len(r.points) == 0 {
		return ""
	}
	hash := hashKey(key)
	i := sort.Search(len(r.points), func(i int) bool {
		return r.points[i].hash >= hash
	})
	if i == len(r.points) {
		i = 0
	}
	return r.points[i].node
}

func hashKey(key string) uint64 {
	sum := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint64(sum[:8])
}
