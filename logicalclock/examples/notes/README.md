# Comparing clocks on shared-note edits

This application imports both independent packages and gives them the same events. It uses a Lamport timestamp plus author ID to select one stored note, while vector timestamps expose concurrency. Neither clock implementation depends on the other.

From `logicalclock/`:

```sh
go run ./examples/notes
go test ./examples/notes -v
```

Expected output:

```text
1. Messages carry causal history
A edits "Draft": L=1, V=map[A:1]
B receives Draft: L=2
B edits "Reviewed": L=3, V=map[A:1 B:2]
A receives Reviewed: L=4
Draft vs Reviewed: before

2. Disconnected replicas make concurrent edits
A edits "Add diagram": L=5, V=map[A:3 B:2]
B edits "Add tests": L=4, V=map[A:1 B:3]
Lamport: B's edit (4) < A's edit (5), but neither edit saw the other.
Vector comparison: concurrent

3. Exchange edits using larger (Lamport timestamp, author ID) as the winner
A receives B's edit: L=6
B receives A's edit: L=6
A stores "Add diagram"; B stores "Add diagram"
Both converge, but B's concurrent edit is lost. Vector clocks can flag the conflict; an application must resolve it.
```

`edit` models creating and publishing an update as one event. `receive` records a separate receipt event, while the incoming edit keeps its original timestamp. `newer` implements the application's last-writer-wins policy by comparing `(Lamport timestamp, author ID)`.

The final replicas agree because they received the same edits and use the same deterministic selection rule. That convergence does not preserve both edits: B's concurrent work is lost from the stored note. The vector comparison detects this conflict but does not change the example's storage policy.

Try these experiments:

1. Remove the draft/review exchange. Each first edit gets Lamport time 1; the author-ID tie-break chooses the winner while vectors show concurrency.
2. Deliver A's edit to B before B edits. B's new edit now follows A's in both clock systems.
3. Reverse delivery order or repeat an update. The stored winner is unchanged, but each receipt advances the receiving clock.
4. Replace `newer` with a policy retaining concurrent revisions, then model a user merging them. The independent `examples/vector` example demonstrates the clock operations for that merge.
