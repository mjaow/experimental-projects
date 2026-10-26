package vector_test

import (
	"maps"
	"testing"

	"github.com/mjaow/experimental-projects/logicalclock/vector"
)

func TestCompare(t *testing.T) {
	reverse := map[vector.Relation]vector.Relation{
		vector.Before: vector.After, vector.After: vector.Before,
		vector.Equal: vector.Equal, vector.Concurrent: vector.Concurrent,
	}
	for _, tc := range []struct {
		name string
		a, b vector.Timestamp
		want vector.Relation
	}{
		{"before", vector.Timestamp{"A": 1}, vector.Timestamp{"A": 1, "B": 2}, vector.Before},
		{"after", vector.Timestamp{"A": 2, "B": 1}, vector.Timestamp{"A": 1}, vector.After},
		{"equal", vector.Timestamp{"A": 1}, vector.Timestamp{"A": 1}, vector.Equal},
		{"missing means zero", nil, vector.Timestamp{"A": 0}, vector.Equal},
		{"independent", vector.Timestamp{"A": 1}, vector.Timestamp{"B": 1}, vector.Concurrent},
		{"shared history", vector.Timestamp{"A": 3, "B": 2}, vector.Timestamp{"A": 1, "B": 3}, vector.Concurrent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.Compare(tc.b); got != tc.want {
				t.Fatalf("%v compared to %v = %s, want %s", tc.a, tc.b, got, tc.want)
			}
			if got := tc.b.Compare(tc.a); got != reverse[tc.want] {
				t.Fatalf("reverse comparison = %s, want %s", got, reverse[tc.want])
			}
		})
	}
}

func TestReceiveMergesEveryComponent(t *testing.T) {
	b := vector.New("B")
	b.Tick()
	b.Tick()
	for _, tc := range []struct {
		remote, want vector.Timestamp
	}{
		{vector.Timestamp{"A": 4, "C": 1}, vector.Timestamp{"A": 4, "B": 3, "C": 1}},
		{vector.Timestamp{"A": 2, "C": 5}, vector.Timestamp{"A": 4, "B": 4, "C": 5}},
		{vector.Timestamp{"A": 2, "C": 5}, vector.Timestamp{"A": 4, "B": 5, "C": 5}},
	} {
		if got := b.Receive(tc.remote); !maps.Equal(got, tc.want) {
			t.Fatalf("Receive(%v) = %v, want %v", tc.remote, got, tc.want)
		}
	}
}

func TestNewRejectsEmptyID(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected an empty process ID to be rejected")
		}
	}()
	vector.New("")
}

func TestReceivePreservesHistory(t *testing.T) {
	a, b, c := vector.New("A"), vector.New("B"), vector.New("C")
	sent := a.Tick()
	b.Tick()
	b.Tick()
	received := b.Receive(sent)
	if want := (vector.Timestamp{"A": 1, "B": 3}); !maps.Equal(received, want) {
		t.Fatalf("receipt = %v, want %v", received, want)
	}
	forwarded := b.Tick()
	delivered := c.Receive(forwarded)
	if sent.Compare(received) != vector.Before || received.Compare(forwarded) != vector.Before || sent.Compare(delivered) != vector.Before {
		t.Fatal("messages did not preserve transitive causal history")
	}
	if want := (vector.Timestamp{"A": 1, "B": 5}); !maps.Equal(b.Receive(sent), want) {
		t.Fatalf("delayed message must preserve B's history and count another receipt")
	}
}

func TestTimestampsAreSnapshots(t *testing.T) {
	a, b := vector.New("A"), vector.New("B")
	sent := a.Tick()
	a.Tick()
	received := b.Receive(sent)
	if !maps.Equal(sent, vector.Timestamp{"A": 1}) {
		t.Fatalf("later events changed the sent message: %v", sent)
	}
	sent["A"] = 99
	received["B"] = 99
	if next := a.Tick(); next["A"] != 3 {
		t.Fatalf("caller changed A's internal clock: %v", next)
	}
	if next := b.Tick(); !maps.Equal(next, vector.Timestamp{"A": 1, "B": 2}) {
		t.Fatalf("message or returned snapshot changed B's internal clock: %v", next)
	}
}
