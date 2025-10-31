# ⚡ Exercise 2: Adding the "=>" Arrow Operator for Goroutines

In this exercise, you'll add a new "=>" arrow operator to Go that works as an alternative syntax for starting goroutines! 🚀 This will teach you how to modify Go's scanner to recognize new operators and map them to existing functionality.

## 🎯 Learning Objectives

By the end of this exercise, you will:

- ✅ Understand how Go's scanner tokenizes operators
- ✅ Know how to add new operator syntax to Go
- ✅ Modify the scanner's lexical analysis logic
- ✅ Test your scanner modification with working code
- ✅ Successfully extend Go's operator vocabulary

## 🧠 Background: How This Scanner Modification Works

This exercise demonstrates **scanner-level modifications** to add new operator syntax to Go. We'll modify the scanner logic to recognize a new operator sequence "=>" and map it to an existing token. Here's what we'll accomplish:

- **Scanner Enhancement**: Add recognition for the "=>" operator sequence
- **Token Mapping**: Map "=>" to the existing `_Go` token (same as the "go" keyword)
- **Alternative Syntax**: Create `=> myFunction()` as equivalent to `go myFunction()`
- **Minimal Impact**: No parser or compiler changes needed - just scanner logic

This approach allows us to create elegant alternative syntax without changing the deeper parts of the compiler!

## 🔍 Step 1: Navigate to the Scanner

```bash
cd go/src/cmd/compile/internal/syntax
```

### 🔑 Understanding the Scanner Structure

Let's examine how the scanner handles the "=" operator in `scanner.go`. Look at line 325:

```go
// go/src/cmd/compile/internal/syntax/scanner.go:325
case '=':
    if s.ch == '=' {
        s.nextch()
        s.tok = _Operator
        break
    }
    s.tok = _Assign
```

The scanner checks for "==" (equals comparison) first, then falls back to "=" (assignment).

## Step 2: Add the Arrow Operator Logic

We need to add logic to recognize "=>" and treat it as the `_Go` token.

**Edit `scanner.go`:**

Find the "=" case at line 325 and modify it to also check for ">":

```go
// go/src/cmd/compile/internal/syntax/scanner.go:325
case '=':
    if s.ch == '=' {
        s.nextch()
        s.tok = _Operator
        break
    }
    if s.ch == '>' {
        s.nextch()
        s.lit = "=>"
        s.tok = _Go
        break
    }
    s.tok = _Assign
```

### 🔧 Understanding the Code Change

- **`if s.ch == '>'`**: Check if the next character after "=" is ">"
- **`s.nextch()`**: consumes the ">" character from the lexer
- **`s.lit = "=>"`**: Set the literal value for debugging/error messages
- **`s.tok = _Go`**: Assign the same token as the "go" keyword
- **`break`**: Exit the case to avoid falling through to `_Assign`

## Step 3: Rebuild the Compiler

Now let's rebuild the Go toolchain with our changes:

```bash
cd ../../../  # back to go/src
./make.bash
```

If there are any compilation errors, review your changes and fix them.

## Step 4: Test the New Arrow Operator

Create a test program to verify our new "=>" operator works:

```bash
mkdir -p /tmp/arrow-test
cd /tmp/arrow-test
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
    fmt.Println("Testing => arrow operator...")

    // Test regular go keyword
    go sayHello("regular go")

    // Test our new => operator
    => sayHello("arrow operator")

    // Wait a bit to see output
    time.Sleep(100 * time.Millisecond)
    fmt.Println("All done!")
}
```

Execute the test program with your custom Go:

```bash
/path/to/workshop/go/bin/go run test.go
```

You should see output like: ✨

```
Testing => arrow operator...
Hello from regular go!
Hello from arrow operator!
All done!
```

## 🧪 Step 5: Test Mixed Go Operators

Let's test mixed scenarios using both the traditional "go" keyword and our new "=>" arrow operator:

Create a mixed-test.go file:

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    fmt.Printf("Worker %d starting\n", id)
    time.Sleep(50 * time.Millisecond)
    fmt.Printf("Worker %d done\n", id)
}

func main() {
    var wg sync.WaitGroup

    fmt.Println("Starting workers with mixed syntax...")

    // Mix of regular go and => operators
    for i := 1; i <= 4; i++ {
        wg.Add(1)
        if i%2 == 0 {
            go worker(i, &wg)  // Regular go
        } else {
            => worker(i, &wg)  // Arrow operator
        }
    }

    wg.Wait()
    fmt.Println("All workers completed!")
}
```

Execute the mixed test program:

```bash
/path/to/workshop/go/bin/go run mixed-test.go
```

## Step 6: Run Scanner Tests

Let's make sure we didn't break the scanner:

```bash
cd /path/to/workshop/go/src
../bin/go test cmd/compile/internal/syntax -short
```

## Understanding What We Did

1. **Modified Scanner Logic**: Added "=>" recognition to the existing "=" case
2. **Reused Existing Token**: Mapped "=>" to `_Go` token instead of creating new token
3. **Preserved Existing Functionality**: "=" and "==" operators still work normally
4. **Minimal Change Impact**: No parser or IR changes needed

## 🎓 What We Learned

- 🔤 **Scanner Logic**: How Go tokenizes operator sequences
- 📝 **Operator Recognition**: Adding new operators through scanner modification
- 🔄 **Token Reuse**: Mapping new syntax to existing tokens
- 🧪 **Testing Strategy**: Validating scanner changes with real code
- 🔨 **Build Process**: Rebuilding Go with scanner modifications

## 💡 Extension Ideas

Try these additional modifications: 🚀

1. ➕ Add ":>" as another alternative to "go"
2. ➕ Add "~>" for async operations
3. 🔤 Add ">>>" as a triple-arrow operator
4. 🎨 Make the arrow operator work in different contexts

## ➡️ Next Steps

Great work! 🎉 You've successfully added a new operator to Go's scanner. You now understand how to modify the scanner to create alternative syntax for existing functionality. This technique can be applied to create other operator shortcuts and syntax sugar in the language.

In Exercise 3, we'll take a different approach and explore **parser modifications** - learning how to modify the parser to handle multiple consecutive tokens.

## Cleanup

To restore the original Go source:

```bash
cd /path/to/workshop/go/src/cmd/compile/internal/syntax
git checkout scanner.go
cd ../../../
./make.bash  # Rebuild with original code
```

## Summary

The "=>" arrow operator now works as an alternative to "go" for launching goroutines:

```go
// These are now equivalent:
go myFunction()
=> myFunction()

// Both create goroutines the same way!
```

This exercise demonstrated how scanner-level modifications can add new syntax with minimal code changes! 🚀✨

---

*Continue to [Exercise 3](03-parser-multiple-go.md) or return to the [main workshop](../README.md)*
