# Contributing to RNX Homebrew Tap

## Formula Updates

The formula is automatically updated when new releases are published in the main [joblet repository](https://github.com/ehsaniara/joblet).

### Manual Updates

If you need to manually update the formula:

1. **Update version and URLs** in `Formula/rnx.rb`
2. **Calculate new checksums**:
   ```bash
   curl -sL "URL_TO_ARCHIVE" | shasum -a 256
   ```
3. **Test the formula**:
   ```bash
   brew audit --strict Formula/rnx.rb
   brew install --build-from-source ./Formula/rnx.rb
   ```
4. **Submit pull request**

### Testing Changes

Before submitting changes:

```bash
# Syntax check
brew audit --strict Formula/rnx.rb

# Test installation
brew install --build-from-source ./Formula/rnx.rb --with-admin
brew install --build-from-source ./Formula/rnx.rb --without-admin

# Test functionality
rnx --version
rnx --help
```

### Reporting Issues

- **Formula issues**: Open issue in this repository
- **RNX functionality**: Open issue in [main repository](https://github.com/ehsaniara/joblet/issues)