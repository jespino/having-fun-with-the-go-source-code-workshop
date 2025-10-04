# ⚡ Exercise 5: gofmt Modification - Indentation & AST Transformation

In this exercise, you'll modify Go's formatting tool `gofmt` to use 4 spaces instead of tabs, and then add a custom AST transformation to automatically replace the word "hello" with "helo" in string literals and comments! 🔄 This will teach you how Go's formatter works, how printer modes control indentation, and how to add custom transformations to the AST processing pipeline.

## 🎯 Learning Objectives

By the end of this exercise, you will:

- ✅ Understand how gofmt controls indentation and printer modes
- ✅ Learn to modify formatting behavior across gofmt and go/format package
- ✅ Understand how gofmt processes Go source code through AST manipulation
- ✅ Know how to modify string literals and comments in the AST
- ✅ Explore Go's AST (Abstract Syntax Tree) structure
- ✅ Create custom source code transformations

## 🧠 Background: How gofmt Works

gofmt operates through these stages:

1. **Parse** → Convert source code to AST (Abstract Syntax Tree)
2. **Transform** → Apply formatting rules to AST
3. **Print** → Convert modified AST back to formatted source code with specific indentation

The indentation behavior is controlled by two key constants:

- **`tabWidth`** → Width of indentation (default: 8)
- **`printerMode`** → Flags controlling spacing behavior:
  - `printer.UseSpaces` → Use spaces for padding
  - `printer.TabIndent` → Use tabs for indentation
  - `printerNormalizeNumbers` → Normalize number literals

### 🌳 AST Structure

Go represents source code as a tree of nodes we are going to use here this two nodes:

- **`*ast.BasicLit`** → String literals, numbers, etc.
- **`*ast.Comment`** → Comments in source code

## 🔍 Step 1: Navigate to gofmt Source

```bash
cd go/src/cmd/gofmt
ls -la
```

Key files:

- **`gofmt.go`** → Main program logic and file processing
- **`simplify.go`** → AST simplification transformations

## 📏 Step 2: Change Indentation to 4 Spaces

Before adding custom transformations, let's change gofmt to use 4 spaces instead of tabs for indentation.

### Modify gofmt.go

**Edit `go/src/cmd/gofmt/gofmt.go`:**

Find the constants around line 47 (look for the comment "Keep these in sync with go/format/format.go"):

```go
const (
	tabWidth    = 8
	printerMode = printer.UseSpaces | printer.TabIndent | printerNormalizeNumbers
```

Change to:

```go
const (
	tabWidth    = 4
	printerMode = printer.UseSpaces | printerNormalizeNumbers
```

**What changed:**

- **`tabWidth`**: Changed from `8` to `4` (4 spaces per indentation level)
- **`printerMode`**: Removed `printer.TabIndent` flag (this removes tab characters and uses spaces only)

### Modify go/format Package

The `go/format` package also needs to be updated to keep behavior consistent.

**Edit `go/src/go/format/format.go`:**

Find the constants around line 28 (same comment as above):

```go
const (
	tabWidth    = 8
	printerMode = printer.UseSpaces | printer.TabIndent | printerNormalizeNumbers
```

Change to:

```go
const (
	tabWidth    = 4
	printerMode = printer.UseSpaces | printerNormalizeNumbers
```

### 🔧 Understanding the Changes

- **`tabWidth = 4`**: Each indentation level uses 4 spaces
- **Removing `TabIndent`**: Without this flag, the printer uses only spaces (no tab characters)
- **`UseSpaces`**: Ensures spaces are used for padding and alignment
- **Both files must match**: gofmt and go/format must use the same settings for consistency

## 🔨 Step 3: Rebuild and Test Indentation

```bash
cd ../../../  # back to go/src
./make.bash
```

Create a test file `indent_test.go`:

```go
package main

import "fmt"

func main() {
	if true {
		for i := 0; i < 10; i++ {
			fmt.Println(i)
		}
	}
}
```

Test the new indentation:

```bash
cd ..  # to go/ directory
./bin/gofmt indent_test.go
```

Expected output (notice 4 spaces for each level):

```go
package main

import "fmt"

func main() {
    if true {
        for i := 0; i < 10; i++ {
            fmt.Println(i)
        }
    }
}
```

🎉 Each indentation level now uses 4 spaces instead of tabs!

## Step 4: Add Hello→Helo Transformation

**Edit `gofmt.go`:**

Add this transformation function around line 75 (after the `usage()` function):

```go
// transformHelloToHelo walks the AST and replaces "hello" with "helo"
// in string literals and comments.
func transformHelloToHelo(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BasicLit:
			// Handle string literals
			if node.Kind == token.STRING {
				if strings.Contains(node.Value, "hello") {
					node.Value = strings.ReplaceAll(node.Value, "hello", "helo")
				}
			}
		case *ast.Comment:
			// Handle comments
			if strings.Contains(node.Text, "hello") {
				node.Text = strings.ReplaceAll(node.Text, "hello", "helo")
			}
		}
		return true // continue traversing
	})
}
```

### 🔧 Understanding the Code

- **`ast.Inspect()`** - Traverses all nodes in the AST
- **`*ast.BasicLit`** - Matches string literals
- **`node.Kind == token.STRING`** - Checks if it's a string (not a number)
- **`*ast.Comment`** - Matches comments
- **`strings.ReplaceAll()`** - Performs the replacement

## Step 5: Integrate the Transformation

**Still in `gofmt.go`:**

Find the `processFile` function around line 256. Look for:

```go
	if *simplifyAST {
		simplify(file)
	}
```

Add our transformation right after:

```go
	if *simplifyAST {
		simplify(file)
	}

	// Apply our custom hello→helo transformation
	transformHelloToHelo(file)
```

## Step 6: Rebuild gofmt

```bash
cd ../../../  # back to go/src
./make.bash
```

## Step 7: Test Both Modifications Together

Create a `hello_test.go` file:

```go
package main

import "fmt"

func main() {
    // Say hello to everyone
    message := "hello world"
    greeting := "Say hello!"

    /* This is a hello comment block */
    fmt.Println(message)
    fmt.Println(greeting)

    // Another hello comment
    fmt.Printf("hello %s\n", "Go")
}
```

```bash
../go/bin/gofmt hello_test.go
```

Expected output (notice both 4-space indentation AND hello→helo transformation):

```go
package main

import "fmt"

func main() {
    // Say helo to everyone
    message := "helo world"
    greeting := "Say helo!"

    /* This is a helo comment block */
    fmt.Println(message)
    fmt.Println(greeting)

    // Another helo comment
    fmt.Printf("helo %s\n", "Go")
}
```

🎉 Two changes applied:

1. All "hello" instances are replaced with "helo"
2. Indentation uses 4 spaces instead of tabs

## Step 8: Test In-Place Formatting

```bash
# Format and overwrite the file
../go/bin/gofmt -w hello_test.go

# Verify the changes
cat hello_test.go
```

The file is now permanently transformed with "helo" instead of "hello" and using 4-space indentation!

## Understanding What We Did

1. **Modified Printer Settings**: Changed tabWidth and printerMode to use 4 spaces
2. **Synced Two Packages**: Updated both gofmt and go/format for consistency
3. **Added AST Visitor**: Created function to traverse and modify AST nodes
4. **Pattern Matching**: Identified string literals and comments
5. **Text Replacement**: Modified node values to replace "hello" with "helo"
6. **Integration**: Called transformation during gofmt processing
7. **Testing**: Verified both indentation and transformation changes

## 🎓 What We Learned

- 📏 **Printer Configuration**: How gofmt controls indentation through tabWidth and printerMode
- 🔄 **Package Consistency**: Why gofmt and go/format must stay in sync
- 🌳 **AST Manipulation**: How to traverse and modify Go's Abstract Syntax Tree
- 🔧 **Tool Modification**: How to extend existing Go tools with multiple changes
- 🔍 **Code Transformation**: Implementing systematic source code changes
- 🏗️ **Build Process**: Rebuilding Go toolchain components
- 🧪 **Testing**: Verifying custom tool behavior

## 💡 Extension Ideas

Try these additional modifications: 🚀

1. ➕ Add a command-line flag to enable/disable the transformation
2. ➕ Support multiple word replacements (hello→helo, world→universe)
3. ➕ Add case-sensitive option
4. ➕ Only replace whole words (not substrings within words)
5. ➕ Make tabWidth configurable via command-line flag
6. ➕ Add option to switch between tabs and spaces

Example flag addition:
```go
var replaceHello = flag.Bool("helo", false, "replace hello with helo")

// In processFile():
if *replaceHello {
    transformHelloToHelo(file)
}
```

## Cleanup

To restore the original gofmt:

```bash
cd go/src/cmd/gofmt
git checkout gofmt.go
cd ../go/format
git checkout format.go
cd ../../../src
./make.bash
```

## Summary

You've successfully modified gofmt in two powerful ways!

```
Indentation:   tabs (8 width) → 4 spaces
Transformation: "hello world"  → "helo world"
                // Say hello    → // Say helo

Changes:  tabWidth=4 + remove TabIndent flag
         + ast.Inspect() → pattern match → replace text
```

You now understand how tools like `gofmt`, `goimports`, and `go fix` work at both the printer and AST levels! ⚡🌳

---

*Continue to [Exercise 6](06-ssa-power-of-two-detector.md) or return to the [main workshop](../README.md)*
