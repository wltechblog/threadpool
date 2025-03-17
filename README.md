# ThreadQueue

A simple and effective way to manage running lots of goroutines in your Go software.

## Overview

ThreadQueue provides a thread pool implementation for Go that limits the number of concurrently running goroutines. This helps prevent resource exhaustion when dealing with a large number of concurrent tasks.

## Installation

```bash
go get github.com/wltechblog/threadqueue
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "sync"
    "time"

    "github.com/wltechblog/threadqueue"
)

func main() {
    // Create a new thread pool with a maximum of 5 concurrent goroutines
    pool := threadqueue.New(5)

    var wg sync.WaitGroup

    // Launch 20 goroutines, but only 5 will run at a time
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            // Join the pool (blocks until a slot is available)
            pool.Join()
            defer pool.Leave()

            // Do some work
            fmt.Printf("Worker %d is running\n", id)
            time.Sleep(100 * time.Millisecond)
            fmt.Printf("Worker %d is done\n", id)
        }(i)
    }

    wg.Wait()
}
```

### Best Practices

1. Always call `Leave()` after `Join()`, preferably using `defer`:

```go
pool.Join()
defer pool.Leave()
```

2. Use a wait group to ensure all goroutines complete before your program exits:

```go
var wg sync.WaitGroup
wg.Add(numTasks)

for i := 0; i < numTasks; i++ {
    go func() {
        defer wg.Done()
        pool.Join()
        defer pool.Leave()
        // Do work
    }()
}

wg.Wait()
```

3. Choose an appropriate pool size based on your system resources and the nature of your tasks:
   - CPU-bound tasks: typically use `runtime.NumCPU()` or slightly higher
   - I/O-bound tasks: can use a higher number as these tasks spend time waiting

## API Reference

### New(max int) *ThreadPool

Creates a new thread pool with the specified maximum number of concurrent goroutines.

```go
pool := threadqueue.New(10) // Create a pool with max 10 concurrent goroutines
```

### Join()

Blocks until the calling goroutine can enter the thread pool. If the pool is at capacity, the goroutine will wait until another goroutine leaves.

```go
pool.Join() // Wait until we can enter the pool
```

### Leave()

Notifies the pool that the calling goroutine has completed its work and is exiting the pool. This frees up a slot for another waiting goroutine.

```go
pool.Leave() // Signal that we're done with the pool
```

## License

[License information]
