package logicalclock_test

import (
	"testing"

	"github.com/mjaow/experimental-projects/logicalclock/lamport"
	"github.com/mjaow/experimental-projects/logicalclock/vector"
)

func TestClocksAgainstMessageHistory(t *testing.T) {
	// A nonnegative from identifies the earlier event whose message is received.
	history := []struct {
		name string
		node int
		from int
	}{
		{"A publishes draft", 0, -1},
		{"B publishes independent work", 1, -1},
		{"A publishes revision", 0, -1},
		{"B receives draft", 1, 0},
		{"C receives revision", 2, 2},
		{"C publishes review", 2, -1},
		{"A receives review", 0, 5},
		{"B publishes independent edit", 1, -1},
		{"A receives B's edit", 0, 7},
		{"A publishes merged revision", 0, -1},
		{"C receives delayed B message", 2, 1},
		{"C receives duplicate B message", 2, 1},
		{"B receives A's joined history", 1, 9},
	}
	scalars := [3]lamport.Clock{}
	vectors := []*vector.Clock{vector.New("A"), vector.New("B"), vector.New("C")}
	previous := []int{-1, -1, -1}
	times := make([]uint64, len(history))
	stamps := make([]vector.Timestamp, len(history))
	ancestors := make([]map[int]bool, len(history))
	for i, event := range history {
		// Build causality only from process order and message edges, not clocks.
		ancestors[i] = map[int]bool{}
		for _, parent := range []int{previous[event.node], event.from} {
			if parent >= 0 {
				ancestors[i][parent] = true
				for ancestor := range ancestors[parent] {
					ancestors[i][ancestor] = true
				}
			}
		}
		previous[event.node] = i
		if event.from < 0 {
			times[i] = scalars[event.node].Tick()
			stamps[i] = vectors[event.node].Tick()
		} else {
			times[i] = scalars[event.node].Receive(times[event.from])
			stamps[i] = vectors[event.node].Receive(stamps[event.from])
		}
	}

	var unequalConcurrent bool
	for i, a := range history {
		for j, b := range history {
			want := vector.Concurrent
			switch {
			case i == j:
				want = vector.Equal
			case ancestors[j][i]:
				want = vector.Before
				if times[i] >= times[j] {
					t.Errorf("%s -> %s, but Lamport %d >= %d", a.name, b.name, times[i], times[j])
				}
			case ancestors[i][j]:
				want = vector.After
			default:
				unequalConcurrent = unequalConcurrent || times[i] != times[j]
			}
			if got := stamps[i].Compare(stamps[j]); got != want {
				t.Errorf("%s vs %s: vector says %s, message history says %s", a.name, b.name, got, want)
			}
		}
	}
	if !unequalConcurrent {
		t.Fatal("history must demonstrate concurrent events with unequal Lamport timestamps")
	}
}
