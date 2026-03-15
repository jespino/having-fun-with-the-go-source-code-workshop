# Exercise 7: Patient Go - Making Go Wait for Goroutines

> 📖 **Want to learn more?** Read [The Bootstrap](https://internals-for-interns.com/posts/understanding-go-runtime/) and [The Scheduler](https://internals-for-interns.com/posts/go-runtime-scheduler/) on Internals for Interns for a deep dive into Go's runtime startup and goroutine scheduling.

In this exercise, you'll modify the Go runtime to wait for all goroutines to complete before the program exits. Currently, when `main()` returns, Go immediately terminates even if goroutines are still running. We'll make Go "patient" by waiting for all goroutines to finish.

## Learning Objectives

By the end of this exercise, you will:

- Understand Go's program termination process
- Know how to count active goroutines
- Modify the main runtime function to change program behavior
- Understand the trade-offs of automatic goroutine waiting

## Background: Go's Current Termination Behavior

Currently, when you write:

```go
package main

import "time"

func main() {
    go func() {
        time.Sleep(2 * time.Second)
        println("Goroutine finished!")
    }()
    println("Main finished!")
    // Program exits immediately, goroutine never completes
}
```

**Output:**
```
Main finished!
```

The goroutine never gets to print because the program exits when `main()` returns.

We'll change this so Go waits patiently for all goroutines to finish:

**New Output:**
```
Main finished!
Goroutine finished!
```

## Step 1: Understanding the Runtime Main Function

The Go runtime's `main()` function in `runtime/proc.go` is responsible for running your program's `main()` function. Let's examine how this works:

```bash
cd go/src/runtime
```

Open `proc.go` and find the `main()` function. Near the top (around line 136-137), you'll see how the runtime links to your program's main:

```go
//go:linkname main_main main.main
func main_main()
```

This `//go:linkname` directive tells the linker to connect the runtime's `main_main` function to your program's `main.main` function. This is how the runtime can call code from your main package.

Further down in the same `main()` function (around line 289), you'll see where this gets called:

```go
fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
fn()

... // tear-down process continues
```

**How it works:**

1. The go runtime boostrap process happens
2. The runtime's `main()` function runs first
3. A bit more of boostrap process
4. The `main_main` (which is your program's `main()` function via linkname) is called
5. Your `main()` function executes - **responsibility is delegated to your code**
6. When your `main()` returns, control returns to the runtime's `main()` function
7. The runtime continues with the program **tear-down process** (cleanup and exit)

Currently, the tear-down starts immediately after your `main()` returns, without waiting for other goroutines.

## Step 2: Add Goroutine Waiting Logic

We'll add code to wait until only 1 goroutine remains (the main goroutine itself).

**Edit `runtime/proc.go`:**

Find the section around line 289-290 where `main_main` is called:

```go
fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
fn()
```

Add the waiting logic right after the `fn()` call:

```go
fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
fn()

// Wait until only 1 goroutine is running (the main goroutine)
for gcount(false) > 1 {
	Gosched()
}
```

### Understanding the Code

- **`gcount(false)`** - Runtime function that returns the number of active goroutines (the `false` argument excludes system goroutines from the count)
- **`gcount(false) > 1`** - While more than just the main goroutine is running
- **`Gosched()`** - Yields the processor, allowing other goroutines to run
- **Loop terminates** - When only the main goroutine remains (count = 1)

## Step 3: Rebuild Go Toolchain

```bash
cd go/src
./make.bash
```

This rebuilds the runtime with your patient goroutine waiting logic.

## Step 4: Test Basic Goroutine Waiting

Create a test file to verify the behavior:

Create `patient_demo.go`:

```go
package main

import "time"

func main() {
	println("Main starting...")

	go func() {
		time.Sleep(1 * time.Second)
		println("Goroutine 1 finished!")
	}()

	go func() {
		time.Sleep(2 * time.Second)
		println("Goroutine 2 finished!")
	}()

	println("Main finished, but Go will wait...")
}
```

Run with your modified Go:

```bash
./bin/go run patient_demo.go
```

**Expected output:**

```
Main starting...
Main finished, but Go will wait...
Goroutine 1 finished!
Goroutine 2 finished!
```

Success! Go now waits for all goroutines to complete.

## What We Learned

- **Program Termination**: How Go programs exit and cleanup
- **Goroutine Tracking**: The `gcount()` function tracks active goroutines
- **Cooperative Scheduling**: `Gosched()` yields to allow other goroutines to run
- **Runtime Modification**: How a small change affects all Go programs
- **Design Trade-offs**: Benefits and drawbacks of automatic waiting

## Extension Ideas

Try these additional modifications:

1. Add a timeout: Wait maximum 10 seconds for goroutines
2. Add logging: Print when waiting starts and which goroutines remain
3. Make it configurable: Use environment variable to enable/disable
4. Add a warning: Detect infinite loops in goroutines

## Cleanup

To restore standard Go behavior:

```bash
cd go/src/runtime
git checkout proc.go
cd ..
./make.bash
```

## Summary

You've successfully modified Go's runtime to be "patient" and wait for all goroutines!

```
Before:  main() returns → immediate exit → goroutines abandoned
After:   main() returns → wait for goroutines → all complete → exit

Changes: runtime/proc.go main() function
Result:  No goroutine left behind!
```

This modification demonstrates:

- Deep understanding of the Go runtime
- How program termination works
- The relationship between main() and goroutines
- Real-world trade-offs in language design

Your Go is now patient.

---

*Continue to [Exercise 8](08-goroutine-sleep-detective.md) or return to the [main workshop](../README.md)*
