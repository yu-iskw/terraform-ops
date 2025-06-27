# Homebrew Setup Guide for terraform-ops

This guide walks you through the complete process of setting up Homebrew distribution for the `terraform-ops` tool.

## Overview

To distribute `terraform-ops` via Homebrew, you need to:

1. Create a Homebrew tap repository
2. Set up automated releases with GitHub Actions
3. Create and maintain the Homebrew formula
4. Test and publish the formula

## Step 1: Create Homebrew Tap Repository

1. Create a new GitHub repository named `homebrew-terraform-ops`
   - The `homebrew-` prefix is required for Homebrew to recognize it as a tap
   - Make it public so users can access it

2. Clone the repository locally:

   ```bash
   git clone https://github.com/yu/homebrew-terraform-ops.git
   cd homebrew-terraform-ops
   ```

3. Copy the formula from your main repository:

   ```bash
   cp ../terraform-ops/Formula/terraform-ops.rb .
   ```

4. Commit and push:
   ```bash
   git add terraform-ops.rb
   git commit -m "Initial Homebrew formula for terraform-ops"
   git push origin main
   ```

## Step 2: Set Up Automated Releases

The GitHub Actions workflow in `.github/workflows/release.yml` will automatically:

1. Build binaries for all platforms when you create a tag
2. Create a GitHub release
3. Upload the binaries as release assets
4. Calculate SHA256 hashes

### Creating a Release

1. Tag a new version:

   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

2. The workflow will automatically:
   - Build `terraform-ops-linux-amd64`
   - Build `terraform-ops-darwin-amd64`
   - Build `terraform-ops-darwin-arm64`
   - Build `terraform-ops-windows-amd64.exe`
   - Create a release with these assets

3. Check the release page for the SHA256 hashes in the workflow output

## Step 3: Update the Formula

After each release, update the formula with the new hashes:

1. Use the automated script:

   ```bash
   ./scripts/update-homebrew-formula.sh 0.1.0 <linux-sha256> <darwin-amd64-sha256> <darwin-arm64-sha256>
   ```

2. Or manually update `Formula/terraform-ops.rb`:
   - Update the version number
   - Replace the placeholder SHA256 hashes with actual values

3. Test the formula locally:

   ```bash
   make homebrew-test
   ```

4. Commit and push the updated formula:
   ```bash
   git add Formula/terraform-ops.rb
   git commit -m "Update Homebrew formula for v0.1.0"
   git push origin main
   ```

## Step 4: Update the Tap Repository

1. Copy the updated formula to your tap repository:

   ```bash
   cp Formula/terraform-ops.rb ../homebrew-terraform-ops/
   cd ../homebrew-terraform-ops
   ```

2. Commit and push:
   ```bash
   git add terraform-ops.rb
   git commit -m "Update formula for v0.1.0"
   git push origin main
   ```

## Step 5: Test the Installation

Test that users can install your tool:

```bash
# Add the tap
brew tap yu/terraform-ops

# Install the tool
brew install terraform-ops

# Test that it works
terraform-ops --help
```

## Maintenance

### For Each New Release

1. Create a new tag: `git tag v0.2.0 && git push origin v0.2.0`
2. Wait for the GitHub Actions workflow to complete
3. Get the SHA256 hashes from the workflow output
4. Update the formula using the script
5. Update both the main repository and tap repository
6. Test the installation

### Formula Validation

Homebrew provides tools to validate your formula:

```bash
# Check formula syntax
brew audit --formula Formula/terraform-ops.rb

# Test installation
brew install --formula Formula/terraform-ops.rb

# Test the installed binary
brew test terraform-ops
```

## Troubleshooting

### Common Issues

1. **SHA256 mismatch**: Ensure you're using the correct hashes from the latest release
2. **Formula syntax errors**: Use `brew audit` to check for issues
3. **Installation failures**: Check that the binary URLs are accessible
4. **Platform detection**: The formula automatically handles different architectures

### Debugging

```bash
# Check formula details
brew info terraform-ops

# See what would be installed
brew install --formula Formula/terraform-ops.rb --dry-run

# Check formula dependencies
brew deps --formula Formula/terraform-ops.rb
```

## Best Practices

1. **Version Management**: Always update the version number in the formula
2. **Testing**: Test the formula locally before publishing
3. **Documentation**: Keep installation instructions up to date
4. **Automation**: Use scripts to reduce manual errors
5. **Validation**: Use Homebrew's built-in validation tools

## Alternative Distribution Methods

If you prefer not to maintain a custom tap, you can also:

1. **Submit to Homebrew Core**: For widely-used tools (requires approval)
2. **Use a Cask**: For GUI applications (not applicable for CLI tools)
3. **Direct Download**: Provide installation scripts that download from releases

## References

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Tap Documentation](https://docs.brew.sh/Taps)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Cross-Compilation](https://golang.org/doc/install/source#environment)
