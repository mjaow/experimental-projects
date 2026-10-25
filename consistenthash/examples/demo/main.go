package main

import (
	"fmt"

	"github.com/mjaow/experimental-projects/consistenthash"
)

func main() {
	for _, virtualNodes := range []int{1, 100} {
		r := consistenthash.New(virtualNodes)
		for _, name := range []string{"node-a", "node-b", "node-c"} {
			r.Add(name)
		}
		before := make([]string, 10000)
		counts := map[string]int{}
		for i := range before {
			before[i] = r.Lookup(fmt.Sprintf("key-%d", i))
			counts[before[i]]++
		}
		fmt.Printf("\n%d virtual nodes per server\nBefore: %v\n", virtualNodes, counts)
		r.Add("node-d")
		counts = map[string]int{}
		moved := 0
		for i, owner := range before {
			next := r.Lookup(fmt.Sprintf("key-%d", i))
			counts[next]++
			if next != owner {
				moved++
			}
		}
		fmt.Printf("After:  %v\nKeys moved when adding node-d: %d / %d\n", counts, moved, len(before))
	}
}
