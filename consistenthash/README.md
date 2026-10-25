# Consistent hashing

A small Go library and runnable examples for learning shard routing. No external dependencies.

## Virtual nodes

`New(100)` places each physical node at 100 positions: hash `db-a#0`, `db-a#1`, ..., `db-a#99`. These points divide the ring into smaller intervals, spread across physical nodes. More virtual nodes usually make key distribution more even, at the cost of a larger ring.

`Lookup(key)` hashes the key, finds the first point at or after it, and returns that point's **physical node**. Lookup wraps to the first point at the end of the ring. Both keys and virtual node IDs use the first 64 bits of SHA-256.

Adding a physical node adds all its virtual points; only keys taken by that node move. Removing a node removes all its points; its keys move to the surviving successors. Virtual nodes are ring positions, **not copies of data**.

## Use the library

```go
import "github.com/mjaow/experimental-projects/consistenthash"

ring := consistenthash.New(100)
ring.Add("db-a")
ring.Add("db-b")
ring.Add("db-c")

shard := ring.Lookup("user:42") // Physical node, e.g. "db-b".
// Use the database connection or HTTP client for this shard.

ring.Remove("db-b")
```

Construct rings with `New` and a positive virtual-node count. Node names must be nonempty. Adding the same node again or removing an unknown node has no effect. Lookup returns `""` for an empty ring. Configure membership before serving concurrent lookups; `Add` and `Remove` are not synchronized with readers.

## Compare distribution

From this directory:

```sh
go run ./examples/demo
go test ./...
```

The demo compares 1 and 100 virtual nodes per server across 10,000 keys. It prints per-server counts and how many keys move when a fourth server joins. With many well-distributed points, each of four equally sized servers should own roughly a quarter of the keys; the example measures the actual result.

## Build a service: sharded key-value storage

`examples/kv` runs in two modes. A storage shard keeps a local map. A router imports this library, maps the key to a shard URL, and forwards PUT and GET requests to that shard. The router stores no values. `X-Shard` in the response shows which physical shard served the request.

Run each command in a separate terminal, from this directory:

```sh
go run ./examples/kv -listen localhost:9001
go run ./examples/kv -listen localhost:9002
go run ./examples/kv -listen localhost:9003
go run ./examples/kv -listen localhost:8080 -shards http://localhost:9001,http://localhost:9002,http://localhost:9003
```

In PowerShell, send requests through the router:

```powershell
curl.exe -i -X PUT http://localhost:8080/kv/user:42 --data "Alice"
curl.exe -i http://localhost:8080/kv/user:42
curl.exe -i -X PUT http://localhost:8080/kv/user:43 --data "Bob"
curl.exe -i http://localhost:8080/kv/user:43
```

PUT returns 204; GET returns the stored value, or 404 if absent. Both requests for a key reach the same physical shard. Storage is in memory and disappears when the shard stops.

For a database team, the same pattern is `user ID -> ring.Lookup -> shard connection pool -> SQL query`. Use a stable partition key, such as the user ID, for all records that need to live together. A cache client can instead map a cache key to a cache server and refill misses from its source database. Each client or router must use the same node IDs, hash algorithm, and virtual-node count to agree on ownership.

### What happens when a shard joins?

Start another storage shard on port 9004, then restart only the router with all four shard URLs. Some existing keys will now map to the empty new shard and return 404. This demonstrates an essential boundary: **the ring changes routing; it does not move stored data**. A database service must migrate affected records and coordinate routing changes. A cache can often refill those keys on a miss. This example keeps membership fixed while the router runs so that this distinction stays visible.

Read `ring.go` for the algorithm, `examples/demo/main.go` for balance and movement, then `newRouter` in `examples/kv/main.go` for service integration.
