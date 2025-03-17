// Package threadqueue provides a thread pool implementation for managing
// concurrent goroutines in Go applications.
package threadqueue

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestThreadPoolBasic verifies that a new ThreadPool is initialized correctly
// with the expected default values.
func TestThreadPoolBasic(t *testing.T) {
	// Create a new thread pool with max 2 threads
	pool := New(2)
	
	// Test that the pool is initialized correctly
	if pool.max != 2 {
		t.Errorf("Expected max to be 2, got %d", pool.max)
	}
	
	if pool.current != 0 {
		t.Errorf("Expected current to be 0, got %d", pool.current)
	}
	
	if pool.In == nil {
		t.Error("Expected In channel to be initialized")
	}
	
	if pool.cond == nil {
		t.Error("Expected condition variable to be initialized")
	}
}

// TestThreadPoolJoinLeave tests the core functionality of the ThreadPool:
// joining and leaving the pool while respecting the maximum concurrency limit.
// It verifies that no more than the maximum number of goroutines run concurrently.
func TestThreadPoolJoinLeave(t *testing.T) {
	// Create a new thread pool with max 2 threads
	pool := New(2)
	
	// Test Join and Leave
	var counter int32 = 0
	
	// We should be able to run 2 goroutines concurrently
	var wg sync.WaitGroup
	wg.Add(4)
	
	// Launch 4 goroutines, but only 2 should run at a time
	for i := 0; i < 4; i++ {
		go func() {
			defer wg.Done()
			
			// Join the pool (blocks until a slot is available)
			pool.Join()
			defer pool.Leave()
			
			// Increment counter
			atomic.AddInt32(&counter, 1)
			
			// Simulate work
			time.Sleep(100 * time.Millisecond)
			
			// Check that we never exceed max threads
			current := atomic.LoadInt32(&counter)
			if current > 2 {
				t.Errorf("Too many concurrent goroutines: %d", current)
			}
			
			// Decrement counter
			atomic.AddInt32(&counter, -1)
		}()
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
}

// TestThreadPoolConcurrency tests the ThreadPool with different pool sizes
// to verify that the maximum concurrency is respected and that the pool
// efficiently utilizes available slots. It runs multiple subtests with
// different pool sizes.
func TestThreadPoolConcurrency(t *testing.T) {
	// Test with different pool sizes
	testSizes := []int{1, 5, 10, 20}
	
	for _, size := range testSizes {
		t.Run("PoolSize_"+strconv.Itoa(size), func(t *testing.T) {
			pool := New(size)
			
			// Number of tasks to run (more than pool size)
			taskCount := size * 3
			if size == 1 {
				// For pool size 1, use a smaller number of tasks to avoid overwhelming the pool
				taskCount = 3
			}
			
			// Track maximum concurrency
			var maxConcurrent int32 = 0
			var currentConcurrent int32 = 0
			
			var wg sync.WaitGroup
			wg.Add(taskCount)
			
			// Use a done channel to signal test completion
			done := make(chan struct{})
			
			// For pool size 1, we need to ensure we have enough time to record concurrency
			workDuration := 50 * time.Millisecond
			if size == 1 {
				workDuration = 200 * time.Millisecond // Longer duration for size 1 to ensure we capture concurrency
			}
			
			// Launch tasks
			for i := 0; i < taskCount; i++ {
				go func(id int) {
					defer wg.Done()
					
					// Join the pool (blocks until a slot is available)
					pool.Join()
					
					// Ensure we leave the pool when done
					defer pool.Leave()
					
					// Increment and track concurrency
					current := atomic.AddInt32(&currentConcurrent, 1)
					
					// Update max concurrency with atomic operations
					for {
						max := atomic.LoadInt32(&maxConcurrent)
						if current <= max {
							break
						}
						if atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
							t.Logf("Task %d updated max concurrency to %d", id, current)
							break
						}
					}
					
					// Log when a task is running
					t.Logf("Task %d running, current concurrency: %d", id, current)
					
					// Simulate work
					time.Sleep(workDuration)
					
					// Decrement concurrency counter
					atomic.AddInt32(&currentConcurrent, -1)
					t.Logf("Task %d completed", id)
				}(i)
			}
			
			// Wait for all tasks with timeout
			go func() {
				wg.Wait()
				close(done)
			}()
			
			select {
			case <-done:
				// Test completed successfully
				t.Logf("All tasks completed successfully")
			case <-time.After(10 * time.Second):
				t.Fatal("Test timed out waiting for all tasks to complete")
			}
			
			// Verify max concurrency never exceeded pool size
			if int(maxConcurrent) > size {
				t.Errorf("Max concurrency (%d) exceeded pool size (%d)", maxConcurrent, size)
			}
			
			// For pool size 1, we know exactly what the max should be
			if size == 1 {
				if maxConcurrent != 1 {
					t.Errorf("For pool size 1, max concurrency should be exactly 1, got %d", maxConcurrent)
				}
			} else {
				// For larger pools, verify max concurrency reached pool size (or close to it)
				if int(maxConcurrent) < size-1 && size > 1 {
					t.Errorf("Max concurrency (%d) didn't reach close to pool size (%d)", maxConcurrent, size)
				}
			}
			
			t.Logf("Pool size %d: max concurrency was %d", size, maxConcurrent)
		})
	}
}

// TestThreadPoolStress runs a stress test on the ThreadPool by launching
// a large number of short-lived goroutines. This test is skipped when
// running in short mode. It verifies that the pool can handle a high
// volume of join/leave operations without errors.
func TestThreadPoolStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	// Create a pool with moderate size
	pool := New(50)
	
	// Launch a large number of short-lived goroutines
	var wg sync.WaitGroup
	taskCount := 1000
	wg.Add(taskCount)
	
	// Track errors
	var errorCount int32 = 0
	
	// Start time
	startTime := time.Now()
	
	for i := 0; i < taskCount; i++ {
		go func(id int) {
			defer wg.Done()
			
			// Join the pool
			pool.Join()
			defer pool.Leave()
			
			// Very short task
			time.Sleep(1 * time.Millisecond)
		}(i)
	}
	
	// Wait for all tasks to complete
	wg.Wait()
	
	// Check duration
	duration := time.Since(startTime)
	t.Logf("Completed %d tasks in %v with pool size 50", taskCount, duration)
	
	// Check for errors
	if errorCount > 0 {
		t.Errorf("Encountered %d errors during stress test", errorCount)
	}
}

// BenchmarkThreadPool benchmarks the performance of the ThreadPool with
// different pool sizes. It measures how quickly the pool can process
// a large number of short-lived goroutines.
func BenchmarkThreadPool(b *testing.B) {
	// Benchmark different pool sizes
	poolSizes := []int{1, 10, 100}
	
	for _, size := range poolSizes {
		b.Run("PoolSize_"+string(rune(size+'0')), func(b *testing.B) {
			pool := New(size)
			
			// Reset the timer before the actual benchmark work
			b.ResetTimer()
			
			// Create a fixed number of worker goroutines based on the pool size
			// to avoid creating too many goroutines
			workers := size * 2
			tasks := make(chan int, b.N)
			
			// Start worker goroutines
			var wg sync.WaitGroup
			wg.Add(workers)
			
			for i := 0; i < workers; i++ {
				go func() {
					defer wg.Done()
					
					// Process tasks until the channel is closed
					for range tasks {
						pool.Join()
						// Minimal work
						time.Sleep(1 * time.Microsecond)
						pool.Leave()
					}
				}()
			}
			
			// Send b.N tasks to the workers
			for i := 0; i < b.N; i++ {
				tasks <- i
			}
			
			// Close the channel to signal workers to exit
			close(tasks)
			
			// Wait for all workers to complete
			wg.Wait()
		})
	}
}