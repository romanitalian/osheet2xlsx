# GitHub Actions

This directory contains GitHub Actions configurations for the osheet2xlsx project.

## Pipelines

### CI Pipeline (`.github/workflows/ci.yml`)

Main pipeline for continuous integration.

**Triggers:**
- Push to `main` and `develop` branches
- Pull requests to `main` and `develop` branches

**Jobs:**
- `test`: Testing on Linux, macOS, Windows
- `lint`: Code linting with golangci-lint
- `build`: Application build on all platforms
- `security`: Security checks
- `examples`: Testing examples

### Release Pipeline (`.github/workflows/release.yml`)

Automatic releases when creating tags.

**Triggers:**
- Push tags with prefix `v*` (e.g., `v1.0.0`)

**Jobs:**
- `release`: Creating releases for all platforms
- `test-release`: Testing release builds

### Security Pipeline (`.github/workflows/security.yml`)

Security checks and code analysis.

**Triggers:**
- Weekly (Monday at 2:00 AM)
- Push to `main` branch
- Pull requests to `main` branch

**Jobs:**
- `security-scan`: Security scanning with gosec
- `dependency-check`: Dependency checking
- `code-quality`: Code quality analysis

### CodeQL Pipeline (`.github/workflows/codeql.yml`)

Code security analysis using CodeQL.

**Triggers:**
- Push to `main` branch
- Pull requests to `main` branch
- Weekly (Sunday at 1:30 AM)

## Dependabot (`.github/dependabot.yml`)

Automatic dependency updates.

**Settings:**
- Go modules: weekly on Mondays
- GitHub Actions: weekly on Mondays
- Automatic pull request creation

## Commands for Local Testing

```bash
# Testing
make test
make test-race
make test-coverage

# Linting
make lint
make lint-fix

# Building
make build
make build-race

# Examples
make run-example
make run-typed
```

## Supported Platforms

- **Linux**: ubuntu-latest (amd64, arm64)
- **macOS**: macos-latest (amd64, arm64)
- **Windows**: windows-latest (amd64, arm64)

## Integrations

- **Codecov**: Code coverage reports
- **GitHub Security**: Security analysis
- **GitHub Releases**: Automatic releases
- **Dependabot**: Dependency management

## Monitoring

All pipelines can be monitored in the "Actions" section on GitHub:

1. Go to the repository on GitHub
2. Click on the "Actions" tab
3. Select the desired workflow to view details

## Troubleshooting

### Linter Issues
```bash
# Local check
make lint

# Auto-fix
make lint-fix

# Check new files
make lint-new
```

### Test Issues
```bash
# Run all tests
make test

# Tests with race detector
make test-race

# Tests with coverage
make test-coverage
```

### Build Issues
```bash
# Clean artifacts
make clean

# Rebuild
make build
```

## Branch Protection Setup

It's recommended to configure branch protection rules for the `main` branch:

1. Go to Settings → Branches
2. Add a rule for the `main` branch
3. Enable:
   - Require status checks to pass before merging
   - Require branches to be up to date before merging
   - Require pull request reviews before merging

## Creating a Release

To create a release:

1. Create a tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. GitHub Actions will automatically:
   - Build the application for all platforms
   - Create a release on GitHub
   - Generate release notes
   - Upload artifacts for download
