# Exercise 11: D&D Work Stealing - Rolling for Goroutines

> **Want to learn more?** Read [The Scheduler](https://internals-for-interns.com/posts/go-runtime-scheduler/) on Internals for Interns for a deep dive into Go's runtime and goroutine scheduling.

In this exercise, you'll add a d20 dice roll to the Go scheduler's work stealing algorithm. When a processor (P) tries to steal goroutines from another P's run queue, it must first roll above 10 on a twenty-sided die. Failed rolls mean the steal is blocked, making the scheduler's work distribution visible and entertaining.

## Learning Objectives

By the end of this exercise, you will:

- Understand how Go's work stealing scheduler distributes goroutines across processors
- Know where the `stealWork` function lives and how it iterates over other P's
- Modify the steal logic to add a randomized gate
- Observe work stealing attempts in real time

## Background: Work Stealing

Go's scheduler uses work stealing to balance load across processors. When a P runs out of goroutines to execute, it looks at other P's queues and steals half their work:

```
Before (current behavior):
  P0: [g1, g2, g3, g4]    P1: []  (idle)
  P1 tries to steal from P0 → always succeeds
  P0: [g1, g2]             P1: [g3, g4]

After (our modification):
  P0: [g1, g2, g3, g4]    P1: []  (idle)
  P1 rolls d20 to steal from P0 → rolled 7, failed!
  P1 rolls d20 to steal from P0 → rolled 16, stole!
  P0: [g1, g2]             P1: [g3, g4]
```

## Step 1: Understanding the Steal Mechanism

The work stealing logic lives in the `stealWork` function in `proc.go`:

```bash
cd go/src/runtime
grep -n "func stealWork" proc.go
```

You'll find `stealWork` around line 3828. This function is called by `findRunnable` when a P has no local work. It iterates over other P's in a randomized order, trying to steal goroutines from their run queues.

## Step 2: Find the Steal Attempt

Inside `stealWork`, look for the actual steal attempt around line 3883:

```bash
grep -n "runqsteal" proc.go | head -5
```

You'll see this code block (around line 3883-3887):

```go
// go/src/runtime/proc.go:3883-3887
// Don't bother to attempt to steal if p2 is idle.
if !idlepMask.read(enum.position()) {
    if gp := runqsteal(pp, p2, stealTimersOrRunNextG); gp != nil {
        return gp, false, now, pollUntil, ranTimer
    }
}
```

Key variables at this point:
- **`pp`** - The current P (the thief), type `*p`, has field `pp.id` (int32)
- **`p2`** - The target P (the victim), type `*p`, has field `p2.id`
- **`runqsteal(pp, p2, ...)`** - Moves goroutines from p2's queue to pp's queue

## Step 3: Add the D&D Dice Roll

Replace lines 3883-3887 with our dice-gated version:

```go
// go/src/runtime/proc.go:3883-3887
// Don't bother to attempt to steal if p2 is idle.
if !idlepMask.read(enum.position()) {
    if mainStarted && gogetenv("GODND") != "" {
        // D&D Work Stealing: Roll a d20 to attempt stealing!
        roll := cheaprandn(20) + 1 // Roll 1-20
        if roll > 10 {
            if gp := runqsteal(pp, p2, stealTimersOrRunNextG); gp != nil {
                println("🎲 [P", pp.id, "] Rolling to steal from P", p2.id, "... rolled", roll, ". Stole!")
                return gp, false, now, pollUntil, ranTimer
            }
        } else {
            println("🎲 [P", pp.id, "] Rolling to steal from P", p2.id, "... rolled", roll, ". Failed!")
        }
    } else {
        if gp := runqsteal(pp, p2, stealTimersOrRunNextG); gp != nil {
            return gp, false, now, pollUntil, ranTimer
        }
    }
}
```

### Understanding the Code

- **`mainStarted`** - A boolean already in `proc.go` that flips to `true` early in the main goroutine's startup — before `sysmon`, GC, and `init()` run. It cuts out the very earliest scheduler noise, but some pre-main prints will still appear (see the callout below)
- **`gogetenv("GODND")`** - The runtime's internal env var reader (equivalent of `os.Getenv` — the runtime can't import `os`). Gates everything behind `GODND=1` so the scheduler behaves normally unless you opt in
- **`cheaprandn(20) + 1`** - Rolls 1-20. `cheaprandn(20)` returns 0-19, the `+ 1` shifts it to a proper d20 range
- **`roll > 10`** - 50% success rate (rolls 11-20 succeed, rolls 1-10 fail)
- **`println(...)`** - Runtime's builtin print function, writes to stderr via raw syscall, no imports needed
- **`pp.id` / `p2.id`** - The processor ID fields (int32), defined in the `p` struct in `runtime2.go`
- We only print "Stole!" when `runqsteal` returns non-nil (the target queue might have emptied between the idle check and the steal)

> **🐉 Fun fact: The scheduler already rolls with advantage — times four!**
>
> Look at the outer loop in `stealWork`: `const stealTries = 4`. The scheduler doesn't just try to steal once — it loops **4 times** over all P's, re-shuffling the order with `cheaprand()` each pass. So your d20 gate doesn't just get rolled once per steal attempt — a determined P gets up to 4 chances per target. In D&D terms, that's rolling with advantage... squared.
>
> And the 4th pass is special: it sets `stealTimersOrRunNextG = true`, which unlocks stealing the victim's `runnext` goroutine — the one it was about to run next. The source code comment literally says *"stealing from the other P's runnext should be the last resort."* So the final pass is the desperate, gloves-off round where everything is fair game.
>
> The Go scheduler was already playing D&D before you got here. You're just making the dice visible.

## Step 4: Rebuild Go Toolchain

```bash
cd ../  # back to go/src
./make.bash
```

## Step 5: Test the D&D Scheduler

Create the `dnd_steal_demo.go` file:

```bash
# Create the file
touch /tmp/dnd_steal_demo.go
```

```go
package main

import (
    "runtime"
    "sync"
)

func busyWork(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    sum := 0
    for i := 0; i < 1_000_000; i++ {
        sum += i
    }
    println("Goroutine", id, "finished")
}

func main() {
    runtime.GOMAXPROCS(4) // 4 P's for visible stealing
    println("=== D&D Work Stealing Demo ===")
    println()

    var wg sync.WaitGroup
    for i := 1; i <= 20; i++ {
        wg.Add(1)
        go busyWork(i, &wg)
    }
    wg.Wait()
    println()
    println("=== All goroutines completed! ===")
}
```

Build and run with D&D mode enabled:

```bash
../go/bin/go build -o dnd_steal_demo /tmp/dnd_steal_demo.go
GODND=1 ./dnd_steal_demo
```

**Why `GODND=1`?** Without it, the scheduler runs normally — no dice rolls, no prints. This keeps `./make.bash` and `go build` clean and silent. Note that some pre-main and post-main prints will still appear even with `GODND=1` — see the callouts below for why.

**Why build first then run?** The `go build` command itself uses your modified Go. Building separately (without `GODND=1`) keeps the compiler's work stealing silent.

Expected output (varies every run):

```
🎲 [P 3 ] Rolling to steal from P 0 ... rolled 13 . Stole!
=== D&D Work Stealing Demo ===

🎲 [P 2 ] Rolling to steal from P 0 ... rolled 15 . Stole!
🎲 [P 3 ] Rolling to steal from P 0 ... rolled 12 . Stole!
🎲 [P 1 ] Rolling to steal from P 0 ... rolled 11 . Stole!
Goroutine 15 finished
🎲 [P 0 ] Rolling to steal from P 2 ... rolled 7 . Failed!
🎲 [P 0 ] Rolling to steal from P 3 ... rolled 20 . Stole!
Goroutine 1 finished
Goroutine 12 finished
...
=== All goroutines completed! ===
🎲 [P %
```

You'll see the dice rolls interleaved with goroutine completion messages. Rolls of 1-10 fail, rolls of 11-20 succeed. Notice how a P with no work keeps retrying different targets until it rolls high enough!

> **📖 Why are there dice rolls before `=== D&D Work Stealing Demo ===`?**
>
> You didn't write those — the Go runtime did. Between `mainStarted` flipping to `true` and your first `println`, the runtime starts `sysmon`, enables the GC, and runs all `init()` functions — each spawning goroutines that idle P's immediately try to steal. You've instrumented the scheduler itself, so you're now seeing activity that was always there, just silent.

> **📖 Why is there a truncated `🎲 [P %` after `=== All goroutines completed! ===`?**
>
> Same story, other end. After `wg.Wait()` returns, the idle P's are still spinning in `stealWork` looking for more work. One starts a `println` just as `main()` returns and the process exits — the print never finishes. The `%` is your shell saying the output didn't end with a newline. The scheduler doesn't stop for you on the way out either.

> Welcome to the inside of Go!


## Understanding What We Did

1. **Found the Steal Logic**: Located `stealWork` in `proc.go` where P's steal goroutines from each other
2. **Added a Dice Gate**: Used `cheaprandn(20) + 1` to generate a d20 roll (1-20) before each steal attempt
3. **Logged the Rolls**: Added `println()` calls showing which P is stealing from which, the roll result, and whether it succeeded
4. **Observed the Effect**: Saw work stealing attempts in real time, with some failing due to low rolls

## What We Learned

- **Work Stealing**: How Go distributes goroutines across processors when queues are imbalanced
- **`stealWork` Function**: The core loop that iterates over P's looking for work to steal
- **`cheaprandn`**: The runtime's fast pseudo-random number generator, used throughout the scheduler
- **Scheduler Observability**: How to add logging to the scheduler without breaking its behavior
- **P Identity**: Each processor has a unique `id` field that identifies it in scheduling decisions

## Extension Ideas

1. Make the difficulty configurable: roll must beat 15 instead of 10 (harder steals)
2. Add a "critical hit" on natural 20: steal ALL goroutines from target, not just half
3. Add a "fumble" on natural 1: the stealing P yields for one cycle with `Gosched()`
4. Track and print total rolls, successes, and failures at program exit

## Cleanup

To remove the dice roll:

```bash
cd go/src/runtime
git checkout proc.go
cd ../
./make.bash
```

## Summary

You've turned Go's work stealing scheduler into a tabletop RPG encounter:

```
Before:  P tries to steal -> always succeeds if target has work
After:   P tries to steal -> must roll > 10 on a d20 first

Changes: runtime/proc.go stealWork() function (~14 lines)
Result:  The scheduler now plays D&D!
```

---

*Congratulations on completing all workshop exercises! Return to the [main workshop](../README.md)*
