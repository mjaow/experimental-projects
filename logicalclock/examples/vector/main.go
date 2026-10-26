package main

import (
	"fmt"

	"github.com/mjaow/experimental-projects/logicalclock/vector"
)

type revision struct {
	text string
	time vector.Timestamp
}

func main() {
	a, b := vector.New("A"), vector.New("B")
	draft := revision{"Draft", a.Tick()}
	b.Receive(draft.time)

	// Each replica edits without receiving the other's new revision.
	left := revision{"Add diagram", a.Tick()}
	right := revision{"Add tests", b.Tick()}
	fmt.Printf("%q vs %q: %s\n", left.text, right.text, left.time.Compare(right.time))
	if left.time.Compare(right.time) == vector.Concurrent {
		fmt.Println("Keep both revisions for a user to merge.")
	}

	a.Receive(right.time)
	// The application/user supplies the merged text; the clock only labels it.
	merged := revision{"Add diagram and tests", a.Tick()}
	fmt.Printf("%q vs merged: %s\n", left.text, left.time.Compare(merged.time))
	fmt.Printf("%q vs merged: %s\n", right.text, right.time.Compare(merged.time))
	fmt.Printf("Merged revision: %q\n", merged.text)
}
