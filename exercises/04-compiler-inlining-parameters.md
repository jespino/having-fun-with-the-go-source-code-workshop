# ⚡ Exercise 4: Compiler Inlining Parameters - Tuning for Binary Size Control

In this exercise, you'll explore and modify Go's inlining parameters to see their dramatic effects on binary size! 🎛️ This will teach you how Go's compiler decides when to inline functions and how tweaking these parameters can significantly change your compiled programs.

## 🎯 Learning Objectives

By the end of this exercise, you will:

- ✅ Understand Go's inlining budget system and parameters
- ✅ Know where inlining decisions are made in the compiler
- ✅ Modify inlining thresholds to control optimization behavior
- ✅ Measure the impact on binary size

## 🧠 Background: Function Inlining in Go

Function inlining is a compiler optimization where function calls are replaced by the actual function body. This trades binary size for performance:

**Benefits:**

- ⚡ Eliminates call overhead
- 🎯 Enables further optimizations at call site
- 🚀 Better instruction pipeline utilization

**Costs:**

- 📦 Larger binary size
- 💾 Increased memory usage (for the program)

Go uses a sophisticated **budget system** to decide when inlining is profitable!

## 🔍 Step 1: Understanding Go's Inlining Budget

Let's examine the current inlining parameters:

```bash
cd go/src/cmd/compile/internal/inline
```

Open `inl.go` and look at the key parameters around lines 48-86:

### 🎛️ Key Inlining Parameters

From `go/src/cmd/compile/internal/inline/inl.go:48-86`:

```go
const (
    inlineMaxBudget       = 80    // Maximum "cost" for inlining
    inlineExtraAppendCost = 0     // Extra cost for append operations
    inlineExtraCallCost   = 57    // Cost penalty for function calls
    inlineParamCallCost   = 17    // Reduced cost when calling a parameter
    inlineExtraPanicCost  = 1     // Low cost for panic calls
    inlineExtraThrowCost  = 80    // High cost discourages inlining runtime.throw

    // Function size thresholds
    inlineBigFunctionNodes      = 5000  // Nodes that make a function "big"
    inlineBigFunctionMaxCost    = 20    // Max inline cost into "big" functions
    inlineClosureCalledOnceCost = 800   // Special budget for single-use closures

    // PGO (Profile Guided Optimization) parameters
    inlineHotMaxBudget = 2000   // Much larger budget for "hot" functions
)
```

### 🔬 How the Budget System Works

Each Go statement/expression has a **cost**:

- Simple statements: 1 point
- Function calls: 57+ points
- Loops, conditions: 1 point each
- Complex expressions: Variable points

The compiler sums up costs and compares against the budget!

## 📊 Step 2: Use Go Compiler Binary for Size Comparison

Instead of creating toy programs, let's use the Go compiler binary itself as our test subject! The Go compiler (`bin/go`) is perfect for demonstrating inlining effects because:

- 🏗️ **Large codebase** - Shows meaningful size differences
- 🔧 **Real-world code** - Contains the actual patterns we're optimizing
- 🎯 **Workshop relevance** - We're building it throughout the exercises
- 📊 **Dramatic results** - Large enough to show significant inlining impact

### 🎯 Test Different Inlining Settings on Go Binary

Let's rebuild the entire Go toolchain with different inlining settings and compare the `bin/go` binary sizes:

```bash
cd go/src
```

### 📊 Baseline Build - Default Settings

First, let's build with default inlining settings and backup the binary:

```bash
# Build with default settings
./make.bash

# Copy the default Go binary for comparison
cp ../bin/go ../bin/go-default

# Check the size
ls -lh ../bin/go-default
wc -c ../bin/go-default
```

### 🔬 Check Current Inlining Impact on Go Compiler Build

We can examine how inlining affects the Go compiler itself during compilation:

```bash
# See inlining decisions when compiling the Go compiler
# This shows how inlining parameters affect the compiler's own build process
cd cmd/compile
../../bin/go build -gcflags="-m" . 2>&1 | grep "can inline" | wc -l
echo "Functions that can be inlined during Go compiler build"
```

## ⚙️ Step 3: Modify Inlining Parameters

Now let's modify the inlining parameters to see their effects!

### 🔧 Experiment 1: Aggressive Inlining

Edit `go/src/cmd/compile/internal/inline/inl.go` around line 50:

```go
const (
    inlineMaxBudget       = 200   // Increased from 80
    inlineExtraCallCost   = 20    // Decreased from 57
    inlineBigFunctionMaxCost = 50 // Increased from 20
)
```

**Rebuild the compiler:**

```bash
cd go/src
./make.bash
```

**Test aggressive inlining on Go binary:**

```bash
# Copy the aggressively-inlined Go binary
cp ../bin/go ../bin/go-aggressive

# Compare sizes
echo "Default size: $(wc -c < ../bin/go-default)"
echo "Aggressive size: $(wc -c < ../bin/go-aggressive)"

# Calculate size difference
default_size=$(wc -c < ../bin/go-default)
aggressive_size=$(wc -c < ../bin/go-aggressive)
echo "Size difference: $(($aggressive_size - $default_size)) bytes"
echo "Percentage increase: $(echo "scale=2; ($aggressive_size - $default_size) * 100 / $default_size" | bc)%"
```

### 🔧 Experiment 2: Conservative Inlining

Now try conservative settings. Edit the parameters:

```go
const (
    inlineMaxBudget       = 40    // Decreased from 80
    inlineExtraCallCost   = 100   // Increased from 57
    inlineBigFunctionMaxCost = 5  // Decreased from 20
)
```

**Rebuild and test:**

```bash
cd go/src
./make.bash

# Copy the conservatively-inlined Go binary
cp ../bin/go ../bin/go-conservative

# Compare all three Go binaries
echo "Conservative size: $(wc -c < ../bin/go-conservative)"
echo "Default size: $(wc -c < ../bin/go-default)"
echo "Aggressive size: $(wc -c < ../bin/go-aggressive)"
```

## 📊 Step 4: Comprehensive Binary Size Analysis

Let's test extreme inlining settings to see dramatic effects on the Go compiler binary:

### 🔧 Experiment 3: No Inlining At All

For comparison, let's disable inlining entirely:

```go
const (
    inlineMaxBudget       = 0     // No inlining budget
    inlineExtraCallCost   = 1000  // Prohibitive call cost
    inlineBigFunctionMaxCost = 0  // No big function inlining
)
```

```bash
cd go/src
./make.bash

# Copy the no-inlining Go binary
cp ../bin/go ../bin/go-no-inline
```

### 🔧 Experiment 4: Extreme Inlining - Breaking Point Demonstration

Let's try extremely aggressive settings to see what happens when we push inlining too far:

```go
const (
    inlineMaxBudget       = 500   // Very high budget
    inlineExtraCallCost   = 5     // Very low call cost
    inlineBigFunctionMaxCost = 200 // Very high big function budget
)
```

```bash
cd go/src
./make.bash
```

**⚠️ Expected Result:** This will likely fail to compile or produce a broken compiler! This demonstrates that there are limits to how aggressive inlining can be. You may see compilation errors or the build may hang.

If it fails (which is expected), you'll learn that:
- Extreme inlining can cause compilation to fail
- There are practical limits to compiler optimizations
- The default parameters are carefully balanced for good reason

## 📋 Step 5: Analyze Results

Compare the Go compiler binary sizes:

```bash
cd go

echo "=== GO COMPILER BINARY SIZE COMPARISON ==="
echo "No Inlining:  $(wc -c < bin/go-no-inline) bytes"
echo "Conservative: $(wc -c < bin/go-conservative) bytes"
echo "Default:      $(wc -c < bin/go-default) bytes"
echo "Aggressive:   $(wc -c < bin/go-aggressive) bytes"

echo ""
echo "=== SIZE DIFFERENCES ==="
no_inline_size=$(wc -c < bin/go-no-inline)
conservative_size=$(wc -c < bin/go-conservative)
default_size=$(wc -c < bin/go-default)
aggressive_size=$(wc -c < bin/go-aggressive)

echo "No-inline vs Default: $(($default_size - $no_inline_size)) bytes difference"
echo "Default vs Aggressive: $(($aggressive_size - $default_size)) bytes difference"
echo "Full Range (No-inline to Aggressive): $(($aggressive_size - $no_inline_size)) bytes difference"

# Calculate percentages
echo ""
echo "=== PERCENTAGE DIFFERENCES ==="
echo "Aggressive vs Default: $(echo "scale=2; ($aggressive_size - $default_size) * 100 / $default_size" | bc)%"
echo "Default vs No-inline: $(echo "scale=2; ($default_size - $no_inline_size) * 100 / $no_inline_size" | bc)%"
```


## 🔧 Understanding What We Modified

### 🎛️ Key Parameter Functions

| Parameter | Purpose | Impact |
|-----------|---------|--------|
| `inlineMaxBudget` | Maximum cost for any inlined function | Higher = more inlining |
| `inlineExtraCallCost` | Penalty for function calls inside inlined functions | Lower = more aggressive |
| `inlineBigFunctionMaxCost` | Max cost when inlining into large functions | Higher = more inlining in big funcs |
| `inlineBigFunctionNodes` | Threshold for "big" function detection | Lower = more functions considered "big" |

### 📊 Typical Results You Should See

With the Go compiler binary, you should observe dramatic size differences based on actual measurements:

- **No Inlining**: Smallest binary (~25MB / 25,176 KB)
- **Conservative**: Small binary (~26MB / 25,968 KB)
- **Default**: Balanced size (~27MB / 27,544 KB)
- **Aggressive**: Largest binary (~36MB / 35,904 KB) - **30% larger than default!**

**Key Insights:**

- **Aggressive inlining** can increase binary size by **8+ MB** (30% larger)
- **No inlining vs Default** shows a **2+ MB difference** (8% smaller)

The exact sizes depend on your system, but you should see similar dramatic differences!

## 🎓 What We Learned

- 🎛️ **Budget System**: How Go uses cost-based analysis for inlining decisions
- 📊 **Parameter Impact**: How different settings affect binary size and performance
- 🔬 **Measurement Techniques**: Using debug flags to understand compiler decisions
- ⚖️ **Trade-offs**: The fundamental tension between binary size and performance
- 🛠️ **Compiler Tuning**: How to modify compiler behavior for specific needs

## 💡 Extension Ideas

Try these additional experiments: 🚀

1. 📈 Create a script to automate testing different parameter combinations
2. 🎯 Test with real-world Go programs (like building Go itself!)
3. 📊 Measure compilation time differences with various settings
4. 🌡️ Experiment with PGO (Profile-Guided Optimization) parameters
5. 🔬 Analyze assembly output differences between inlined and non-inlined calls

## ➡️ Next Steps

Excellent work! 🎉 You've learned how to tune Go's inlining behavior and seen its real-world impact on binary size and performance. In the next exercises, we'll explore modifiying the gofmt tool.

## 🧹 Cleanup

To restore original inlining parameters and clean up test binaries:

```bash
cd go/src/cmd/compile/internal/inline
git checkout inl.go
cd ../../../../

# Rebuild with original parameters
cd src
./make.bash

# Clean up test binaries
rm -f ../bin/go-default ../bin/go-aggressive ../bin/go-conservative ../bin/go-no-inline
```

## 📚 Key Takeaways

1. **Inlining is a Trade-off**: More inlining = larger binaries but potentially faster execution
2. **Budget System**: Go uses sophisticated cost analysis to make inlining decisions
3. **Parameter Impact**: Small parameter changes can have significant effects on output
4. **Debug Tools**: Go provides excellent tools for understanding compiler decisions
5. **Real-World Relevance**: These parameters affect every Go program you compile!

The Go compiler team has carefully tuned these defaults through extensive benchmarking - but now you understand how to adjust them for your specific needs! ⚡🎯

---

*Continue to [Exercise 5](05-gofmt-ast-transformation.md) or return to the [main workshop](../README.md)*
