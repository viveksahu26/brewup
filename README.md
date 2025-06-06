# brewup

brewup is a Go-based CLI tool that automates the process of updating Homebrew formula files for Interlynk projects (e.g., sbomasm, sbomqs, sbommv, sbomgr, sbomex). It updates the version number and SHA256 checksums in the formula file by downloading the latest release binaries from GitHub and calculating their checksums. This eliminates the manual, error-prone task of updating formulas for each release.

## Features

- Updates the version field in a Homebrew formula file.
- Updates url fields for macOS (arm64/amd64) and Linux (arm64/amd64) binaries to point to a specified release version.
- Automatically calculates SHA256 checksums by downloading binaries from GitHub.
- Supports a dry-run mode to preview changes without modifying the file.

## Prerequisites

- Go: Version 1.21 or later (for building the tool).
- GitHub Access: The tool downloads binaries from `https://github.com/interlynk-io/<repo>/releases`. Ensure the specified release exists.
- Homebrew Formula File: A formula file (e.g., sbomasm.rb) with a structure similar to the one used by Interlynk projects.

## Installation

### 1. Clone the Repository:

```bash
git clone https://github.com/viveksahu26/brewup.git
cd brewup
```

### 2. Initialize Go Module (if not already done):

```bash
go mod init github.com/viveksahu26/brewup
```

### 3. Install Dependencies:

```bash
go get github.com/spf13/cobra@v1.8.0
```

### 4. Build the Tool:

```bash
go build -o brewup main.go
```

### 5. Move to PATH (optional, for global access):

```bash
mv brewup /usr/local/bin/
```


## Usage

Run brewup with the required flags to update a Homebrew formula file:

```bash
./brewup --repo <repository> --version <version> --file <formula_file>
```

### Flags

- `--repo, -r`: The repository name (e.g., sbomasm). Required.
- `--version, -v`: The release version (e.g., v1.0.5). Must start with v. Required.
- `--file, -f`: Path to the Homebrew formula file (e.g., sbomasm.rb). Required.
- `--dry-run`: Preview changes without modifying the file (optional).

## Examples

### 1. Update sbomasm.rb for release v1.0.4:

```bash
./brewup --repo sbomasm --version v1.0.4 --file sbomasm.rb
```

```bash
Output:
Changes to sbomasm.rb:
Version: version "v1.0.3" -> version "v1.0.5"
Checksum (darwin-arm64): df798139... -> <new_checksum>
Checksum (darwin-amd64): 240ceccc... -> <new_checksum>
Checksum (linux-arm64): d3af03dd... -> <new_checksum>
Checksum (linux-amd64): cc7dd985... -> <new_checksum>
Successfully updated sbomasm.rb
```

### 2. Preview changes with dry-run:

```bash
./brewup --repo sbomasm --version v1.0.4 --file sbomasm.rb --dry-run
```

```bash
Output:
Changes to sbomasm.rb:
Version: version "v1.0.3" -> version "v1.0.4"
Checksum (darwin-arm64): df798139... -> <new_checksum>
Checksum (darwin-amd64): 240ceccc... -> <new_checksum>
Checksum (linux-arm64): d3af03dd... -> <new_checksum>
Checksum (linux-amd64): cc7dd985... -> <new_checksum>
Dry-run mode: No changes written to file
Updated content preview:
<updated_formula_content>
```
