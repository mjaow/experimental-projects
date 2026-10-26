package lamport_test

import (
	"testing"

	"github.com/mjaow/experimental-projects/logicalclock/lamport"
)

func TestCausalChain(t *testing.T) {
	var a, b, c lamport.Clock
	sent := a.Tick()
	received := b.Receive(sent)
	forwarded := b.Tick()
	delivered := c.Receive(forwarded)
	if !(sent < received && received < forwarded && forwarded < delivered) {
		t.Fatalf("causal chain is out of order: %d, %d, %d, %d", sent, received, forwarded, delivered)
	}
}

func TestLocalEvents(t *testing.T) {
	var clock lamport.Clock
	for want := uint64(1); want <= 3; want++ {
		if got := clock.Tick(); got != want {
			t.Fatalf("local event = %d, want %d", got, want)
		}
	}
}

func TestReceive(t *testing.T) {
	for _, tc := range []struct {
		name         string
		remote, want uint64
	}{
		{"sender behind", 1, 4},
		{"sender equal", 3, 4},
		{"sender ahead", 8, 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var clock lamport.Clock
			for i := 0; i < 3; i++ {
				clock.Tick()
			}
			if got := clock.Receive(tc.remote); got != tc.want {
				t.Fatalf("Receive(%d) = %d, want %d", tc.remote, got, tc.want)
			}
			if got := clock.Tick(); got != tc.want+1 {
				t.Fatalf("local event after receipt = %d, want %d", got, tc.want+1)
			}
		})
	}
}

func TestIndependentEvents(t *testing.T) {
	var a, b lamport.Clock
	firstA, firstB := a.Tick(), b.Tick()
	if firstA != firstB {
		t.Fatal("independent processes can produce equal timestamps")
	}
	if secondA := a.Tick(); secondA <= firstB {
		t.Fatal("independent processes can also produce unequal timestamps")
	}
}
