# Exercise 9: Predictable Select - Making Select Statements Deterministic

> 📖 **Want to learn more?** Read [The Scheduler](https://internals-for-interns.com/posts/go-runtime-scheduler/) on Internals for Interns for a deep dive into Go's runtime and goroutine scheduling.

In this exercise, you'll modify Go's `select` statement to be deterministic instead of random. By default, Go randomizes which case is chosen when multiple channels are ready. We'll change it to always choose cases in the same order.

## Learning Objectives

By the end of this exercise, you will:

- Understand how Go's `select` statement is implemented
- Know why Go uses randomization (fairness vs. starvation)
- Modify the runtime's channel selection algorithm
- Test deterministic vs. random selection behavior

## Introduction: How Does Select Work Internally?

The `select` statement is implemented in the runtime function `selectgo()` in `runtime/select.go`. When your code reaches a `select` with multiple cases, the runtime needs to decide which case to execute. It does this using two arrays:

- **`pollorder`**: Determines the order in which cases are **checked** for readiness. By default, this order is randomized using `cheaprandn()` to ensure fairness — no single channel gets priority over others.
- **`lockorder`**: Determines the order in which channel locks are **acquired** (sorted by address to prevent deadlocks).

The runtime first shuffles the cases into `pollorder`, then iterates through them in that order. If a case's channel is ready (has data to receive or space to send), that case is selected. If no case is ready and there's a `default`, the default runs. Otherwise, the goroutine parks itself on all the channels' wait queues and waits until one becomes ready.

The randomization in `pollorder` is what makes `select` non-deterministic — running the same `select` with the same ready channels will pick different cases each time. This is a deliberate design choice to prevent programs from accidentally depending on case ordering.

## Background: Go Randomizes Select

By default, when multiple channels are ready, Go randomizes which case executes:

```go
select {
case v := <-ch1:  // Sometimes chosen
case v := <-ch2:  // Sometimes chosen
case v := <-ch3:  // Sometimes chosen
}
// Random selection prevents starvation
```

We'll make it deterministic:

```go
select {
case v := <-ch1:  // ALWAYS chosen first when ready
case v := <-ch2:  // Only if ch1 not ready
case v := <-ch3:  // Only if ch1 and ch2 not ready
}
// Predictable, source-order selection
```

## Step 1: Create a Test to See Current Randomization

Create a `random_select_demo.go` file:

```go
package main

func main() {
    ch1 := make(chan int, 1)
    ch2 := make(chan int, 1)
    ch3 := make(chan int, 1)

    // Fill all channels so they're all ready
    ch1 <- 1
    ch2 <- 2
    ch3 <- 3

    // Run select 10 times to see randomization
    for i := 0; i < 10; i++ {
        select {
        case v := <-ch1:
            println("Round", i, ": Selected ch1 (value", v, ")")
            ch1 <- 1 // Refill
        case v := <-ch2:
            println("Round", i, ": Selected ch2 (value", v, ")")
            ch2 <- 2 // Refill
        case v := <-ch3:
            println("Round", i, ": Selected ch3 (value", v, ")")
            ch3 <- 3 // Refill
        }
    }
}
```

Run with current Go to see random selection:

```bash
go run random_select_demo.go
```

Output shows random selection:

```
Round 0: Selected ch3 (value 3)
Round 1: Selected ch1 (value 1)
Round 2: Selected ch2 (value 2)
...
```

## Step 2: Navigate to the Select Implementation

```bash
cd go/src/runtime
```

The `select.go` file contains the entire select statement implementation. The key function is `selectgo()` which handles case selection.

## Step 3: Understand the Randomization Code

Look around line 191 in `select.go`:

```go
// go/src/runtime/select.go:191
j := cheaprandn(uint32(norder + 1))  // Random index!
pollorder[norder] = pollorder[j]
pollorder[j] = uint16(i)
norder++
```

This implements the algorithm to randomize case order:

- `cheaprandn()` generates a pseudo-random number
- Cases are placed in random positions in the `pollorder` array
- Select then checks cases in this randomized order

## Step 4: Make Select Deterministic

**Edit `select.go`:**

Find line 191 and change the randomization to be deterministic:

```go
// go/src/runtime/select.go:191
// Original:
j := cheaprandn(uint32(norder + 1))
pollorder[norder] = pollorder[j]
pollorder[j] = uint16(i)

// Change to:
pollorder[norder] = uint16(len(scases)-1-i)
```

### Understanding the Code Change


- **`uint16(len(scases)-1-i)`**: Use inverse order here
- **Result**: pollorder is now always ordered in the source code order
- **Effect**: Cases maintain their source code order in `pollorder`

## Step 5: Rebuild Go Runtime

```bash
cd ../  # back to go/src
./make.bash
```

## Step 6: Test Deterministic Behavior

```bash
../go/bin/go run random_select_demo.go
```

Now you should see **deterministic output**:

```
Round 0: Selected ch1 (value 1)
Round 1: Selected ch1 (value 1)
Round 2: Selected ch1 (value 1)
Round 3: Selected ch1 (value 1)
...
```

Perfect! `ch1` is **always** chosen because is the first one in the code, no more random order.

## Understanding What We Did

1. **Removed Randomization**: Replaced `cheaprandn()` with deterministic index
2. **Maintained Source Order**: Cases are now checked in the order they appear
3. **Performance Boost**: Slightly faster (no random number generation)
4. **Changed Semantics**: Same syntax, different runtime behavior

## What We Learned

- **Runtime Modification**: How to alter fundamental language behavior
- **Design Trade-offs**: Fairness vs. determinism in concurrent systems
- **Select Internals**: How `selectgo` and `pollorder` work
- **Behavioral Testing**: Validating semantic changes with test programs

## Extension Ideas

Try these additional modifications:

1. Add a reverse-order mode (check cases last to first)
2. Add priority levels based on case position
3. Track selection statistics for debugging
4. Make randomization configurable via environment variable

## Cleanup

To restore Go's original random behavior:

```bash
cd go/src/runtime
git checkout select.go
cd ../
./make.bash
```

## Summary

You've transformed Go's `select` from a fair, random chooser into a predictable, deterministic priority system:

```go
// Before: Random selection (fair but unpredictable)
select {
case <-ch1: // 33% chance
case <-ch2: // 33% chance
case <-ch3: // 33% chance
}

// After: Deterministic selection (predictable but may starve)
select {
case <-ch1: // Always chosen when ready
case <-ch2: // Only if ch1 not ready
case <-ch3: // Only if ch1 and ch2 not ready
}
```

This exercise demonstrated how runtime modifications can fundamentally change language behavior and exposed important trade-offs in concurrent system design.

---

*Continue to [Exercise 10](10-java-style-stack-traces.md) or return to the [main workshop](../README.md)*
