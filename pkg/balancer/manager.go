package balancer

import (
	"context"
	"log"
	"sync/atomic"
	"time"
)

// manageSubConnections initiates two goroutines to handle the pick results.
// One goroutine enqueues and tracks pick results while the other handles dequeued pick results.
func (b *consistentHashBalancer) manageSubConnections() {
	// Goroutine to listen for pick results and enqueue them for further processing.
	go func() {
		for {
			pr := <-b.pickerResultChannel
			b.enqueueAndTrackPickResult(pr)
		}
	}()

	// Goroutine to process and handle dequeued pick results.
	go func() {
		for {
			v, ok := b.activePickResults.DeQueue()
			if !ok {
				time.Sleep(ConnectionLifetime)
				continue
			}
			pr := v.(PickResult)
			b.handleDequeuedPickResult(pr)
		}
	}()
}

// enqueueAndTrackPickResult enqueues a given pick result and tracks its state.
// It also enqueues a shadow pick result with a limited lifespan.
func (b *consistentHashBalancer) enqueueAndTrackPickResult(pr PickResult) {
	// Enqueue the original pick result.
	b.activePickResults.EnQueue(pr)

	// Create a shadow context with a predefined lifespan.
	shadowCtx, cancel := context.WithTimeout(context.Background(), ConnectionLifetime)
	defer cancel()

	// Enqueue the shadow pick result.
	b.activePickResults.EnQueue(PickResult{Ctx: shadowCtx, SC: pr.SC})

	// Update the sub-connection counts in a thread-safe manner.
	b.balancerLock.Lock()
	cnt, ok := b.subConnPickCounts[pr.SC]
	if !ok {
		cnt = new(int32)
		b.subConnPickCounts[pr.SC] = cnt
	}
	*cnt += 2
	b.balancerLock.Unlock()
}

// handleDequeuedPickResult processes a dequeued pick result.
// Depending on the state of the context and the sub-connection, various actions are taken.
func (b *consistentHashBalancer) handleDequeuedPickResult(pr PickResult) {
	select {
	// If the context associated with the pick result is done...
	case <-pr.Ctx.Done():
		b.balancerLock.Lock()
		defer b.balancerLock.Unlock()

		// If the sub-connection is in a certain status, re-enqueue the pick result.
		if b.subConnStatusMap[pr.SC] {
			b.activePickResults.EnQueue(pr)
			return
		}

		// Decrease the count for the sub-connection.
		cnt, ok := b.subConnPickCounts[pr.SC]
		if !ok {
			return
		}

		atomic.AddInt32(cnt, -1)
		// If count becomes zero, reset the sub-connection.
		if *cnt == 0 {
			delete(b.subConnPickCounts, pr.SC)

			b.subConnStatusMap[pr.SC] = true
			if err := b.resetSubConn(pr.SC); err != nil {
				log.Printf("Failed to reset SubConn: %v", err)
			}
			delete(b.subConnStatusMap, pr.SC)
		}
	// If the context isn't done yet, re-enqueue the pick result.
	default:
		b.activePickResults.EnQueue(pr)
	}
}
