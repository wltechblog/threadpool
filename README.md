# ThreadQueue

A simple and effective way to manage running lots of goroutines in your Go software.

## Overview

ThreadQueue provides a thread Queue implementation for Go that limits the number of concurrently running goroutines. This helps prevent resource exhaustion when dealing with a large number of concurrent tasks.

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
    // Create a new thread Queue with a maximum of 5 concurrent goroutines
    Queue := threadqueue.New(5)

    var wg sync.WaitGroup

    // Launch 20 goroutines, but only 5 will run at a time
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            // Join the Queue (blocks until a slot is available)
            Queue.Join()
            defer Queue.Leave()

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
Queue.Join()
defer Queue.Leave()
```

2. Use a wait group to ensure all goroutines complete before your program exits:

```go
var wg sync.WaitGroup
wg.Add(numTasks)

for i := 0; i < numTasks; i++ {
    go func() {
        defer wg.Done()
        Queue.Join()
        defer Queue.Leave()
        // Do work
    }()
}

wg.Wait()
```

3. Choose an appropriate Queue size based on your system resources and the nature of your tasks:
   - CPU-bound tasks: typically use `runtime.NumCPU()` or slightly higher
   - I/O-bound tasks: can use a higher number as these tasks spend time waiting

## API Reference

### New(max int) *ThreadQueue

Creates a new thread Queue with the specified maximum number of concurrent goroutines.

```go
Queue := threadqueue.New(10) // Create a Queue with max 10 concurrent goroutines
```

### Join()

Blocks until the calling goroutine can enter the thread Queue. If the Queue is at capacity, the goroutine will wait until another goroutine leaves.

```go
Queue.Join() // Wait until we can enter the Queue
```

### Leave()

Notifies the Queue that the calling goroutine has completed its work and is exiting the Queue. This frees up a slot for another waiting goroutine.

```go
Queue.Leave() // Signal that we're done with the Queue
```

## License

Provided under the GPL2.0 license.