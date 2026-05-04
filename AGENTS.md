# AGENTS.md


## Essential Commands

```bash
# Test
gotestsum --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...

# Test (verbose)
gotestsum --format testdox --format-hide-empty-pkg -- ./...

# Watch mode
gotestsum --watch --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...

# Lint
golangci-lint run ./...

# Format
golangci-lint fmt


