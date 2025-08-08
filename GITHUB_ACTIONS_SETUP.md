# GitHub Actions Setup for osheet2xlsx

## What's Created

✅ **All necessary GitHub Actions files have been created:**

1. **`.github/workflows/ci.yml`** - Main CI pipeline
2. **`.github/workflows/release.yml`** - Automatic releases
3. **`.github/workflows/security.yml`** - Security checks
4. **`.github/workflows/codeql.yml`** - Code security analysis
5. **`.github/dependabot.yml`** - Automatic dependency updates
6. **`.github/README.md`** - Actions documentation

## How to Activate

### 1. Commit and push changes:

```bash
git add .github/
git commit -m "feat: add GitHub Actions CI/CD pipelines"
git push origin main
```

### 2. Check Actions work:

1. Go to the repository on GitHub
2. Click on the "Actions" tab
3. Make sure pipelines started and passed successfully

### 3. Set up branch protection (recommended):

1. Go to Settings → Branches
2. Add a rule for the `main` branch
3. Enable:
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging
   - ✅ Require pull request reviews before merging

## What Pipelines Do

### 🔄 CI Pipeline
- **Triggers**: push to main/develop, pull requests
- **Testing**: on Linux, macOS, Windows
- **Linting**: golangci-lint
- **Building**: artifacts for all platforms
- **Security**: basic checks
- **Examples**: testing examples

### 🚀 Release Pipeline
- **Triggers**: push tags `v*`
- **Building**: for all platforms (Linux, macOS, Windows)
- **Archives**: .tar.gz for Linux/macOS, .zip for Windows
- **Releases**: automatic creation on GitHub

### 🛡️ Security Pipeline
- **Triggers**: weekly, push to main, pull requests
- **gosec**: security scanning
- **govulncheck**: vulnerability checking
- **nancy**: dependency analysis

### 🔍 CodeQL Pipeline
- **Triggers**: push to main, pull requests, weekly
- **Analysis**: code security
- **Integration**: with GitHub Security

## Testing Commands

```bash
# Local testing
make test          # Tests
make lint          # Linter
make build         # Build
make run-example   # Examples

# GitHub Actions will use the same commands
```

## Creating First Release

```bash
# Create a tag
git tag v1.0.0

# Push the tag
git push origin v1.0.0

# GitHub Actions will automatically create a release
```

## Monitoring

- **Actions**: https://github.com/romanitalian/osheet2xlsx/actions
- **Security**: https://github.com/romanitalian/osheet2xlsx/security
- **Releases**: https://github.com/romanitalian/osheet2xlsx/releases

## Troubleshooting

### If Actions don't start:
1. Check that files in `.github/workflows/` are committed
2. Make sure YAML syntax is correct
3. Check logs in the Actions section

### If tests fail:
1. Run locally: `make test`
2. Check linter: `make lint`
3. Fix errors and commit

### If build fails:
1. Check dependencies: `go mod tidy`
2. Clean cache: `make clean`
3. Rebuild: `make build`

## Additional Settings

### Codecov (optional)
1. Connect Codecov to the repository
2. Add token to GitHub Secrets
3. Configure coverage reports

### Slack notifications (optional)
1. Add Slack webhook to Secrets
2. Configure notifications in workflows

### Automatic deployment (optional)
1. Add deployment workflow
2. Configure deployment environments
3. Add approval gates

## Status

✅ **All pipelines created and ready to use**

Next step: commit and push changes to the repository!
