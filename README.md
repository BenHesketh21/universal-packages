# Universal Packages

**Universal Packages** is a CLI tool for pushing and pulling private code packages using [OCI (Open Container Initiative)](https://opencontainers.org/) registries. It provides a ecosystem-agnostic, decentralized, and standardized way to distribute private SDKs and packages across teams and ecosystems.

> 🔧 Currently supports **npm** packages. Future support for pip, NuGet, and Go is planned.

---

## ✨ Goals

- ✅ Support for installing and publishing private packages across ecosystems.
- ✅ Use standard OCI registries (e.g. GitHub, AWS, Azure).
- ✅ Minimal ecosystem-specific logic; rely on established package formats.
- ✅ Reproducible and interoperable workflows.

---

## 🔧 Developer Setup

```bash
# Clone the repo
git clone https://github.com/BenHesketh21/universal-packages.git
cd universal-packages

# Install Go dependencies
task tidy

# Build
task build

# Test
task test

# Lint (requires golangci-lint)
task lint
```

## Usage

```bash
# Pull a package from OCI and update your project (npm)
upgk install ghcr.io/org/mypackage:version
npm install

# Package (via npm pack) and push to OCI registry
npm pack
upgk push ghcr.io/org/mypackage:version
```
