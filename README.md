# Having Fun with the Go Source Code Workshop

Welcome to an interactive workshop where you'll learn how to modify and experiment with the Go programming language source code! This hands-on workshop will guide you through understanding, building, and making changes to the Go compiler and runtime.

**This workshop uses Go version 1.26.1** - we'll check out the specific release tag to ensure consistency across all exercises.

## Prerequisites

- Basic knowledge of Go programming
- Familiarity with command line tools
- Git installed on your system
- **Go compiler version 1.24 or newer** (required for bootstrapping the build process)
- At least 4GB of free disk space

## Workshop Overview

This workshop consists of 10 exercises that will take you through the process from building Go from source, and making modifications at different places in the compiler, tooling and runtime. You'll gain some insights about the Go internals, from things like the lexer or parser, to runtime behaviors:

### [Exercise 0: Introduction and Setup](./exercises/00-introduction-setup.md)

Get started by cloning and setting up the Go source code environment.

### [Exercise 1: Compiling Go Without Changes](./exercises/01-compile-go-unchanged.md)

Learn to build the Go toolchain from source without any modifications.

### [Exercise 2: Adding the "=>" Arrow Operator for Goroutines](./exercises/02-scanner-arrow-operator.md)

Learn scanner/lexer modification by adding "=>" as an alternative syntax for starting goroutines.

### [Exercise 3: Multiple "go" Keywords - Parser Enhancement](./exercises/03-parser-multiple-go.md)

Learn parser modification by enabling multiple consecutive "go" keywords (go go go myFunction).

### [Exercise 4: Inline Parameters - Function Inlining Experiments](./exercises/04-compiler-inlining-parameters.md)

Explore the inliner behavior by modifying function inlining parameters.

### [Exercise 5: gofmt Modification - Indentation & AST Transformation](./exercises/05-gofmt-ast-transformation.md)

Modify gofmt to use 4 spaces instead of tabs and add a custom AST transformation replacing "hello" with "helo".

### [Exercise 6: SSA Pass - Detecting Division by Powers of Two](./exercises/06-ssa-power-of-two-detector.md)

Create a custom SSA compiler pass that detects division operations by powers of two that could be optimized to bit shifts.

### [Exercise 7: Patient Go - Making Go Wait for Goroutines](./exercises/07-runtime-patient-go.md)

Modify the Go runtime to wait for all goroutines to complete before program termination.

### [Exercise 8: Goroutine Sleep Detective - Runtime State Monitoring](./exercises/08-goroutine-sleep-detective.md)

Add logging to the Go scheduler to monitor goroutines going to sleep.

### [Exercise 9: Predictable Select - Removing Randomness from Go's Select Statement](./exercises/09-predictable-select.md)

Modify Go's select statement implementation to be deterministic instead of random.

### [Exercise 10: Java-Style Stack Traces - Making Go Panics Look Familiar](./exercises/10-java-style-stack-traces.md)

Transform Go's verbose stack traces into Java-style formatting.

## Getting Started

1. Start with [Exercise 0](./exercises/00-introduction-setup.md) to set up your environment
2. Work through the exercises in order
3. After exercise 1, you can pick and choose the exercise that you want.

## Repository Structure

```
.
├── README.md                 # This file
├── exercises/               # Individual exercise files (markdown)
│   ├── 00-introduction-setup.md
│   ├── 01-compile-go-unchanged.md
│   ├── 02-scanner-arrow-operator.md
│   └── ...
├── website-generator/       # Go program to generate website from markdown
│   ├── main.go
│   ├── templates.go
│   └── README.md
├── website/                 # Generated website (HTML)
│   ├── index.html
│   ├── 00-introduction-setup.html
│   └── ...
├── Makefile                 # Build automation
└── go/                      # Go source code (cloned during setup)
```

## Website Generator

This repository includes a Go program that automatically generates a static website from the markdown exercise files.

### Generate the Website

```bash
# Using make (recommended)
make website

# Or run directly
cd website-generator
go run . -exercises ../exercises -output ../website
```

### Serve Locally

```bash
# Start a local web server
make serve

# Then open http://localhost:8000 in your browser
```

The website generator:
- Converts markdown to HTML using [blackfriday](https://github.com/russross/blackfriday)
- Preserves all formatting, emojis, and code blocks
- Generates navigation between exercises
- Creates an index page with exercise overview
- Includes responsive CSS styling

See [website-generator/README.md](website-generator/README.md) for more details.

## Tips for Success

- Take your time with each exercise - compiler internals are complex!
- Don't hesitate to explore the Go source code beyond what's required
- Use `git` to track your changes and revert when needed
- Test your modifications thoroughly with various Go programs

## Resources

- [Go Compiler Overview](https://github.com/golang/go/tree/master/src/cmd/compile)
- [Go Language Specification](https://go.dev/ref/spec)
- [Go Runtime Documentation](https://pkg.go.dev/runtime)

### Video References

These workshop exercises are based on insights from my talks:

- [Understanding the Go Compiler](https://www.youtube.com/watch?v=qnmoAA0WRgE) - Deep dive into Go's compilation process
- [Understanding the Go Runtime](https://www.youtube.com/watch?v=YpRNFNFaLGY) - Exploration of Go's runtime system

## Workshop Completion

Upon completing all exercises, you'll have:

- **Built Go from source** and understood the bootstrap process
- **Modified language syntax** by changing scanner and parser behavior
- **Customized development tools** like gofmt and compiler optimizations
- **Implemented SSA optimizations** in the compiler backend
- **Modified runtime behavior** including program entry points and scheduler monitoring
- **Altered concurrency algorithms** like select statement randomization
- **Customized error reporting** with Java-style stack trace formatting

**Congratulations!** You'll have gained the confidence to keep exploring the Go source code. This knowledge enables you to:

- Start small contributions to the Go project
- Build custom language variants and tools
- Understand some trade-offs in language and runtime design

## Contributing

Found an issue, have an improvement idea or want to add more exercises? Please [open an issue](https://github.com/jespino/having-fun-with-the-go-source-code-workshop/issues) or submit a pull request!

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Happy coding and welcome to the world of Go internals!**
