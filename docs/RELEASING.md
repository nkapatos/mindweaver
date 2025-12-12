# Releasing Mindweaver

Technical documentation for the release process, tooling, and troubleshooting.

## Overview

Mindweaver uses automated releases via GitHub Actions and goreleaser. Every merge to `main` that affects the codebase triggers a version bump, build, and release.

## Release Tools

### goreleaser

**Purpose:** Builds cross-platform binaries and creates GitHub releases

**Configuration:** `.goreleaser.yml`

**Features:**
- Cross-compilation for multiple OS/arch combinations
- Automatic archive creation with checksums
- GitHub Release creation with notes
- Version embedding via ldflags
- Changelog generation from commits

**Test locally:**
```bash
# Dry run (no release)
goreleaser release --snapshot --clean

# Check what would be released
goreleaser check

# Build only (no release)
goreleaser build --snapshot --clean
```

### GitHub Actions

**Workflows:**
- `.github/workflows/ci.yml` - PR checks
- `.github/workflows/pr-title-check.yml` - PR title validation
- `.github/workflows/release.yml` - Automated releases

**Release workflow logic:**
1. Detects changes to relevant paths
2. Reads current version from `cmd/mindweaver/VERSION`
3. Analyzes commit message type
4. Calculates version bump
5. Updates VERSION file
6. Creates git tag
7. Runs goreleaser
8. Updates CHANGELOG
9. Commits changes back

## Versioning

### Version File

**Location:** `cmd/mindweaver/VERSION`

**Format:** Plain text, single line
```
0.9.0
```

**Used by:**
- GitHub Actions (reading current version)
- goreleaser (via git tags)
- Documentation references

### Git Tags

**Format:** `mindweaver/vX.Y.Z`

**Examples:**
```
mindweaver/v0.9.0
mindweaver/v0.10.0
mindweaver/v1.0.0
```

**Create manually:**
```bash
git tag -a mindweaver/v1.0.0 -m "Release v1.0.0"
git push origin mindweaver/v1.0.0
```

**List tags:**
```bash
git tag -l "mindweaver/*"
```

**Delete tag (if needed):**
```bash
git tag -d mindweaver/v1.0.0
git push origin :refs/tags/mindweaver/v1.0.0
```

### Version Bump Logic

**Implemented in:** `.github/workflows/release.yml`

```bash
# Breaking change (in 0.x.x)
feat(mind)!: breaking change
# OR
feat(mind): new feature

BREAKING CHANGE: details
# Result: 0.9.0 → 0.10.0 (minor bump)

# Feature
feat(mind): add search
# Result: 0.9.0 → 0.10.0 (minor bump)

# Fix
fix(mind): resolve bug
# Result: 0.9.0 → 0.9.1 (patch bump)

# No bump
chore: update docs
docs: fix typo
# Result: No release triggered
```

**After 1.0.0:**
- `feat!:` → MAJOR bump (1.0.0 → 2.0.0)
- `feat:` → MINOR bump (1.0.0 → 1.1.0)
- `fix:` → PATCH bump (1.0.0 → 1.0.1)

## Release Process

### Automatic Release (Normal Flow)

**1. Merge PR to main:**
```bash
# Via GitHub UI or CLI
gh pr merge 123 --squash
```

**2. GitHub Action triggers:**
```
✓ Detect changes to cmd/mindweaver/ or internal/
✓ Read current version: 0.9.0
✓ Analyze commit: "feat(mind): add search"
✓ Calculate bump: minor (0.9.0 → 0.10.0)
✓ Update VERSION file → 0.10.0
✓ Create tag → mindweaver/v0.10.0
✓ Run goreleaser → build binaries
✓ Create GitHub Release
✓ Update CHANGELOG.md
✓ Push changes to main
```

**3. Release published:**
- GitHub Releases page shows new release
- Binaries available for download
- CHANGELOG updated on main branch

### Manual Release

**Scenario 1: Force specific version**

```bash
# Create empty commit with version override
git commit --allow-empty -m "chore(mindweaver): release v1.0.0

RELEASE-AS: 1.0.0"

git push origin main
```

**Scenario 2: Release from local machine**

```bash
# Ensure you're on main and up to date
git checkout main
git pull

# Update VERSION file
echo "1.0.0" > cmd/mindweaver/VERSION

# Commit and tag
git add cmd/mindweaver/VERSION
git commit -m "chore(mindweaver): bump version to 1.0.0"
git tag -a mindweaver/v1.0.0 -m "Release v1.0.0"
git push origin main mindweaver/v1.0.0

# Run goreleaser locally
export GITHUB_TOKEN="your_token_here"
goreleaser release --clean
```

**Scenario 3: Hotfix release**

```bash
# Branch from tag
git checkout -b hotfix-0.9.1 mindweaver/v0.9.0

# Fix bug
git commit -m "fix(mind): critical bug fix"

# Update version
echo "0.9.1" > cmd/mindweaver/VERSION
git add cmd/mindweaver/VERSION
git commit -m "chore(mindweaver): bump version to 0.9.1"

# Tag and push
git tag -a mindweaver/v0.9.1 -m "Hotfix v0.9.1"
git push origin mindweaver/v0.9.1

# Release
goreleaser release --clean

# Merge back to main
git checkout main
git merge hotfix-0.9.1
git push origin main
```

## Build Configuration

### Platforms

**Currently built:**
- `darwin/amd64` - macOS Intel
- `darwin/arm64` - macOS Apple Silicon
- `linux/amd64` - Linux x86-64
- `linux/arm64` - Linux ARM64

**Adding new platforms:**

Edit `.goreleaser.yml`:
```yaml
builds:
  - goos:
      - darwin
      - linux
      - windows  # Add Windows
    goarch:
      - amd64
      - arm64
      - 386  # Add 32-bit
```

### Build Flags

**Embedded version info:**
```go
// Set via ldflags in .goreleaser.yml
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
    builtBy = "unknown"
)
```

**Access in code:**
```go
fmt.Printf("Mindweaver %s (commit: %s, built: %s)\n", version, commit, date)
```

**Manual build with version:**
```bash
go build -ldflags="-X main.version=0.9.0 -X main.commit=$(git rev-parse HEAD)" ./cmd/mindweaver
```

### CGO

**Note:** `CGO_ENABLED=1` in goreleaser config for SQLite support.

**Cross-compilation challenges:**
- macOS → Linux: Requires cross-compilation toolchain
- Currently built on appropriate platforms via GitHub Actions

## CHANGELOG Management

### Format

**Based on:** [Keep a Changelog](https://keepachangelog.com/)

**Structure:**
```markdown
## [Unreleased]

## [X.Y.Z] - YYYY-MM-DD

### Mind Service
#### Features
- Added thing (#PR)

#### Bug Fixes
- Fixed thing (#PR)

### Brain Service
#### Features
- Added thing (#PR)

### Combined
#### Chores
- Updated deps (#PR)
```

### Automatic Updates

The release workflow automatically:
1. Extracts commits since last tag
2. Creates new version section
3. Lists commits as bullet points
4. Commits back to main

### Manual CHANGELOG Edits

**To improve auto-generated changelog:**

```bash
# After release, edit CHANGELOG.md
vim CHANGELOG.md

# Group commits better, add details
# Commit changes
git add CHANGELOG.md
git commit -m "docs(mindweaver): improve CHANGELOG for v0.10.0"
git push origin main
```

**This won't affect the release** (already published) but improves documentation.

## GitHub Releases

### Release Configuration

**In `.goreleaser.yml`:**
- `prerelease: auto` - Marks 0.x.x as pre-release
- Custom header and footer templates
- Release name template
- Changelog grouping by commit type

### Pre-release vs. Stable

**Pre-release (0.x.x):**
- Marked with "Pre-release" badge on GitHub
- Shows warning: "This is a pre-release version"
- Useful for early testing

**Stable (1.x.x+):**
- Not marked as pre-release
- Listed as "Latest" release
- Production-ready signal

### Release Notes

**Auto-generated from commits:**
- Header from template
- Commits grouped by type (Mind/Brain/Features/Fixes)
- Footer with installation instructions
- Link to full changelog

**Manual editing:**

After release is created, edit on GitHub:
1. Go to Releases page
2. Click "Edit" on the release
3. Modify description
4. Save

## Troubleshooting

### Release Didn't Trigger

**Check:**
1. Was PR merged to `main`?
2. Did changes affect relevant paths?
3. Does commit message start with `feat:` or `fix:`?
4. Check Action logs in GitHub

**Fix:**
```bash
# Create empty commit to retrigger
git commit --allow-empty -m "feat(mindweaver): trigger release"
git push origin main
```

### Wrong Version Number

**If version bumped incorrectly:**

```bash
# Delete tag
git tag -d mindweaver/vX.Y.Z
git push origin :refs/tags/mindweaver/vX.Y.Z

# Delete release on GitHub
gh release delete mindweaver/vX.Y.Z

# Fix VERSION file
echo "correct.version" > cmd/mindweaver/VERSION
git add cmd/mindweaver/VERSION
git commit -m "chore(mindweaver): fix version to correct.version"

# Create correct tag
git tag -a mindweaver/vX.Y.Z -m "Release vX.Y.Z"
git push origin main mindweaver/vX.Y.Z

# Re-run goreleaser (automatically or manually)
```

### Build Failed

**Check goreleaser logs in GitHub Actions:**
```bash
# Test locally
goreleaser release --snapshot --clean

# Common issues:
# - CGO cross-compilation
# - Missing dependencies
# - Build errors in code
```

### Tag Already Exists

**If tag exists but release failed:**

```bash
# Delete tag
git tag -d mindweaver/v0.9.0
git push origin :refs/tags/mindweaver/v0.9.0

# Fix issue, retag
git tag -a mindweaver/v0.9.0 -m "Release v0.9.0"
git push origin mindweaver/v0.9.0
```

### Release Artifacts Missing

**If some binaries didn't build:**

Check `.goreleaser.yml` for platform configuration:
```yaml
builds:
  - ignore:
      - goos: windows
        goarch: arm64  # Not supported
```

**Rebuild specific platform:**
```bash
GOOS=darwin GOARCH=arm64 go build ./cmd/mindweaver
```

## Advanced Topics

### Multiple Binaries

**Future: When adding `imex`, `lsp`, etc.**

`.goreleaser.yml`:
```yaml
builds:
  - id: mindweaver
    main: ./cmd/mindweaver
    binary: mindweaver
  
  - id: imex
    main: ./cmd/imex
    binary: mw-imex
  
  - id: lsp
    main: ./cmd/lsp
    binary: mw-lsp
```

Each binary built and included in release.

### Separate Release Workflows

**When components need independent versioning:**

Create `.github/workflows/release-imex.yml`:
```yaml
on:
  push:
    paths:
      - 'cmd/imex/**'
```

Uses different VERSION file and tag prefix: `imex/vX.Y.Z`

### Docker Images

**Future: Publishing to container registry**

Add to `.goreleaser.yml`:
```yaml
dockers:
  - image_templates:
      - "ghcr.io/nkapatos/mindweaver:{{ .Tag }}"
      - "ghcr.io/nkapatos/mindweaver:latest"
    dockerfile: docker/Dockerfile
```

## Testing Releases

### Snapshot Release (Local)

**Build without releasing:**
```bash
goreleaser release --snapshot --clean --skip=publish

# Check dist/ directory
ls -la dist/
```

**Test binary:**
```bash
./dist/mindweaver_darwin_arm64/mindweaver --version
```

### Test Release Workflow

**Using act (local GitHub Actions):**
```bash
# Install act
brew install act

# Run release workflow locally
act push -W .github/workflows/release.yml
```

**Note:** Some features (GitHub API calls) won't work locally.

## Maintenance

### Periodic Tasks

**Weekly:**
- Review failed release attempts
- Check for goreleaser updates

**Monthly:**
- Review CHANGELOG for clarity
- Audit old releases (consider archiving very old ones)

**Quarterly:**
- Update goreleaser to latest version
- Review GitHub Actions for deprecations
- Test full release flow manually

### Updating goreleaser

```bash
# Update .goreleaser.yml to latest schema version
# Check: https://goreleaser.com/deprecations/

# Test locally
goreleaser check
goreleaser release --snapshot --clean
```

## Resources

- [goreleaser Documentation](https://goreleaser.com/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)

---

**Last Updated:** 2025-12-12  
**Version:** 0.9.0
