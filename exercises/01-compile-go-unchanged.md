# 🔨 Exercise 1: Compiling Go Without Changes

In this exercise, you'll learn how to build the Go toolchain from the source code without making any modifications. This is an essential skill before we start making changes to the language! ⚡

## 🎯 Learning Objectives

By the end of this exercise, you will:

- ✅ Understand the Go build process and bootstrap concept
- ✅ Successfully compile Go from source
- ✅ Know how to explore Go source code structure
- ✅ Know how to test your custom Go build

## 🥾 Step 1: Understanding the Bootstrap Process

Go is written in Go itself! 🤔 This creates a "chicken and egg" problem - how do you compile Go without having Go? The solution is bootstrapping:

1. 📦 The Go team provides pre-compiled binaries
2. 🔨 These binaries compile the current Go source code
3. ✨ The newly compiled version can then be used for development

Let's check if you have Go installed (needed for bootstrapping): 🔍

```bash
go version
# Must show version 1.24.6 or newer
```

**⚠️ Critical**: You must have Go 1.24.6 or newer installed to build Go 1.25.1. If you don't have Go installed or your version is too old, install the latest version from <https://golang.org/dl/> 📥

## 📂 Step 2: Navigate to the Go Source Directory

```bash
cd go/src
pwd
# Should show: /path/to/workshop/go/src

# Verify you're on the correct Go version
git describe --tags
# Should show: go1.25.1
```

## 🚀 Step 3: Start the Build Process

Go provides different scripts for building. Let's start with `make.bash` which builds the toolchain, then explore the source code while it's running! ⚡

**On Unix-like systems (Linux, macOS):**

```bash
./make.bash
```

**On Windows:**

```cmd
make.bat
```

This script will: 📋

1. 🔨 Build the Go toolchain (compiler, linker, runtime, standard library)
2. ⏱️ Take approximately 2-10 minutes depending on your system

**📝 Note:** The first time you run this, it will take longer as it needs to compile everything from scratch.

### 🤔 What about `all.bash` and `run.bash`?

You might wonder about other scripts in the `src/` directory:

- **🔨 `make.bash`**: Builds the Go toolchain only (what we're using)
- **🧪 `run.bash`**: Runs the comprehensive test suite (requires Go to be built first)  
- **📦 `all.bash`**: Convenience script that runs `make.bash` + `run.bash` + prints build info

For this workshop, `make.bash` is perfect because:

- ✅ Faster build time means less waiting
- ✅ We just need a working Go build for our experiments
- ✅ We can run tests later if needed with `run.bash`

## 🔍 Step 4: Explore Source Code While Building

While the build is running, open a **new terminal** or **IDE** and let's explore the Go source code structure! This is a great time to understand what we're building. 🧭

**In your new terminal:**

```bash
cd /path/to/workshop/go  # Navigate to your Go source directory
ls -la
```

### 📁 Repository Structure

Key directories you should see:

- **📁 `src/`**: Contains the Go source code
  - 🛠️ `src/cmd/`: Command-line tools (go, gofmt, etc.)
  - ⚙️ `src/runtime/`: Go runtime system
  - 🌳 `src/go/`: Go language packages (parser, AST, etc.)
- **🧪 `test/`**: Test files for the Go language
- **📋 `api/`**: API compatibility data
- **📚 `doc/`**: Documentation

### 🏗️ Examine the Go Compiler Structure

The Go compiler is located in `src/cmd/compile/`. Let's explore it: 🔧

```bash
cd src/cmd/compile
ls -la
```

Key files and directories:

- **🚪 `main.go`**: Entry point of the compiler
- **📦 `internal/`**: Internal compiler packages
  - 🔤 `internal/syntax/`: Lexer/parser (scanner, parser)
  - ✅ `internal/types2/`: Type checker
  - 🔄 `internal/ir/`: Intermediate representation
  - ⚡ `internal/gc/`: Code generation

## 📊 Step 5: Understanding the Build Output

**Switch back to your original terminal** where the build is running. As the build progresses, you should see output like: 👀

```
Building Go cmd/dist using /usr/local/go. (go1.25.1 darwin/amd64)
Building Go toolchain1 using /usr/local/go.
Building Go bootstrap cmd/go (go_bootstrap) using Go toolchain1.
Building Go toolchain2 using go_bootstrap and Go toolchain1.
Building Go toolchain3 using go_bootstrap and Go toolchain2.
Building packages and commands for darwin/amd64.
```

This shows the multi-stage bootstrap process:

- The compiler is build with the go version installed in your system (toolchain1)
- Then the compiler is built again using the toolchain1 to produce the toolchain2
- Finally the toolchain3 is generated using the toolchain2.
- The toolchain3 and toolchain2 should be identical

## 📍 Step 6: Locate Your Compiled Go Binary

After successful compilation, your new Go binary will be in: 🎯

```bash
ls -la /path/to/workshop/go/bin
```

You should see:

- 🚀 `go` - The main Go command
- 🎨 `gofmt` - Go formatter
- 🛠️ Other Go tools

## 🧪 Step 7: Test Your Custom Go Build

Let's test your newly compiled Go: ✨

```bash
# Check version of your compiled Go
../bin/go version
```

Create a hello.go in the a temporary directory, for example `/tmp`.

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello from my custom Go build!")
}
```

```bash
# Compile and run with your custom Go
/path/to/workshop/go/bin/go run /tmp/hello.go
```

## 🎓 What We Learned

- 🥾 **Bootstrap Process**: Go compiles itself using an existing Go installation
- 🏗️ **Go Source Structure**: Well-organized codebase with clear separation (cmd/, runtime/, etc.)
- ⚡ **Build Process**: `./make.bash` builds everything

## 🎉 Next Steps

Congratulations! 🎊 You now have a working Go toolchain built from source.

You can now proceed with any of the following exercises to learn about different parts of Go:

- [Exercise 2: Adding the "=>" Arrow Operator for Goroutines](./02-scanner-arrow-operator.md) - Scanner modifications
- [Exercise 3: Multiple "go" Keywords - Parser Enhancement](./03-parser-multiple-go.md) - Parser modifications
- [Exercise 4: Inline Parameters - Function Inlining Experiments](./04-compiler-inlining-parameters.md) - Compiler parameters
- [Exercise 5: gofmt Transformation - "hello" to "helo"](./05-gofmt-ast-transformation.md) - AST transformations
- [Exercise 6: SSA Pass - Detecting Division by Powers of Two](./06-ssa-power-of-two-detector.md) - SSA compiler passes
- [Exercise 7: Patient Go - Making Go Wait for Goroutines](./07-runtime-patient-go.md) - Runtime modifications
- [Exercise 8: Goroutine Sleep Detective - Runtime State Monitoring](./08-goroutine-sleep-detective.md) - Scheduler monitoring
- [Exercise 9: Predictable Select - Removing Randomness from Go's Select Statement](./09-predictable-select.md) - Select behavior
- [Exercise 10: Java-Style Stack Traces - Making Go Panics Look Familiar](./10-java-style-stack-traces.md) - Error formatting

Or return to the [main workshop](../README.md) to choose an exercise.
