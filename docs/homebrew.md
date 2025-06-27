# Installing terraform-ops via Homebrew

This document explains how to install `terraform-ops` using Homebrew and how to set up a custom Homebrew tap for distribution.

## For Users

### Option 1: Install from Custom Tap (Recommended)

1. Add the custom tap:

   ```bash
   brew tap yu/terraform-ops
   ```

2. Install terraform-ops:
   ```bash
   brew install terraform-ops
   ```

### Option 2: Install from Local Formula

If you have the formula locally:

```bash
brew install --formula Formula/terraform-ops.rb
```

### Option 3: Install from URL

```bash
brew install https://raw.githubusercontent.com/yu/terraform-ops/main/Formula/terraform-ops.rb
```

## For Developers

### Setting up Homebrew Tap

1. Create a new repository named `homebrew-terraform-ops` (the `homebrew-` prefix is required)

2. Copy the formula to the tap repository:

   ```bash
   mkdir -p homebrew-terraform-ops
   cp Formula/terraform-ops.rb homebrew-terraform-ops/
   ```

3. Update the formula with correct SHA256 hashes after each release

### Updating the Formula

After creating a new release:

1. Download the binaries from the GitHub release
2. Calculate SHA256 hashes:

   ```bash
   sha256sum terraform-ops-darwin-amd64
   sha256sum terraform-ops-darwin-arm64
   sha256sum terraform-ops-linux-amd64
   ```

3. Update the formula with the new hashes and version number

### Release Process

1. Tag a new release:

   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

2. The GitHub Actions workflow will automatically:
   - Build binaries for all platforms
   - Create a GitHub release
   - Upload the binaries
   - Calculate SHA256 hashes

3. Update the Homebrew formula with the new hashes and push to the tap repository

### Testing the Formula

Test the formula locally before publishing:

```bash
brew install --formula Formula/terraform-ops.rb
brew test terraform-ops
```

## Formula Details

The Homebrew formula (`Formula/terraform-ops.rb`) includes:

- Multi-platform support (macOS ARM64/AMD64, Linux AMD64)
- Automatic platform detection
- Proper installation to `/usr/local/bin` (or `/opt/homebrew/bin` on Apple Silicon)
- Test verification that the binary works

## Troubleshooting

### Common Issues

1. **SHA256 mismatch**: Ensure you're using the correct hashes from the latest release
2. **Platform detection**: The formula automatically detects the platform and architecture
3. **Permission issues**: The workflow sets executable permissions on all binaries

### Updating an Existing Installation

```bash
brew update
brew upgrade terraform-ops
```

### Uninstalling

```bash
brew uninstall terraform-ops
```
