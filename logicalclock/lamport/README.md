# Lamport clocks

A scalar logical clock named after Leslie Lamport. Import `github.com/mjaow/experimental-projects/logicalclock/lamport`; this package has no dependency on the vector-clock package.

## Rules and usage

Each process owns one counter `L`, initially zero:

- Local event or send: `L = L + 1`. Attach the result to outgoing messages.
- Receive a message stamped `T`: `L = max(L, T) + 1`.

```go
var checkout, payment lamport.Clock
sent := checkout.Tick()           // 1: send a payment request.
received := payment.Receive(sent) // 2: record receipt before handling it.
charged := payment.Tick()         // 3: record the charge and publish the result.
_, _ = received, charged
```

The zero value is ready to use. `Receive` already counts the receipt event. A later application action gets its own `Tick`. If editing and sending are modeled as separate events, tick for each; if they are a single publish event, tick once.

## What it solves

If `a` happens before `b`, then `L(a) < L(b)`. Local increments preserve process order, while receiving moves beyond both the sender's time and the receiver's current time. This works without synchronized wall clocks.

A log collector can sort a collected batch by `(counter, process ID)`. The process ID breaks ties deterministically. The result respects causality, even if records arrive out of order. A live collector still needs a delivery protocol to know whether earlier records are missing.

## Run the log example

From `logicalclock/`:

```sh
go run ./examples/lamport
go test ./lamport -v
```

The example passes timestamps between checkout and payment services, collects their logs out of order, and sorts them. Expected output:

```text
L=1 checkout: order created
L=2 checkout: sent payment request
L=3 payment: received payment request
L=4 payment: charged card
L=5 checkout: order confirmed
```

## What the number does not tell you

`L(a) < L(b)` does **not** imply `a -> b`. Two disconnected processes can produce timestamps 4 and 5 without either seeing the other's event. Equal timestamps across processes imply concurrency in this model; unequal timestamps may also be concurrent.

A process-ID tie-break chooses an order for independent events; it does not establish a causal relationship. Likewise, selecting the greatest timestamp as a stored-value winner can discard concurrent work. Logical "last" does not necessarily mean physically latest.

Read the sibling [vector-clock guide](../vector/README.md) and run `go run ./examples/vector` to see how richer timestamps identify concurrent revisions and let an application retain both for merging.

Tests cover local increments, receipts from senders behind/equal/ahead, later local events, transitive causal chains, and independent events with equal or unequal timestamps. The clock needs serialized access and assumes counters do not overflow or reset during the modeled lifetime.
