package main

import (
	"fmt"

	"github.com/mjaow/experimental-projects/logicalclock/lamport"
	"github.com/mjaow/experimental-projects/logicalclock/vector"
)

type change struct {
	text    string
	author  string
	lamport uint64
	vector  vector.Timestamp
}

type replica struct {
	id      string
	clock   lamport.Clock
	vector  *vector.Clock
	current change
}

func newReplica(id string) *replica {
	return &replica{id: id, vector: vector.New(id)}
}

func (r *replica) edit(text string) change {
	// Creating and publishing an edit is one event in this model.
	r.current = change{text, r.id, r.clock.Tick(), r.vector.Tick()}
	return r.current
}

func (r *replica) receive(incoming change) uint64 {
	receivedAt := r.clock.Receive(incoming.lamport)
	r.vector.Receive(incoming.vector)
	// Keep the original edit timestamp; receipt is a separate event.
	if newer(incoming, r.current) {
		r.current = incoming
	}
	return receivedAt
}

func newer(a, b change) bool {
	return a.lamport > b.lamport || (a.lamport == b.lamport && a.author > b.author)
}

func printEdit(c change) {
	fmt.Printf("%s edits %q: L=%d, V=%v\n", c.author, c.text, c.lamport, c.vector)
}

func main() {
	a, b := newReplica("A"), newReplica("B")

	fmt.Println("1. Messages carry causal history")
	draft := a.edit("Draft")
	printEdit(draft)
	fmt.Printf("B receives Draft: L=%d\n", b.receive(draft))
	reviewed := b.edit("Reviewed")
	printEdit(reviewed)
	fmt.Printf("A receives Reviewed: L=%d\n", a.receive(reviewed))
	fmt.Printf("Draft vs Reviewed: %s\n", draft.vector.Compare(reviewed.vector))

	fmt.Println("\n2. Disconnected replicas make concurrent edits")
	aEdit := a.edit("Add diagram")
	bEdit := b.edit("Add tests")
	printEdit(aEdit)
	printEdit(bEdit)
	fmt.Printf("Lamport: B's edit (%d) < A's edit (%d), but neither edit saw the other.\n", bEdit.lamport, aEdit.lamport)
	fmt.Printf("Vector comparison: %s\n", aEdit.vector.Compare(bEdit.vector))

	fmt.Println("\n3. Exchange edits using larger (Lamport timestamp, author ID) as the winner")
	fmt.Printf("A receives B's edit: L=%d\n", a.receive(bEdit))
	fmt.Printf("B receives A's edit: L=%d\n", b.receive(aEdit))
	fmt.Printf("A stores %q; B stores %q\n", a.current.text, b.current.text)
	fmt.Println("Both converge, but B's concurrent edit is lost. Vector clocks can flag the conflict; an application must resolve it.")
}
