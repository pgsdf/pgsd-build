# Contributing to PGSD Build System

Thank you for your interest in contributing to PGSD! This document provides guidelines for development and contributions.

## 🛠️ Development Setup

### Prerequisites
- Go 1.21 or later
- FreeBSD or compatible system
- Git
- ZFS support
- Basic FreeBSD tools

### Getting Started

```bash
# Clone the repository
git clone https://github.com/pgsdf/pgsd-build.git
cd pgsd-build

# Install dependencies
make deps

# Build the project
make build

# Run tests
make test
```

## 📝 Code Style

### Go Code Formatting

```bash
# Format all code
make fmt

# Run linter
make lint
```

### Best Practices

1. **Error Handling** - Always provide context and helpful hints
2. **Validation** - Validate early, fail fast with clear messages
3. **Documentation** - Comment complex logic and public APIs
4. **Testing** - Add tests for new functionality
5. **Commits** - Write clear, descriptive commit messages

## 🎯 Error Handling Guidelines

PGSD emphasizes exceptional error handling. Follow these patterns:

### 1. Provide Context

**Bad:**
```go
return fmt.Errorf("failed: %w", err)
```

**Good:**
```go
return fmt.Errorf("failed to create ZFS pool %s: %w", poolName, err)
```

### 2. Add Helpful Hints

**Bad:**
```go
return fmt.Errorf("command not found: %w", err)
```

**Good:**
```go
return fmt.Errorf("ZFS tools not found: %w\nHint: Install with 'pkg install zfs' or load module with 'kldload zfs'", err)
```

### 3. Validate Before Acting

**Bad:**
```go
func BuildImage(cfg Config) error {
    // Start building immediately
    createPool(cfg.ZpoolName)
}
```

**Good:**
```go
func BuildImage(cfg Config) error {
    // Validate first
    if err := validateConfig(&cfg); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }
    // Then build
    createPool(cfg.ZpoolName)
}
```

### 4. Check File Existence

**Bad:**
```go
data, err := os.ReadFile(path)
if err != nil {
    return err
}
```

**Good:**
```go
if _, err := os.Stat(path); err != nil {
    if os.IsNotExist(err) {
        return fmt.Errorf("required file not found: %s\nPlease ensure the file exists", path)
    }
    return fmt.Errorf("cannot access file %s: %w", path, err)
}
```

### 5. Collect Multiple Errors

When validating multiple items, collect all errors instead of failing on the first:

```go
var errors []string
for _, item := range items {
    if err := validate(item); err != nil {
        errors = append(errors, fmt.Sprintf("  - %s: %v", item, err))
        continue
    }
}
if len(errors) > 0 {
    return fmt.Errorf("validation failed:\n%s", strings.Join(errors, "\n"))
}
```

## 🔍 Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/config/
```

### Writing Tests

```go
func TestLoadImageConfig(t *testing.T) {
    // Test valid config
    cfg, err := LoadImageConfig("testdata/valid.lua")
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if cfg.ID != "test-image" {
        t.Errorf("expected ID 'test-image', got: %s", cfg.ID)
    }

    // Test invalid config
    _, err = LoadImageConfig("testdata/invalid.lua")
    if err == nil {
        t.Error("expected error for invalid config")
    }
}
```

## 📦 Project Structure

### Adding New Features

When adding new functionality:

1. **Create internal package** - Place implementation in `internal/`
2. **Add CLI command** - Wire up in `cmd/pgsdbuild/main.go` or `installer/pgsd-inst/main.go`
3. **Add tests** - Create `*_test.go` files
4. **Update documentation** - Update README and relevant docs
5. **Add Makefile target** - If applicable

### Package Organization

```
internal/
├── config/      # Configuration loading and validation
├── image/       # Image build pipeline
├── iso/         # ISO build pipeline
└── common/      # Shared utilities (if needed)

installer/
├── pgsd-inst/         # Main TUI application
└── internal/install/  # Installation pipeline
```

## 🐛 Debugging

### Enable Verbose Output

```bash
# Build with verbose Go output
go build -v -o bin/pgsdbuild ./cmd/pgsdbuild

# Run with trace
go run -race ./cmd/pgsdbuild/...
```

### Common Issues

**Import cycles:**
- Keep `internal/` packages independent
- Use interfaces for cross-package dependencies

**Build failures:**
- Run `make clean` and rebuild
- Check `go.mod` is up to date (`make deps`)

## 📋 Commit Guidelines

### Commit Message Format

```
<type>: <subject>

<body>

<footer>
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code formatting
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance tasks

### Example

```
feat: Add disk size validation in installer

- Validate disk has sufficient space for image
- Show clear error if disk is too small
- Add minimum size requirement to manifest

Closes #123
```

## 🔄 Pull Request Process

1. **Fork the repository**
2. **Create a feature branch** - `git checkout -b feature/my-feature`
3. **Make your changes** - Follow code style guidelines
4. **Add tests** - Ensure tests pass (`make test`)
5. **Update documentation** - Update README if needed
6. **Commit changes** - Use clear commit messages
7. **Push to your fork** - `git push origin feature/my-feature`
8. **Create Pull Request** - Describe your changes clearly

### PR Checklist

- [ ] Code follows project style
- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] Commit messages are clear
- [ ] No breaking changes (or documented)
- [ ] Error messages are helpful

## 🎨 Code Review

Reviewers will check:
- **Functionality** - Does it work as intended?
- **Error Handling** - Are errors handled properly?
- **Testing** - Are there adequate tests?
- **Documentation** - Is it well documented?
- **Style** - Does it follow project conventions?

## 🚀 Release Process

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
4. Push tag: `git push origin v1.0.0`
5. Create GitHub release

## 💡 Feature Requests

Have an idea? Open an issue with:
- Clear description of the feature
- Use cases and benefits
- Proposed implementation (optional)

## 🐞 Bug Reports

Found a bug? Open an issue with:
- Steps to reproduce
- Expected behavior
- Actual behavior
- Error messages (if any)
- System information (OS, Go version)

## 📚 Resources

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [FreeBSD Handbook](https://docs.freebsd.org/en/books/handbook/)
- [ZFS Documentation](https://openzfs.github.io/openzfs-docs/)

## 📞 Getting Help

- Open an issue for bugs or feature requests
- Check existing documentation in `docs/`
- Review code comments for implementation details

## 📄 License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to PGSD! 🎉
