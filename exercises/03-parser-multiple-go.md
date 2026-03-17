# Exercise 3: Multiple "go" Keywords - Parser Enhancement

> 📖 **Want to learn more?** Read [The Parser](https://internals-for-interns.com/posts/the-go-parser/) on Internals for Interns for a deep dive into how Go's parser builds Abstract Syntax Trees.

In this exercise, you'll modify the Go parser to accept multiple consecutive "go" keywords for starting goroutines! This will teach you how to enhance parser logic to handle repetitive syntax patterns while maintaining the same semantic behavior.

## Learning Objectives

By the end of this exercise, you will:

- Understand Go's parser structure and token consumption
- Know how to modify parser logic for syntax extensions
- Test parser modifications with working code

## Introduction: What is a Parser?

The parser is the second phase of the compiler, right after the scanner. While the scanner produces a flat stream of tokens, the parser's job is to give that stream **structure** by building an **Abstract Syntax Tree (AST)** — a tree that represents the hierarchical relationships in your code.

For example, a `go sayHello()` statement becomes a tree node of type `CallStmt` with `Tok: _Go` and a child node representing the function call `sayHello()`. The parser knows that after seeing a `go` token, a function call expression must follow — this is the grammar of the language.

Go's parser uses a technique called **recursive descent**: it has a function for each grammar rule (file, declaration, statement, expression), and these functions call each other top-down. The entry point `fileOrNil()` parses the package clause, then imports, then declarations. Each declaration can contain statements, and each statement can contain expressions.

The parser consumes tokens one at a time using `p.next()`, and checks the current token with `p.tok`. The parser lives in `go/src/cmd/compile/internal/syntax/parser.go`.

## Step 1: Navigate to the Parser

```bash
cd go/src/cmd/compile/internal/syntax
```

### Understanding the Current Parser Logic

Let's examine how the parser currently handles the "go" statement in `parser.go`. Look around line 2675:

```go
// go/src/cmd/compile/internal/syntax/parser.go:2673-2676
...
return s

case _Go, _Defer:
    return p.callStmt()
...
```

The parser recognizes `_Go` token and immediately calls `p.callStmt()` to handle the goroutine creation.

Find the `callStmt()` method in `parser.go` at line 977. This is where we'll add our multiple "go" logic:

```go
// go/src/cmd/compile/internal/syntax/parser.go:976-985
// callStmt parses call-like statements that can be preceded by 'defer' and 'go'.
func (p *parser) callStmt() *CallStmt {
    if trace {
        defer p.trace("callStmt")()
    }

    s := new(CallStmt)
    s.pos = p.pos()
    s.Tok = p.tok // _Defer or _Go
    p.next()
    ...
}
```

The key line is `s.Tok = p.tok` which captures whether this is a "defer" or "go" statement, followed by `p.next()` which consumes the token.

## Step 2: Add Multiple "go" Support

We need to modify the `callStmt()` method to consume multiple consecutive "go" tokens while preserving the same semantic meaning.

**Edit `parser.go`:**

Find line 985 where `p.next()` is called and add our multiple "go" logic right after it:

```go
// go/src/cmd/compile/internal/syntax/parser.go:982-990
s := new(CallStmt)
s.pos = p.pos()
s.Tok = p.tok // _Defer or _Go
p.next()

// Allow multiple consecutive "go" keywords (go go go ...)
if s.Tok == _Go {
    for p.tok == _Go {
        p.next()
    }
}

...
```

### Understanding the Code Change

- **`if s.Tok == _Go`**: Only apply multiple keyword logic to "go" statements (not "defer")
- **`for p.tok == _Go`**: Keep consuming "go" tokens while they appear consecutively
- **`p.next()`**: Advance past each additional "go" token
- **Preservation**: `s.Tok` remains `_Go`, so the semantic meaning is unchanged

## Step 3: Rebuild the Compiler

Now let's rebuild the Go toolchain with our changes:

```bash
cd ../../../  # back to go/src
./make.bash
```

If there are any compilation errors, review your changes and fix them.

## Step 4: Test Multiple "go" Keywords

Create a test program to verify our multiple "go" syntax works:

```bash
mkdir -p /tmp/multiple-go-test
cd /tmp/multiple-go-test
```

Create a test.go file:

```go
package main

import (
    "fmt"
    "time"
)

func sayHello(name string) {
    fmt.Printf("Hello from %s!\n", name)
}

func main() {
    fmt.Println("Testing multiple go keywords...")

    // Test regular single go
    go sayHello("single go")

    // Test double go
    go go sayHello("double go")

    // Test triple go
    go go go sayHello("triple go")

    // Test quadruple go
    go go go go sayHello("quadruple go")

    // Wait a bit to see output
    time.Sleep(100 * time.Millisecond)
    fmt.Println("All done!")
}
```

Execute the test program with your custom Go:

```bash
/path/to/workshop/go/bin/go run test.go
```

You should see output like:

```
Testing multiple go keywords...
Hello from single go!
Hello from double go!
Hello from triple go!
Hello from quadruple go!
All done!
```

## Step 5: Run Parser Tests

Let's make sure we didn't break the parser:

```bash
cd /path/to/workshop/go/src
../bin/go test cmd/compile/internal/syntax -short
```

## Understanding What We Did

1. **Parser Enhancement**: Modified `callStmt()` to handle multiple consecutive "go" tokens
2. **Token Consumption**: Added a loop to consume additional "go" tokens after the first one
3. **Semantic Preservation**: Multiple "go" keywords still create exactly one goroutine
4. **Targeted Change**: Only affects "go" statements, not "defer" statements

## What We Learned

- **Parser Logic**: How Go processes token sequences into statements
- **Token Consumption**: Techniques for consuming multiple tokens of the same type
- **Parser Testing**: Validating parser changes with diverse test cases

## Extension Ideas

Try these additional modifications:

1. Add similar support for "defer defer defer" (more challenging!)
2. Add a maximum limit (e.g., max 5 consecutive "go" keywords)
3. Track how many "go" keywords were used for debugging
4. Make the multiple keywords affect goroutine priority

## Next Steps

You've successfully enhanced Go's parser to handle repetitive syntax patterns.

In [Exercise 4: Compiler Inlining Parameters](./04-compiler-inlining-parameters.md), we'll shift focus to explore how Go's compiler optimization works, learning to tune inlining parameters for binary size control.

## Cleanup

To restore the original Go source:

```bash
cd /path/to/workshop/go/src/cmd/compile/internal/syntax
git checkout parser.go
cd ../../../
./make.bash  # Rebuild with original code
```

## Summary

Multiple "go" keywords now work for starting goroutines:

```go
// These are all equivalent and create exactly one goroutine:
go myFunction()
go go myFunction()
go go go myFunction() 
go go go go myFunction()

// The parser consumes all consecutive "go" tokens
// but the semantic behavior remains the same!
```

This exercise demonstrated how parser-level modifications can add expressive syntactic sugar while preserving the underlying language semantics.

---

*Continue to [Exercise 4](04-compiler-inlining-parameters.md) or return to the [main workshop](../README.md)*
