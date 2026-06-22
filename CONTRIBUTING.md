# Contributing to haunt-space

New to Go? No problem. This guide walks you through everything — setting up Go, building the project, testing it inside Ghostty, installing it for daily use, and shipping a new release via Homebrew.

---

## 1. Install Go

Go is a compiled language with a single toolchain binary. You install it once and it handles everything.

**macOS (recommended):**
```bash
brew install go
```

**Or download directly:** https://go.dev/dl/ — pick the macOS `.pkg` installer and run it.

Verify it worked:
```bash
go version
# go version go1.24.x darwin/arm64
```

---

## 2. Clone the repo

```bash
git clone https://github.com/samtran0331/haunt-space.git
cd haunt-space
```

---

## 3. Understand the project layout

Unlike JavaScript (no `node_modules`) or Python (no virtualenv needed), Go downloads dependencies into a global cache on your machine. The two files that describe dependencies are:

- **`go.mod`** — declares the module name and direct dependencies (like `package.json`)
- **`go.sum`** — locks exact checksums for every dependency (like `package-lock.json`)

You never edit these by hand. The `go` toolchain manages them.

All source files in this project are in one directory (`package main`). There are no subdirectories or packages to navigate — everything at the root is part of the same program.

---

## 4. Download dependencies

```bash
go mod download
```

This pulls all dependencies into your local Go cache (`~/go/pkg/mod`). You only need to run this once, or after `go.mod` changes.

---

## 5. Build the binary

```bash
go build -o hsp .
```

- `go build` compiles everything
- `-o hsp` names the output binary `hsp`
- `.` means "build the package in the current directory"

This produces a single self-contained binary called `hsp` in your current directory. No runtime, no interpreter, no dependencies needed to run it — just the binary.

---

## 6. Run tests

```bash
go test ./...
```

- `./...` means "run tests in this directory and all subdirectories"
- Tests live in files ending in `_test.go`

To run a specific test by name:
```bash
go test -run TestBuildGhosttyCommand .
```

---

## 7. Test inside Ghostty (dev workflow)

Since `hsp` launches Ghostty windows, you need to test it from inside a Ghostty terminal.

**Step 1:** Open Ghostty and `cd` into the repo directory:
```bash
cd ~/path/to/haunt-space
```

**Step 2:** Build the binary:
```bash
go build -o hsp .
```

This creates a file called `hsp` in the current directory.

**Step 3:** Run it using `./hsp`, not `hsp`.

> **Why `./hsp` and not just `hsp`?**
> On macOS, your shell only looks for programs in directories listed in your `PATH` environment variable — things like `/usr/bin` and `/usr/local/bin`. The current directory (`.`) is intentionally not in `PATH` for security reasons. The `./` prefix tells the shell "run the binary right here in this folder."

```bash
./hsp boo
```

The `boo` subcommand opens the interactive TUI wizard. Walk through it to create a layout template. Templates are saved to `~/.config/haunt-space/templates/`.

**Step 4:** Summon (launch) a saved template:
```bash
./hsp summon <your-template-name>
```

This spawns a new Ghostty window using the layout you designed. The process detaches immediately — your current terminal is not blocked.

**Quick iteration loop:**
```bash
go build -o hsp . && ./hsp summon my-template
```

---

## 8. Install for production use (personal machine)

When you're happy with the binary and want `hsp` available everywhere without `./`:

```bash
go install .
```

This builds and copies the binary to `~/go/bin/hsp`. Make sure `~/go/bin` is in your `PATH`:

```bash
# Add to ~/.zshrc or ~/.bash_profile if not already there:
export PATH="$HOME/go/bin:$PATH"
```

Then reload your shell:
```bash
source ~/.zshrc
```

Now you can run `hsp` from any directory.

To uninstall:
```bash
rm ~/go/bin/hsp
```

---

## 9. Publish a release to Homebrew

Homebrew lets users install `hsp` with `brew install`. This requires two things: a GitHub release with binaries, and a Homebrew formula (a Ruby file that describes where to find those binaries).

### Step 1: Tag a release on GitHub

Version tags follow [semver](https://semver.org/): `v1.0.0`, `v1.2.3`, etc.

```bash
git tag v1.0.0
git push origin v1.0.0
```

### Step 2: Build release binaries

Homebrew needs pre-compiled binaries for each platform it supports. Build them locally:

```bash
# macOS Apple Silicon (arm64)
GOOS=darwin GOARCH=arm64 go build -o hsp-darwin-arm64 .

# macOS Intel (amd64)
GOOS=darwin GOARCH=amd64 go build -o hsp-darwin-amd64 .

# Linux (optional, for wider support)
GOOS=linux GOARCH=amd64 go build -o hsp-linux-amd64 .
```

`GOOS` and `GOARCH` are environment variables that tell Go to cross-compile for a different OS/architecture. The output binary won't run on your machine if you cross-compiled for a different arch — but Homebrew users on that platform will run it fine.

### Step 3: Create tar archives and get SHA256 hashes

Homebrew expects `.tar.gz` archives, and its formula needs the exact SHA256 hash of each archive to verify downloads.

```bash
tar -czf hsp-darwin-arm64.tar.gz hsp-darwin-arm64
tar -czf hsp-darwin-amd64.tar.gz hsp-darwin-amd64

shasum -a 256 hsp-darwin-arm64.tar.gz
shasum -a 256 hsp-darwin-amd64.tar.gz
```

Save those SHA256 hashes — you'll need them in the formula.

### Step 4: Upload binaries to the GitHub release

Go to your repo on GitHub → **Releases** → **Draft a new release** → select your tag → upload the `.tar.gz` files as release assets.

Or via CLI:
```bash
gh release create v1.0.0 \
  hsp-darwin-arm64.tar.gz \
  hsp-darwin-amd64.tar.gz \
  --title "v1.0.0" \
  --notes "Initial release"
```

### Step 5: Create a Homebrew tap

A "tap" is just a GitHub repo that Homebrew can read formulas from. Name it `homebrew-haunt-space` (the `homebrew-` prefix is required by Homebrew convention).

Create a new GitHub repo: `samtran0331/homebrew-haunt-space`

### Step 6: Write the formula

Create `Formula/hsp.rb` in your tap repo:

```ruby
class Hsp < Formula
  desc "Declarative terminal window layout manager for Ghostty"
  homepage "https://github.com/samtran0331/haunt-space"
  version "1.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/samtran0331/haunt-space/releases/download/v1.0.0/hsp-darwin-arm64.tar.gz"
      sha256 "PASTE_ARM64_SHA256_HERE"
    else
      url "https://github.com/samtran0331/haunt-space/releases/download/v1.0.0/hsp-darwin-amd64.tar.gz"
      sha256 "PASTE_AMD64_SHA256_HERE"
    end
  end

  def install
    bin.install "hsp-darwin-arm64" => "hsp" if Hardware::CPU.arm?
    bin.install "hsp-darwin-amd64" => "hsp" if Hardware::CPU.intel?
  end

  test do
    assert_match "haunt-space", shell_output("#{bin}/hsp 2>&1", 1)
  end
end
```

Commit and push this file to your `homebrew-haunt-space` repo.

### Step 7: Users can now install

```bash
brew tap samtran0331/haunt-space
brew install hsp
```

### Releasing a new version

1. Make your code changes
2. Commit and push
3. Tag: `git tag v1.x.x && git push origin v1.x.x`
4. Build new binaries, archive them, get new SHA256s
5. Upload to a new GitHub release
6. Update `version`, `url`, and `sha256` in `Formula/hsp.rb` in your tap repo
7. Commit and push the tap — users running `brew upgrade hsp` get the new version automatically

---

## Common Go commands reference

| Command | What it does |
|---|---|
| `go build -o hsp .` | Compile to binary named `hsp` |
| `go test ./...` | Run all tests |
| `go test -run TestName .` | Run one test by name |
| `go vet ./...` | Static analysis (catch bugs before running) |
| `go install .` | Build and install to `~/go/bin` |
| `go mod download` | Download all dependencies |
| `go mod tidy` | Remove unused dependencies from `go.mod` |
