# Homebrew Support Summary

This document summarizes the Homebrew support that has been set up for the `terraform-ops` project.

## What's Been Created

### 1. Homebrew Formula

- **File**: `Formula/terraform-ops.rb`
- **Purpose**: Defines how Homebrew should install `terraform-ops`
- **Features**:
  - Multi-platform support (macOS ARM64/AMD64, Linux AMD64)
  - Automatic platform detection
  - Proper installation to Homebrew's bin directory
  - Test verification that the binary works

### 2. GitHub Actions Release Workflow

- **File**: `.github/workflows/release.yml`
- **Purpose**: Automatically builds and releases binaries when you create a tag
- **Features**:
  - Builds for all supported platforms
  - Creates GitHub releases
  - Uploads binaries as release assets
  - Calculates SHA256 hashes for formula updates

### 3. Update Script

- **File**: `scripts/update-homebrew-formula.sh`
- **Purpose**: Automates updating the formula with new SHA256 hashes
- **Usage**: `./scripts/update-homebrew-formula.sh <version> <linux-sha256> <darwin-amd64-sha256> <darwin-arm64-sha256>`

### 4. Makefile Targets

- **Added**: `homebrew-test`, `homebrew-install`, `homebrew-uninstall`
- **Purpose**: Easy testing and installation of the Homebrew formula

### 5. Documentation

- **File**: `docs/homebrew.md` - User installation guide
- **File**: `docs/homebrew-setup-guide.md` - Complete setup guide for developers
- **Updated**: `README.md` - Added Homebrew installation instructions

## Current Status

✅ **Formula Created**: The Homebrew formula is ready with placeholder SHA256 values
✅ **Build Process**: Multi-platform builds work correctly
✅ **GitHub Actions**: Release workflow is configured
✅ **Testing**: Formula has been tested locally and works correctly
✅ **Documentation**: Complete documentation is available

## Next Steps

### For First Release

1. **Create Homebrew Tap Repository**:

   ```bash
   # Create a new repository named 'homebrew-terraform-ops' on GitHub
   git clone https://github.com/yu/homebrew-terraform-ops.git
   cp Formula/terraform-ops.rb homebrew-terraform-ops/
   cd homebrew-terraform-ops
   git add terraform-ops.rb
   git commit -m "Initial Homebrew formula for terraform-ops"
   git push origin main
   ```

2. **Create First Release**:

   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

3. **Update Formula with Real SHA256 Values**:

   ```bash
   # Get hashes from GitHub Actions workflow output
   ./scripts/update-homebrew-formula.sh 0.1.0 <linux-hash> <darwin-amd64-hash> <darwin-arm64-hash>
   ```

4. **Update Both Repositories**:

   ```bash
   # Update main repository
   git add Formula/terraform-ops.rb
   git commit -m "Update Homebrew formula for v0.1.0"
   git push origin main

   # Update tap repository
   cp Formula/terraform-ops.rb ../homebrew-terraform-ops/
   cd ../homebrew-terraform-ops
   git add terraform-ops.rb
   git commit -m "Update formula for v0.1.0"
   git push origin main
   ```

### For Users

Once the tap is set up, users can install with:

```bash
brew tap yu/terraform-ops
brew install terraform-ops
```

## Verification

The setup has been verified by:

1. ✅ Building binaries for all platforms (`make build-all`)
2. ✅ Testing the formula locally with actual binaries
3. ✅ Verifying the installed binary works correctly
4. ✅ Checking that all documentation is complete

## Maintenance

For future releases:

1. Tag a new version: `git tag v0.2.0 && git push origin v0.2.0`
2. Wait for GitHub Actions to complete
3. Get SHA256 hashes from workflow output
4. Update formula using the script
5. Update both repositories
6. Test installation

## Files Created/Modified

### New Files

- `Formula/terraform-ops.rb` - Homebrew formula
- `.github/workflows/release.yml` - Release workflow
- `scripts/update-homebrew-formula.sh` - Update script
- `docs/homebrew.md` - User guide
- `docs/homebrew-setup-guide.md` - Setup guide

### Modified Files

- `Makefile` - Added Homebrew targets
- `README.md` - Added Homebrew installation instructions

## Support

If you encounter issues:

1. Check the [Homebrew Setup Guide](docs/homebrew-setup-guide.md) for detailed instructions
2. Verify the formula syntax with `brew audit`
3. Test locally with `make homebrew-test`
4. Check GitHub Actions workflow for build issues
