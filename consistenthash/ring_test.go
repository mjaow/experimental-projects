package consistenthash

import (
	"fmt"
	"testing"
)

func TestLookup(t *testing.T) {
	hash := hashKey("key")
	for _, tc := range []struct {
		name   string
		points []point
		want   string
	}{
		{"empty", nil, ""},
		{"single", []point{{"a", 0}}, "a"},
		{"exact", []point{{"a", hash}, {"b", hash + 1}}, "a"},
		{"between", []point{{"a", hash - 1}, {"b", hash + 1}}, "b"},
		{"before first", []point{{"a", hash + 1}, {"b", hash + 2}}, "a"},
		{"wrap", []point{{"a", 0}, {"b", hash - 1}}, "a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &Ring{points: tc.points}
			if got := r.Lookup("key"); got != tc.want {
				t.Fatalf("Lookup(key) = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMembershipChanges(t *testing.T) {
	r := New(100)
	for _, name := range []string{"node-c", "node-a", "node-b", "node-a"} {
		r.Add(name)
	}
	if len(r.points) != 300 {
		t.Fatalf("got %d points, want 300 including duplicate add", len(r.points))
	}
	for _, name := range []string{"node-a", "node-b", "node-c"} {
		for i := 0; i < 100; i++ {
			if r.Lookup(fmt.Sprintf("%s#%d", name, i)) != name {
				t.Fatalf("virtual node %s#%d does not resolve to its physical node", name, i)
			}
		}
	}
	before := make(map[string]string)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		before[key] = r.Lookup(key)
	}
	r.Add("node-d")
	moved := 0
	for key, owner := range before {
		if got := r.Lookup(key); got != owner {
			if got != "node-d" {
				t.Fatalf("adding node-d moved %s from %s to %s", key, owner, got)
			}
			moved++
		}
	}
	if moved == 0 || moved == len(before) {
		t.Fatalf("expected some keys to move, got %d", moved)
	}
	r.Remove("node-d")
	r.Remove("missing")
	if len(r.points) != 300 {
		t.Fatalf("got %d points after removal, want 300", len(r.points))
	}
	for key, owner := range before {
		if got := r.Lookup(key); got != owner {
			t.Fatalf("removing node-d mapped %s to %s, want %s", key, got, owner)
		}
	}
	for _, name := range []string{"node-a", "node-b", "node-c"} {
		r.Remove(name)
	}
	if r.Lookup("key") != "" {
		t.Fatal("expected an empty ring after removing all nodes")
	}
}

func TestNodeOrderDoesNotMatter(t *testing.T) {
	a, b := New(100), New(100)
	for _, name := range []string{"node-a", "node-b", "node-c"} {
		a.Add(name)
	}
	for _, name := range []string{"node-c", "node-b", "node-a"} {
		b.Add(name)
	}
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		if a.Lookup(key) != b.Lookup(key) {
			t.Fatalf("routers disagree on owner for %s", key)
		}
	}
}
