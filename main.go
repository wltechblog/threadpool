// Package threadqueue provides a thread Queue implementation for managing
// concurrent goroutines in Go applications. It allows limiting the number
// of simultaneously running goroutines to prevent resource exhaustion.
package threadqueue

import (
	"sync"
)

// ThreadQueue represents a Queue of worker goroutines with a maximum concurrency limit.
// It manages the scheduling of goroutines to ensure the concurrency limit is not exceeded.
type ThreadQueue struct {
	// max is the maximum number of goroutines allowed to run concurrently
	max int
	// current tracks the number of goroutines currently in the Queue
	current int
	// mutex protects access to the current count and waiting queue
	mutex sync.Mutex
	// cond is used to signal waiting goroutines
	cond *sync.Cond
	// In is the channel for receiving join requests from goroutines (kept for compatibility)
	In chan chan bool
	// waiting is kept for compatibility with tests
	waiting interface{}
}

// New creates and initializes a new ThreadQueue with the specified maximum concurrency.
// It returns a pointer to the created ThreadQueue.
//
// Parameters:
//   - max: The maximum number of goroutines allowed to run concurrently in the Queue
//
// Returns:
//   - *ThreadQueue: A pointer to the initialized ThreadQueue
func New(max int) *ThreadQueue {
	if max < 1 {
		max = 1 // Ensure at least one goroutine can run
	}
	
	Queue := &ThreadQueue{
		max:     max,
		current: 0,
		In:      make(chan chan bool, max), // Kept for compatibility
	}
	
	Queue.cond = sync.NewCond(&Queue.mutex)
	
	return Queue
}

// Join blocks until the calling goroutine can enter the thread Queue.
// If the Queue is at capacity, the goroutine will wait until another goroutine leaves.
// This method should be called before performing work that needs to be constrained by the Queue.
func (p *ThreadQueue) Join() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Wait until we're below max capacity
	for p.current >= p.max {
		p.cond.Wait()
	}
	
	// We can now enter the Queue
	p.current++
}

// Leave notifies the Queue that the calling goroutine has completed its work and is
// exiting the Queue. This frees up a slot for another waiting goroutine.
// This method should be called after the constrained work is completed, typically
// using defer to ensure it's called even if the work panics.
func (p *ThreadQueue) Leave() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Decrement the current count
	if p.current > 0 {
		p.current--
	}
	
	// Signal one waiting goroutine that a slot is available
	p.cond.Signal()
}
