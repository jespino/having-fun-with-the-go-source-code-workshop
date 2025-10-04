# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

This is an educational workshop repository for learning Go compiler internals and source code modification. The workshop teaches participants how to build Go from source and make modifications to the Go compiler, scanner, parser, and runtime.

## Repository Architecture

### Workshop Structure
- **Root**: Contains main README.md with workshop overview and prerequisites
- **exercises/**: Individual exercise markdown files (00-10) with hands-on instructions
- **go/**: Go 1.25.1 source code (cloned during setup, ignored by git)
- **doing-it/**: Participant workspace directory (ignored by git)

### Exercise Flow
1. **Exercise 0**: Environment setup, Go source cloning, and repository exploration
2. **Exercise 1**: Building Go from source using bootstrap process
3. **Exercise 2**: First compiler modification - adding "gogogo" keyword as alternative to "go"
4. **Exercises 3-10**: Placeholder files for future advanced topics

### Key Workshop Requirements
- **Go Version**: Workshop uses Go 1.25.1 specifically (checked out via `git checkout go1.25.1`)
- **Bootstrap Requirement**: Participants need Go 1.24.6+ installed to build Go 1.25.1
- **Build Process**: Uses `./all.bash` (or `./make.bash` for build-only) in go/src directory
- **Target Audience**: Assumes basic Go knowledge but teaches compiler internals from scratch

## Common Workshop Commands

### Initial Setup (Exercise 0)
```bash
# Clone Go source (done by participants)
git clone https://go.googlesource.com/go
cd go
git checkout go1.25.1

# Verify environment
go version  # Must be 1.24.6+
git describe --tags  # Should show go1.25.1
```

### Building Go (Exercise 1)
```bash
cd go/src
./all.bash  # Full build with tests
# OR
./make.bash  # Build only, skip tests

# Test custom build
../bin/go version
../bin/go run test-program.go
```

### Compiler Modifications (Exercise 2+)
```bash
# Key Go compiler files for modification:
# - go/src/cmd/compile/internal/syntax/tokens.go (keyword definitions)
# - go/src/cmd/compile/internal/syntax/scanner.go (lexical analysis)
# - go/src/cmd/compile/internal/syntax/parser.go (parsing rules)

# After modifications, rebuild:
cd go/src
./make.bash

# Test modifications:
../bin/go run test-with-new-syntax.go
```

## Important File Locations

### Workshop Content
- `README.md`: Main workshop overview with prerequisites and structure
- `exercises/00-introduction-setup.md`: Setup instructions and Go source exploration  
- `exercises/01-compile-go-unchanged.md`: Bootstrap build process
- `exercises/02-scanner-go-go-go.md`: Scanner modification tutorial

### Go Source Code (go/ directory)
- `go/src/cmd/compile/`: Go compiler source
  - `internal/syntax/scanner.go`: Lexical tokenization
  - `internal/syntax/tokens.go`: Token and keyword definitions  
  - `internal/syntax/parser.go`: Syntax parsing
- `go/src/runtime/`: Go runtime system
- `go/bin/`: Compiled Go binaries (after build)

## Workshop-Specific Considerations

### Version Management
- Workshop is pinned to Go 1.25.1 for consistency
- All exercises assume this specific version
- Participants must checkout this tag before starting

### Educational Approach  
- Uses emojis throughout for friendly, approachable documentation
- Exercises build progressively from basic compilation to advanced modifications
- Includes comprehensive troubleshooting sections
- Emphasizes understanding over just following instructions

### File Modification Workflow
1. Read and understand existing code structure
2. Make targeted modifications (usually in syntax/ directory)
3. Rebuild Go toolchain
4. Test modifications with sample programs
5. Validate that normal Go code still works

### References
Based on insights from:
- "Understanding the Go Compiler" (YouTube: qnmoAA0WRgE)
- "Understanding the Go Runtime" (YouTube: YpRNFNFaLGY)

## Working with This Repository

When creating new exercises or modifying existing ones:
- Maintain the emoji-friendly tone established in current exercises
- Include comprehensive prerequisites and verification steps
- Provide concrete code examples and expected outputs
- Add troubleshooting sections for common issues
- Ensure exercises build logically on previous ones

When updating workshop content, remember that participants will have varying levels of compiler knowledge, so explanations should be thorough but not overwhelming.