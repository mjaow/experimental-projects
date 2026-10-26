// Package vector implements vector clocks for causality and concurrency detection.
package vector

import "maps"

// Timestamp records known event counts by process ID. Missing entries mean zero.
type Timestamp map[string]uint64

type Relation string

const (
	Before     Relation = "before"
	Equal      Relation = "equal"
	After      Relation = "after"
	Concurrent Relation = "concurrent"
)

// Clock belongs to one process with a unique ID. Calls must be serialized.
type Clock struct {
	node string
	time Timestamp
}

func New(node string) *Clock {
	if node == "" {
		panic("node ID must not be empty")
	}
	return &Clock{node: node, time: Timestamp{}}
}

// Tick returns an independent snapshot so later events cannot change a message.
func (c *Clock) Tick() Timestamp {
	c.time[c.node]++
	return maps.Clone(c.time)
}

func (c *Clock) Receive(remote Timestamp) Timestamp {
	for node, time := range remote {
		if time > c.time[node] {
			c.time[node] = time
		}
	}
	return c.Tick()
}

func (t Timestamp) Compare(other Timestamp) Relation {
	var less, greater bool
	for node, time := range t {
		greater = greater || time > other[node]
	}
	for node, time := range other {
		less = less || t[node] < time
	}
	switch {
	case less && greater:
		return Concurrent
	case less:
		return Before
	case greater:
		return After
	default:
		return Equal
	}
}
