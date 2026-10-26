# Vector clocks

A logical clock that can distinguish causal ordering from concurrency. Import `github.com/mjaow/experimental-projects/logicalclock/vector`; this package has no dependency on the Lamport-clock package.

## Rules and usage

Each process has a unique ID and a map of known event counts, one entry per process:

- Local event or send: increment your own entry.
- Receive: take the maximum for each entry from the local and received vectors, then increment your own entry.
- Return an independent snapshot for the event/message. Later events must not rewrite earlier history.

```go
a, b := vector.New("A"), vector.New("B")
first, second := a.Tick(), b.Tick()
fmt.Println(first.Compare(second)) // concurrent
receipt := b.Receive(first)
fmt.Println(first.Compare(receipt)) // before
```

`New` rejects an empty ID; the caller must assign unique process IDs. Missing vector entries mean zero. Treat returned timestamps as event records; do not modify or share them for mutation.

## Comparison is componentwise

`a.Compare(b)` returns one of four relations:

- `Before`: every entry of `a` is at most the corresponding entry of `b`, and at least one is smaller.
- `After`: the reverse of `Before`.
- `Equal`: every entry matches, treating missing entries as zero.
- `Concurrent`: `a` is larger on at least one entry and smaller on another.

For example, `[A:2, B:0]` and `[A:1, B:2]` are concurrent. The first has more A history and the second has more B history. Sorting these vectors lexicographically would hide that information.

With complete tracking of events and messages, for distinct events `a -> b` **if and only if** `V(a) < V(b)`. This is stronger than the one-way guarantee of [Lamport clocks](../lamport/README.md).

## Run the revision example

From `logicalclock/`:

```sh
go run ./examples/vector
go test ./vector -v
```

Two replicas first share a draft, then edit independently. The example retains both revisions and models a user merging them after receiving both histories. Expected output:

```text
"Add diagram" vs "Add tests": concurrent
Keep both revisions for a user to merge.
"Add diagram" vs merged: before
"Add tests" vs merged: before
Merged revision: "Add diagram and tests"
```

The merged text is supplied explicitly by the application. **The clock detects the conflict and orders the merge after its inputs; it does not decide how to combine text.** A full multi-version store would also need to retain every unresolved concurrent revision.

## Costs and limits

Timestamps grow with the number of tracked participants, and identities must remain meaningful across the history. These costs matter when clients join, leave, or restart. Systems tracking stored object versions often use the related idea of **version vectors** rather than clocking every process event.

Tests cover all four comparison results in both directions, absent entries, componentwise merging, transitive causality, delayed/duplicate receipt, and snapshot isolation. Calls must be serialized; the example assumes no counter overflow, restarts, or untracked causal communication.
