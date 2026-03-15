# ☕ Exercise 10: Java-Style Stack Traces - Making Go Panics Look Familiar

In this exercise, you'll modify Go's stack trace formatting to match Java's style! 🔄 Instead of Go's stack traces, we'll create Java-style traces.

## 🎯 Learning Objectives

By the end of this exercise, you will:

- ✅ Understand how Go formats stack traces in the runtime
- ✅ Know where panic messages are generated
- ✅ Modify runtime output formatting

## 🧠 Background: Stack Trace Styles

We're transforming Go's stacktrace format:

```
panic: Something went wrong

goroutine 1 [running]:
main.methodC()
        /Users/dev/project/main.go:15 +0x45
main.methodB()
        /Users/dev/project/main.go:11 +0x23
main.methodA()
        /Users/dev/project/main.go:7 +0x12
```

Into this Java-style format:

```
Exception in thread "main" go.runtime.Panic: Something went wrong
    at main.methodC(main.go:15)
    at main.methodB(main.go:11)
    at main.methodA(main.go:7)
```

## 🔍 Step 1: Create a Test Program

Create a `stack_trace_demo.go` file:

```go
package main

import "fmt"

func methodC() {
    panic("Something went wrong")
}

func methodB() {
    methodC()
}

func methodA() {
    methodB()
}

func main() {
    fmt.Println("Starting the program...")
    methodA()
}
```

Run with current Go to see the stacktrace format:

```bash
go run stack_trace_demo.go
```

## Step 2: Navigate to Runtime Files

```bash
cd go/src/runtime
```

Key files we'll modify:
- **`panic.go`** - Panic header message
- **`traceback.go`** - Stack frame formatting

## Step 3: Modify the Panic Header

**Edit `panic.go`:**

Find the `printpanics` function around line 734. Look for:

```go
print("panic: ")
printpanicval(p.arg)
```

Change to:

```go
print("Exception in thread \"main\" go.runtime.Panic: ")
printpanicval(p.arg)
```

## Step 4: Remove Goroutine Header

**Edit `traceback.go`:**

Find the `goroutineheader` function around line 1215. Add a return statement at the beginning:

```go
func goroutineheader(gp *g) {
    return  // Add this line to skip printing goroutine info
    level, _, _ := gotraceback()
    // ... rest of original code below (now unreachable)
}
```

## Step 5: Transform Stack Frame Formatting

**Still in `traceback.go`:**

Find the `traceback2` function around line 945. Comment out the `gotraceback()` call (around line 966):

```go
gp := u.g.ptr()
// level, _, _ := gotraceback()  // Comment this out
var cgoBuf [32]uintptr
```

Then find where stack frames are printed (around line 991-1008). Replace this entire section:

```go
printFuncName(name)
print("(")
if iu.isInlined(uf) {
    print("...")
} else {
    argp := unsafe.Pointer(u.frame.argp)
    printArgs(f, argp, u.symPC())
}
print(")\n")
print("\t", file, ":", line)
if !iu.isInlined(uf) {
    if u.frame.pc > f.entry() {
        print(" +", hex(u.frame.pc-f.entry()))
    }
    if gp.m != nil && gp.m.throwing >= throwTypeRuntime && gp == gp.m.curg || level >= 2 {
        print(" fp=", hex(u.frame.fp), " sp=", hex(u.frame.sp), " pc=", hex(u.frame.pc))
    }
}
print("\n")
```

With this Java-style format:

```go
// Extract just the filename (not full path)
fileName := file
for i := len(file) - 1; i >= 0; i-- {
    if file[i] == '/' || file[i] == '\\' {
        fileName = file[i+1:]
        break
    }
}
print("    at ", name, "(", fileName, ":", line, ")\n")
```

## Step 6: Rebuild Go Runtime

```bash
cd ../  # back to go/src
./make.bash
```

## Step 7: Test Java-Style Stack Traces

```bash
../go/bin/go run stack_trace_demo.go
```

You should see:

```
Starting the program...
Exception in thread "main" go.runtime.Panic: Something went wrong
    at main.methodC(stack_trace_demo.go:6)
    at main.methodB(stack_trace_demo.go:10)
    at main.methodA(stack_trace_demo.go:14)
    at main.main(stack_trace_demo.go:19)
```

## Understanding What We Did

1. **Changed Panic Header** (`panic.go` line 747): Changed `"panic: "` to `"Exception in thread \"main\" go.runtime.Panic: "`
2. **Removed Goroutine Info** (`traceback.go` line 1215): Added early `return` in `goroutineheader()`
3. **Simplified Stack Frames** (`traceback.go` line 991-1008): Replaced the go output with the java `"    at name(file:line)"` format
4. **Removed Debug Info**: Commented out `gotraceback()` call and eliminated hex offsets, frame pointers
5. **Basename Only**: Extract filename from full path using loop

## 🎓 What We Learned

- 🔍 **Runtime Formatting**: How Go generates stack traces
- 📝 **Panic Handling**: Where panic messages originate
- 🎨 **Output Control**: Modifying runtime print statements

## 💡 Extension Ideas

Try these additional modifications: 🚀

1. ➕ Add color to the output (red for "Exception")
2. ➕ Make it configurable via environment variable
3. ➕ Add Python-style formatting as another option
4. ➕ Include package path conversion (github.com/user/pkg → github.com.user.pkg)

## Cleanup

To restore Go's original stack trace format:

```bash
cd go/src/runtime
git checkout panic.go traceback.go
cd ../
./make.bash
```

## Summary

You've transformed Go's stack traces into Java-style formatting:

```
// Before: Technical and verbose
goroutine 1 [running]:
main.methodC()
        /full/path/to/main.go:15 +0x45

// After: Clean and familiar
Exception in thread "main" go.runtime.Panic: ...
    at main.methodC(main.go:15)
```

---

*Congratulations on completing all workshop exercises! Return to the [main workshop](../README.md)*
