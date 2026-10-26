// Package lamport implements scalar logical clocks for causal ordering.
package lamport

// Clock belongs to one process. Its zero value is ready to use.
// Calls must be serialized; counters must not overflow.
type Clock struct {
	time uint64
}

// Tick records a local or send event. Attach the result to outgoing messages.
func (c *Clock) Tick() uint64 {
	c.time++
	return c.time
}

// Receive records receipt of a message carrying the sender's timestamp.
func (c *Clock) Receive(remote uint64) uint64 {
	if remote > c.time {
		c.time = remote
	}
	return c.Tick()
}
