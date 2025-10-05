# 🔍 Exercise 6: SSA Pass - Detecting Division by Powers of Two

In this exercise, you'll learn how Go's SSA (Static Single Assignment) compiler passes work by creating a custom optimization pass that detects division operations by powers of two.

## 🎯 Learning Objectives

By the end of this exercise, you will:

- ✅ Understand Go's SSA compiler pass architecture
- ✅ Know how to traverse SSA blocks and values
- ✅ Create a custom analysis pass from scratch
- ✅ Integrate your pass into the compiler pipeline
- ✅ Use SSA dumps to verify your pass works

## 🧠 Background: SSA Compiler Passes

The Go compiler transforms your code through multiple passes:

1. **Parse** - Convert source code to AST
2. **Type Check** - Verify types are correct
3. **IR Generation** - Convert to IR (Intermediate representation) form
3. **SSA Generation** - Convert to SSA (Static Single Assignment) form
4. **Optimization Passes** - Transform SSA (our focus!)
5. **Code Generation** - Produce machine code

We are going to work with the SSA form to know about the possibility of optimizing powers of two.

## 🔍 Step 1: Understanding SSA Pass Structure

SSA passes are registered in `compile.go` and operate on functions. Let's examine the structure:

```bash
cd go/src/cmd/compile/internal/ssa
```

Open `compile.go` and search for `var passes` (around line 457). You'll see:

```go
var passes = [...]pass{
	{name: "number lines", fn: numberLines, required: true},
	{name: "early phielim and copyelim", fn: copyelim},
	// ... many more passes
}
```

Each pass has:

- **name** - Displayed in debug output
- **fn** - Function that performs the transformation
- **required** - Whether this pass must run

## 🔧 Step 2: Create the Power of Two Detector Pass

Create a new file to hold our detector pass:

```bash
cd go/src/cmd/compile/internal/ssa
```

**Create `powoftwodetector.go`:**

```go
package ssa

import (
	"fmt"
	"math/bits"
)

func detectDivByPowerOfTwo(f *Func) {
	count := 0

	for _, b := range f.Blocks {
		for _, v := range b.Values {
			// Check for division operations
			if v.Op == OpDiv64 || v.Op == OpDiv32 || v.Op == OpDiv16 || v.Op == OpDiv8 ||
				v.Op == OpDiv64u || v.Op == OpDiv32u || v.Op == OpDiv16u || v.Op == OpDiv8u {

				// Check if the divisor (second argument) is a constant
				if len(v.Args) >= 2 {
					divisor := v.Args[1]

					// Check if it's a constant value
					if divisor.Op == OpConst64 || divisor.Op == OpConst32 ||
						divisor.Op == OpConst16 || divisor.Op == OpConst8 {

						constValue := divisor.AuxInt

						// Check if the constant is a power of two
						if isPowerOfTwo(constValue) {
							count++
							if f.pass.debug > 0 {
								fmt.Printf("  [PowerOfTwo] Found division by power of 2: %v / %d (could be >> %d) at %v\n",
									v.Args[0], constValue, bits.TrailingZeros64(uint64(constValue)), v.Pos)
							}
						}
					}
				}
			}
		}
	}

	if count > 0 {
		fmt.Printf("[PowerOfTwo Detector] Function %s: found %d division(s) by power of 2\n", f.Name, count)
	}
}
```

### 🔍 Understanding the Code

- **`f *Func`** - The SSA function being analyzed
- **`f.Blocks`** - All basic blocks in the function
- **`b.Values`** - All SSA values (operations) in a block
- **`v.Op`** - The operation type (division, addition, etc.)
- **`v.Args`** - Operands to the operation
- **`divisor.AuxInt`** - The constant value
- **`isPowerOfTwo()`** - Helper function that already exists in `rewrite.go`
- **`bits.TrailingZeros64()`** - Calculates how many bits to shift

## Step 3: Register the Pass in the Compiler

**Edit `compile.go`:**

Find the `var passes` array (around line 457) and add your pass as the **first** entry:

```go
var passes = [...]pass{
	{name: "detect div by power of two", fn: detectDivByPowerOfTwo, required: true},
	{name: "number lines", fn: numberLines, required: true},
	// ... rest of the passes
```

This runs your detector early in the pipeline, before other optimizations might eliminate the division.

## 📝 Step 4: Rebuild the Compiler

```bash
cd go/src
./make.bash
```

This compiles your new pass into the Go compiler.

## 🧪 Step 5: Create Test Programs

Create `test_divisions.go`:

```go
package main

func testDivisions() int {
	x := 100

	// These should be detected (powers of 2)
	a := x / 2   // 2 = 2^1, could be >> 1
	b := x / 4   // 4 = 2^2, could be >> 2
	c := x / 8   // 8 = 2^3, could be >> 3
	d := x / 16  // 16 = 2^4, could be >> 4

	// These should NOT be detected (not powers of 2)
	e := x / 3
	f := x / 5
	g := x / 7

	return a + b + c + d + e + f + g
}

func main() {
	result := testDivisions()
	println("Result:", result)
}
```

## Step 6: Run and See the Detection

```bash
../go/bin/go build test_divisions.go
```

**Expected output:**
```
[PowerOfTwo Detector] Function main.testDivisions: found 4 division(s) by power of 2
```

Your detector found the 4 divisions by powers of 2! 🎉

## Step 7: Test with Debug Output

For detailed information about each detection:

```bash
GOSSAFUNC=testDivisions ../go/bin/go build -gcflags="-d=ssa/detect div by power of two/debug=1" test_divisions.go
```

**Expected output:**
```
  [PowerOfTwo] Found division by power of 2: v10 / 2 (could be >> 1) at test_divisions.go:6
  [PowerOfTwo] Found division by power of 2: v14 / 4 (could be >> 2) at test_divisions.go:7
  [PowerOfTwo] Found division by power of 2: v18 / 8 (could be >> 3) at test_divisions.go:8
  [PowerOfTwo] Found division by power of 2: v22 / 16 (could be >> 4) at test_divisions.go:9
[PowerOfTwo Detector] Function main.testDivisions: found 4 division(s) by power of 2
```

This shows exact locations and shift amounts!

## 🎓 What We Learned

- 🏗️ **SSA Pass Architecture**: How to create and register compiler passes
- 🔍 **SSA Traversal**: Walking through blocks and values to analyze code
- 🎯 **Operation Detection**: Identifying specific SSA operations
- 📊 **Analysis vs Transformation**: Our pass analyzes but doesn't modify (yet!)

## 💡 Extension Ideas

Try these additional enhancements: 🚀

1. ➕ **Actually implement the optimization**: Replace division with shifts
2. ➕ **Detect multiplication by powers of 2**: Could use left shifts instead
3. ➕ **Count total optimizations**: Track how many across entire build
4. ➕ **Report efficiency gains**: Estimate cycle savings from the optimization

## 🧹 Cleanup

To remove your custom pass:

```bash
cd go/src/cmd/compile/internal/ssa
rm powoftwodetector.go
# Edit compile.go and remove your pass from the passes array
cd ../../src
./make.bash
```

## 📊 Summary

You've successfully created a custom SSA compiler pass that detects optimization opportunities!

```
Pass Name:     "detect div by power of two"
Input:         SSA function representation
Analysis:      Finds x / (power of 2) operations
Output:        Reports potential optimizations
Location:      Early in compiler pipeline

Example:       x / 8  →  Reports: "could be >> 3"
```

This demonstrates how Go's compiler infrastructure allows custom analysis and optimization passes. Real optimizations use the same patterns - they just modify the SSA instead of only reporting! 🚀✨

---

*Continue to [Exercise 7](07-runtime-patient-go.md) or return to the [main workshop](../README.md)*
