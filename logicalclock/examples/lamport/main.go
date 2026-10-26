package main

import (
	"fmt"
	"sort"

	"github.com/mjaow/experimental-projects/logicalclock/lamport"
)

type entry struct {
	service string
	action  string
	time    uint64
}

func main() {
	var checkout, payment lamport.Clock
	created := checkout.Tick()
	requested := checkout.Tick()
	received := payment.Receive(requested)
	charged := payment.Tick()
	confirmed := checkout.Receive(charged)

	// A log collector can receive records in a different order from the events.
	logs := []entry{
		{"payment", "charged card", charged},
		{"checkout", "order confirmed", confirmed},
		{"checkout", "order created", created},
		{"payment", "received payment request", received},
		{"checkout", "sent payment request", requested},
	}
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].time == logs[j].time {
			return logs[i].service < logs[j].service
		}
		return logs[i].time < logs[j].time
	})
	for _, log := range logs {
		fmt.Printf("L=%d %s: %s\n", log.time, log.service, log.action)
	}
}
