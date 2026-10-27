# Logical clocks

Two independent Go implementations for learning distributed event ordering: **Lamport clocks** and **vector clocks**. Both belong to the logical-clock family. Neither measures seconds or elapsed time. No external dependencies.

## Organization

```text
logicalclock/
  go.mod                  One module for this learning project
  lamport/                Scalar clock, focused tests, and its own guide
  vector/                 Vector clock, focused tests, and its own guide
  examples/lamport/       Order checkout/payment logs using only Lamport clocks
  examples/vector/        Detect revision conflicts using only vector clocks
  causality_test.go       Check both algorithms against an explicit message history
```

The packages are peers: neither imports the other. There is no shared clock interface because a scalar timestamp and a vector expose different information and comparisons. An application imports whichever package it needs.

## Run and learn

From the repository root:

```sh
cd logicalclock
go run ./examples/lamport
go run ./examples/vector
go test ./...
```

Read the [Lamport guide](lamport/README.md), then the [vector guide](vector/README.md). Each example uses a fixed sequence of in-memory message deliveries so the clock rules stay visible without networking or scheduling noise.

## The shared problem: causality without synchronized time

A sends a message at wall-clock time 10:00:05; B's slow clock records receipt at 10:00:02. Sorting by physical timestamps can put the effect before its cause. Independent local counters also fail unless messages carry ordering information across process boundaries.

Both algorithms track **happened-before**, written `a -> b`:

1. Earlier events in the same process happen before later events.
2. Sending a message happens before receiving it.
3. The relation is transitive: `a -> b` and `b -> c` imply `a -> c`.

Events are **concurrent** if neither happens before the other. They need not execute at the same wall-clock instant.

## Their different guarantees

**Lamport:** one counter per process; `a -> b` implies `L(a) < L(b)`. The reverse is false. Adding a process ID breaks ties to produce a deterministic total order, but it arbitrarily orders concurrent events. Timestamps have constant size.

**Vector:** one counter per tracked process; with complete event tracking, `a -> b` if and only if `V(a) < V(b)`. Vector comparison is componentwise, not lexicographic. Incomparable vectors identify concurrency. Metadata grows with the number of tracked participants.

For example, A performs two events and B performs one, with no communication. Their latest Lamport timestamps are 2 and 1, but the events are concurrent. Their vector timestamps `[A:2, B:0]` and `[A:0, B:1]` reveal that neither has seen the other's history.

Vector clocks retain more information at higher cost. They are a different logical-clock algorithm, not a component of Lamport clocks or a universal replacement.

For concurrent writes, choosing the largest `(Lamport timestamp, process ID)` gives replicas with the same writes a deterministic winner, but can discard an independent edit. Vector clocks distinguish a causally superseding version from a concurrent version, allowing the application to retain both for merging. Use the ordering rule when a single deterministic winner is sufficient; use causal metadata when concurrent versions need different handling. An application does not generally need both clocks.

## Test each layer

From this directory:

```sh
go test ./lamport -v
go test ./vector -v
go test . -v
go test ./examples/...
go vet ./...
```

Package tests check each algorithm independently: local events, receipt, causal propagation, comparisons, and vector snapshot isolation. `causality_test.go` builds an independent ancestor graph from process order and message edges, then checks every pair of events against both clocks. Its history includes independent edits, forks, joins, delayed messages, and duplicates.

Both focused examples have executable expected-output tests. The shared causality test also verifies that concurrent events can have unequal Lamport timestamps while their vectors remain incomparable.

## Beyond clocks

- **Hybrid logical clocks (HLC)** combine physical time and a logical counter, retaining causal ordering with compact timestamps close to wall-clock time under suitable assumptions. They cannot generally detect concurrency like vectors.
- **CRDTs** define data types and merge rules for concurrent updates, such as preserving independent counter increments. Conflict detection alone does not supply a merge policy.
- **Consensus, such as Raft,** agrees on committed operations despite failures. Clock labels alone do not provide that agreement.

Both implementations assume serialized calls, complete tracking of modeled messages, no counter overflow, and no restarts. Vector clocks and Lamport tie-breaking additionally require unique, stable process IDs. Real systems need transport, persistence, identity management, and delivery protocols as appropriate.
