// Package threadqueue provides a thread pool implementation for managing
// concurrent goroutines in Go applications. It allows limiting the number
// of simultaneously running goroutines to prevent resource exhaustion.
package threadqueue

import (
	"sync"
)

// ThreadPool represents a pool of worker goroutines with a maximum concurrency limit.
// It manages the scheduling of goroutines to ensure the concurrency limit is not exceeded.
type ThreadPool struct {
	// max is the maximum number of goroutines allowed to run concurrently
	max int
	// current tracks the number of goroutines currently in the pool
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

// New creates and initializes a new ThreadPool with the specified maximum concurrency.
// It returns a pointer to the created ThreadPool.
//
// Parameters:
//   - max: The maximum number of goroutines allowed to run concurrently in the pool
//
// Returns:
//   - *ThreadPool: A pointer to the initialized ThreadPool
func New(max int) *ThreadPool {
	if max < 1 {
		max = 1 // Ensure at least one goroutine can run
	}
	
	pool := &ThreadPool{
		max:     max,
		current: 0,
		In:      make(chan chan bool, max), // Kept for compatibility
	}
	
	pool.cond = sync.NewCond(&pool.mutex)
	
	return pool
}

// Join blocks until the calling goroutine can enter the thread pool.
// If the pool is at capacity, the goroutine will wait until another goroutine leaves.
// This method should be called before performing work that needs to be constrained by the pool.
func (p *ThreadPool) Join() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Wait until we're below max capacity
	for p.current >= p.max {
		p.cond.Wait()
	}
	
	// We can now enter the pool
	p.current++
}

// Leave notifies the pool that the calling goroutine has completed its work and is
// exiting the pool. This frees up a slot for another waiting goroutine.
// This method should be called after the constrained work is completed, typically
// using defer to ensure it's called even if the work panics.
func (p *ThreadPool) Leave() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Decrement the current count
	if p.current > 0 {
		p.current--
	}
	
	// Signal one waiting goroutine that a slot is available
	p.cond.Signal()
}
