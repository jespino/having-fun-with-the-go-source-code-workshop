# Exercise 8: Goroutine Sleep Detective - Runtime State Monitoring

> 📖 **Want to learn more?** Read [The Scheduler](https://internals-for-interns.com/posts/go-runtime-scheduler/) on Internals for Interns for a deep dive into Go's goroutine scheduling and state transitions.

In this exercise, you'll modify the Go runtime scheduler to log goroutine state transitions. Every time a goroutine goes to sleep waiting for something, it will announce itself: "Hello, I'm goroutine 42, going to sleep waiting for channel receive".

## Learning Objectives

By the end of this exercise, you will:

- Understand Go's goroutine scheduler state transitions
- Know where goroutines block in the runtime
- Modify the scheduler for debugging insights

## Introduction: How Does the Scheduler Work?

Go's scheduler uses the **GMP model** (Goroutines, Machines, Processors) to map potentially thousands of goroutines onto a small number of OS threads. The key insight is that when an OS thread blocks (e.g., on a syscall), the scheduling resources (the P) can detach and move to another thread, keeping work flowing.

Goroutines don't have a dedicated scheduler thread managing them. Instead, they manage their own transitions through a **self-service pattern**: when a goroutine needs to wait (for a channel, mutex, sleep, etc.), it calls `gopark()` which parks itself, adds itself to the appropriate wait queue, and then calls `schedule()` to find the next runnable goroutine. When the wait condition is satisfied, `goready()` moves the goroutine back to a runnable state.

The scheduler picks the next goroutine to run following a priority order: GC work first, then the local `runnext` slot, then the local run queue, then the global queue (checked every 61st call to prevent starvation), then network poller results, and finally work-stealing from other Ps.

Understanding this scheduling flow is essential because in this exercise we'll be adding logging at the exact point where goroutines transition to the waiting state.

## Background: Goroutine States

Go manages goroutines through different states:

- **`_Grunnable`** - Ready to run but not executing
- **`_Grunning`** - Currently executing
- **`_Gwaiting`** - Blocked waiting for something (our target!)
- **`_Gsyscall`** - Executing a system call
- ...

When a goroutine needs to wait (for channels, mutexes, sleep, etc.), it "parks" and transitions to `_Gwaiting` state.

## Step 1: Understanding the Park Mechanism

The `gopark` function is called by ALL synchronization primitives when a goroutine needs to wait.

```bash
cd go/src/runtime
grep -n "func gopark" proc.go
```

Key functions:

- **`gopark()`** - Initiates parking a goroutine
- **`park_m()`** - Actually changes the state to `_Gwaiting`

## Step 2: Find the State Transition Code

```bash
# Look at where the state actually changes
grep -n -A 5 "func park_m" proc.go
```

Around line 4275, you'll see:

```go
casgstatus(gp, _Grunning, _Gwaiting)
```

This is the exact line where a goroutine transitions from running to waiting. Perfect for our logging!

## Step 3: Add Goroutine Sleep Logging

**Edit `proc.go`:**

You'll need to add logging in three locations where goroutines transition to the waiting state:

### Location 1: `casGToWaiting` function (around line 1388)

Find the `casGToWaiting` function and add logging after setting the wait reason:

```go
func casGToWaiting(gp *g, old uint32, reason waitReason) {
	// Set the wait reason before calling casgstatus, because casgstatus will use it.
	gp.waitreason = reason
	if gp.goid > 1 { // Skip system goroutines 0 and 1
		print("Hello, I'm goroutine ", gp.goid, ", going to sleep waiting for ", gp.waitreason.String(), "\n")
	}
	casgstatus(gp, old, _Gwaiting)
}
```

### Location 2: `casGFromPreempted` function (around line 1430)

Find where preempted goroutines transition to waiting. Add logging after the `waitreason` is set but before the `CompareAndSwap`:

```go
func casGFromPreempted(gp *g, old, new uint32) bool {
	if old != _Gpreempted || new != _Gwaiting {
		throw("bad g transition")
	}
	gp.waitreason = waitReasonPreempted
	if gp.goid > 1 { // Skip system goroutines 0 and 1
		print("Hello, I'm goroutine ", gp.goid, ", going to sleep waiting for ", gp.waitreason.String(), "\n")
	}
	if !gp.atomicstatus.CompareAndSwap(_Gpreempted, _Gwaiting) {
		return false
	}
	if bubble := gp.bubble; bubble != nil {
		bubble.changegstatus(gp, _Gpreempted, _Gwaiting)
	}
	return true
}
```

### Location 3: `park_m` function (around line 4275)

Find the `park_m` function and add logging before the direct `casgstatus` call:

```go
// Add this before: casgstatus(gp, _Grunning, _Gwaiting)
if gp.goid > 1 { // Skip system goroutines 0 and 1
    print("Hello, I'm goroutine ", gp.goid, ", going to sleep waiting for ", gp.waitreason.String(), "\n")
}
casgstatus(gp, _Grunning, _Gwaiting)
```

### Understanding the Code

- **`gp.goid`** - Unique goroutine ID
- **`gp.waitreason.String()`** - Human-readable reason for waiting (channel, mutex, sleep, etc.)
- **`print()`** - Runtime print function (outputs to stderr)
- **`gp.goid > 1`** - Skip system goroutines to reduce noise

## Step 4: Rebuild Go Runtime

```bash
cd ../  # back to go/src
./make.bash
```

## Step 5: Test Channel Blocking

Create a `channel_demo.go` file:

```go
package main

import "time"

func main() {
    ch := make(chan string)

    // Start goroutine that will block on receive
    go func() {
        msg := <-ch  // Should trigger our logging!
        println("Received:", msg)
    }()

    // Let the goroutine start and block
    time.Sleep(100 * time.Millisecond)

    // Send something
    ch <- "Hello!"
    time.Sleep(10 * time.Millisecond)
}
```

Build and run with our modified Go:

```bash
../go/bin/go build channel_demo.go
./channel_demo
```

**Note:** We build the binary first and then run it directly. This avoids mixing goroutines from the compiler/build process with goroutines from our program, giving us cleaner output!

Expected output:

```
Hello, I'm goroutine 4, going to sleep waiting for GC scavenge wait
Hello, I'm goroutine 3, going to sleep waiting for GC sweep wait
Hello, I'm goroutine 2, going to sleep waiting for force gc (idle)
Hello, I'm goroutine 6, going to sleep waiting for chan receive
Hello, I'm goroutine 5, going to sleep waiting for GOMAXPROCS updater (idle)
Received: Hello!
```

You can now see goroutines blocking.

## Understanding What We Did

1. **Found the Park Function**: Located where goroutines transition to waiting state
2. **Added Logging**: Inserted print statement before state change
3. **Captured Wait Reason**: Used `gp.waitreason.String()` for human-readable output
4. **Tested Scenarios**: Verified with channels, mutexes, sleep, and select

Common wait reasons you'll see:

- `chan receive` / `chan send`
- `sync mutex lock`
- `sleep`
- `GC`

## What We Learned

- **Goroutine Lifecycle**: How goroutines transition between states
- **Park Mechanism**: The `gopark` and `park_m` functions
- **Synchronization Internals**: Where channels, mutexes, and select cause blocking
- **Runtime Debugging**: How to add observability to the Go runtime
- **Concurrency Visibility**: Real-time observation of blocking operations

## Extension Ideas

Try these additional modifications:

1. Add goroutine wakeup logging (when they resume running)
2. Add icons for different wait reasons (channel, mutex, sleep)
3. Include timestamps to measure blocking duration
4. Filter logging by specific wait reasons only

## Cleanup

To remove the logging:

```bash
cd go/src/runtime
git checkout proc.go
cd ../
./make.bash
```

## Summary

You've gained X-ray vision into Go's concurrency model! Your modified runtime now announces every goroutine blocking operation:

```
Hello, I'm goroutine 18, going to sleep waiting for chan receive
Hello, I'm goroutine 19, going to sleep waiting for sync mutex lock
Hello, I'm goroutine 20, going to sleep waiting for sleep
```

This exercise revealed the internal workings of Go's scheduler and how synchronization primitives interact with the runtime.

---

*Continue to [Exercise 9](09-predictable-select.md) or return to the [main workshop](../README.md)*
