package main

import (
	"testing"

	"github.com/mjaow/experimental-projects/logicalclock/vector"
)

func TestCausalEditWinsDespiteAuthorOrder(t *testing.T) {
	a, b := newReplica("A"), newReplica("B")
	first := b.edit("Draft")
	a.receive(first)
	second := a.edit("Reviewed")
	b.receive(second)
	if first.vector.Compare(second.vector) != vector.Before {
		t.Fatal("review must follow the draft")
	}
	if a.current.text != "Reviewed" || b.current.text != "Reviewed" {
		t.Fatal("causally later edit did not win on both replicas")
	}
}

func TestConcurrentEditsWithUnequalLamportTimes(t *testing.T) {
	a, b := newReplica("A"), newReplica("B")
	b.receive(a.edit("Draft"))
	a.receive(b.edit("Reviewed"))
	aEdit, bEdit := a.edit("Add diagram"), b.edit("Add tests")
	if bEdit.lamport >= aEdit.lamport || aEdit.vector.Compare(bEdit.vector) != vector.Concurrent {
		t.Fatal("expected unequal Lamport timestamps for concurrent edits")
	}
	a.receive(bEdit)
	b.receive(aEdit)
	if a.current.text != "Add diagram" || b.current.text != "Add diagram" {
		t.Fatal("replicas did not converge to the same winner")
	}
}

func TestTieBreakAndDeliveryOrder(t *testing.T) {
	a, b := newReplica("A"), newReplica("B")
	aEdit, bEdit := a.edit("A's edit"), b.edit("B's edit")
	if aEdit.lamport != bEdit.lamport || aEdit.vector.Compare(bEdit.vector) != vector.Concurrent {
		t.Fatal("independent first edits should tie but be concurrent")
	}
	for _, delivery := range [][]change{{aEdit, bEdit}, {bEdit, aEdit}} {
		r := newReplica("reader")
		for _, edit := range delivery {
			r.receive(edit)
		}
		r.receive(aEdit)
		if r.current.author != "B" || r.current.lamport != bEdit.lamport {
			t.Fatal("arrival order or a duplicate changed the winning edit")
		}
		if next := r.edit("Merged"); next.lamport <= bEdit.lamport || bEdit.vector.Compare(next.vector) != vector.Before {
			t.Fatal("a new edit must follow the edits already received")
		}
	}
}
