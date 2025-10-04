# 🌱 Exercise 0: Introduction and Setup

Welcome to the Go Source Code Workshop! 🎉 In this introductory exercise, you'll set up your environment and get familiar with the Go source code repository.

## 🎯 Learning Objectives

By the end of this exercise, you will:

- ✅ Have a working Go development environment
- ✅ Know how to get the Go source code

## 📋 Prerequisites

Make sure you have the following installed:

- 🗂️ Git
- 💾 At least 4GB of free disk space

## ⚡ Step 1: Install or Upgrade Go

**⚠️ Important**: You need an existing Go installation (version 1.24.6 or newer) to build Go from source. This is called "bootstrapping" - using an existing Go compiler to build the new one.

### Check Your Current Go Version

```bash
go version
# Should show: go version go1.24.6 or higher
```

### If Go is Not Installed or Too Old

If you don't have Go installed, or your version is older than 1.24.6:

1. 📥 **Download Go**: Visit <https://go.dev/dl/> and download the appropriate installer for your operating system
2. 🛠️ **Install Go**: Follow the official installation guide for your platform:
3. ✅ **Verify Installation**: Open a new terminal and run:

   ```bash
   go version
   # Should show: go version go1.24.6 or higher
   ```

**📚 Installation Help**: If you need detailed installation instructions, see the [official Go installation guide](https://go.dev/doc/install).

## 📥 Step 2: Clone the Go Source Code

Let's clone the official Go repository. This might take a few minutes as it's a large repository. ⏳

```bash
git clone https://go.googlesource.com/go
cd go
```

## 🏷️ Step 3: Checkout Go Version 1.25.1

For consistency across the workshop, we'll use Go version 1.25.1. Let's check out the specific release tag: 📌

```bash
git checkout go1.25.1
```

Verify you're on the correct version:

```bash
git describe --tags
# Should show: go1.25.1
```

## 🎓 What We Accomplished

- ✅ Installed or verified Go 1.24.6+ for bootstrapping
- ✅ Cloned the official Go repository
- ✅ Checked out Go version 1.25.1 for consistency
- ✅ Environment is ready for building Go from source

## ➡️ Next Steps

Perfect! Your environment is now set up and ready. In [Exercise 1: Compiling Go Without Changes](./01-compile-go-unchanged.md), you'll build the Go toolchain from source and explore the Go compiler structure while it's building! 🔨
