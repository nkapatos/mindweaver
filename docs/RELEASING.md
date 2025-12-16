# Releasing Mindweaver

Technical documentation for the release process, tooling, and troubleshooting.

## Overview

The monorepo uses [release-please](https://github.com/googleapis/release-please) for automated, independent releases of each component. Each component (mindweaver server, neoweaver client) has its own version and release cycle.

## Release Tools

### release-please

**Purpose:** Manages independent versioning and creates release PRs

**Configuration:**
- `release-please-config.json` - Component definitions
- `.release-please-manifest.json` - Current versions

**Features:**
- Independent component versioning
- Automatic CHANGELOG generation
- Conventional commit parsing
- Separate release PRs per component
- Version calculation (semantic versioning)

**How it works:**
1. Scans commits merged to `main`
2. Groups commits by scope (`mindweaver`, `neoweaver`) - NOTE: Use exact component names, NOT shortcuts like `nvim`
3. Creates release PR(s) with CHANGELOG updates
4. When release PR merged → creates git tag
5. Tag triggers component-specific release workflow

### goreleaser

**Purpose:** Builds cross-platform binaries for mindweaver server

**Configuration:** `packages/mindweaver/.goreleaser.yml`

**Features:**
- Cross-compilation for multiple OS/arch combinations
- Automatic archive creation with checksums
- GitHub Release artifact upload
- Version embedding via ldflags

**Test locally:**
```bash
cd packages/mindweaver

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
- `.github/workflows/release-please.yml` - Creates release PRs
- `.github/workflows/release-mindweaver.yml` - Builds mindweaver binaries
- `.github/workflows/release-neoweaver.yml` - Packages neoweaver client

**Release workflow:**
1. `release-please.yml` runs on every push to `main`
2. Creates/updates release PRs for affected components
3. When release PR merged → tag created (`mindweaver/v*` or `neoweaver/v*`)
4. Tag triggers component-specific release workflow
5. Artifacts built and attached to GitHub Release

## Versioning

### Version Management

Versions tracked in:
- Component `CHANGELOG.md` files
- `.release-please-manifest.json` (managed automatically by release-please)

### Git Tags

**Format:** `<component>/vX.Y.Z` (e.g., `mindweaver/v0.9.0`, `neoweaver/v0.5.0`)

Tags are created automatically when release PRs are merged.

**List tags:**
```bash
# All mindweaver tags
git tag -l "mindweaver/*"

# All neoweaver tags
git tag -l "neoweaver/*"

# All tags
git tag -l
```

**Create manually (if needed):**
```bash
git tag -a mindweaver/v1.0.0 -m "Release mindweaver v1.0.0"
git push origin mindweaver/v1.0.0
```

**Delete tag (if needed):**
```bash
git tag -d mindweaver/v1.0.0
git push origin :refs/tags/mindweaver/v1.0.0
```

### Version Bump Logic

**Implemented by:** release-please (automatic)

Component determined by commit scope. Only `feat:` and `fix:` trigger release PRs.

**Required Scopes:**
- Use `mindweaver` (NOT `server`, `backend`, or `mind`/`brain` for releases)
- Use `neoweaver` (NOT `nvim`, `neovim`, or `client` for releases)
- Sub-component scopes like `mind`, `brain`, `api` are included in `mindweaver` releases
- Incorrect scopes will NOT trigger release PRs

**Pre-1.0 semantics (0.x.x):**
- `feat:` → MINOR bump (0.9.0 → 0.10.0)
- `fix:` → PATCH bump (0.9.0 → 0.9.1)
- `feat!:` → MINOR bump (0.9.0 → 0.10.0)

**Post-1.0 semantics:**
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

**2. release-please workflow runs:**

Scans commits, groups by scope, calculates version bumps, and creates/updates release PRs with CHANGELOG updates.

**3. Review and merge release PR(s):**

Release PRs can be merged independently. Merging creates a git tag which triggers the component-specific release workflow.

**4. Component-specific workflows run:**

Each component has its own workflow (see `.github/workflows/release-*.yml`):
- mindweaver: GoReleaser builds multi-platform binaries
- neoweaver: Packages client as archives

**5. Releases published:**

Each component gets a separate GitHub Release with its artifacts.

### Manual Release

**Scenario 1: Force specific version**

```bash
# For mindweaver
git commit --allow-empty -m "chore(mindweaver): release 1.0.0

RELEASE-AS: 1.0.0"

# For neoweaver
git commit --allow-empty -m "chore(neoweaver): release 1.0.0

RELEASE-AS: 1.0.0"

git push origin main

# release-please creates PR with specified version
# Merge the release PR to trigger release
```

**Scenario 2: Manual tag and release**

```bash
# Ensure you're on main and up to date
git checkout main
git pull

# Manually update CHANGELOG and manifest
vim packages/mindweaver/CHANGELOG.md
vim .release-please-manifest.json

# Commit changes
git add .
git commit -m "chore(mindweaver): prepare v1.0.0"
git push origin main

# Create and push tag
git tag -a mindweaver/v1.0.0 -m "Release mindweaver v1.0.0"
git push origin mindweaver/v1.0.0

# This triggers release-mindweaver.yml workflow
```

**Scenario 3: Hotfix release**

```bash
# Branch from tag
git checkout -b hotfix-mindweaver-0.9.1 mindweaver/v0.9.0

# Fix bug
git commit -m "fix(mindweaver): critical bug fix"

# Update CHANGELOG and manifest
vim packages/mindweaver/CHANGELOG.md  # Add 0.9.1 section
vim .release-please-manifest.json     # Update to 0.9.1

git add .
git commit -m "chore(mindweaver): bump to 0.9.1"

# Create tag
git tag -a mindweaver/v0.9.1 -m "Hotfix mindweaver v0.9.1"
git push origin mindweaver/v0.9.1

# Tag triggers release workflow automatically

# Merge back to main
git checkout main
git merge hotfix-mindweaver-0.9.1
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

**Each component has its own CHANGELOG:**
- `packages/mindweaver/CHANGELOG.md`
- `clients/neoweaver/CHANGELOG.md`

**Structure (mindweaver):**
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

### API
#### Breaking Changes
- Changed endpoint (#PR)
```

**Structure (neoweaver):**
```markdown
## [Unreleased]

## [X.Y.Z] - YYYY-MM-DD

### Features
- Added UI feature (#PR)

### Bug Fixes
- Fixed display issue (#PR)
```

### Automatic Updates

release-please automatically:
1. Scans commits since last release for that component
2. Groups by conventional commit type
3. Creates new version section in CHANGELOG
4. Includes in release PR
5. CHANGELOG updated when release PR merged

### Manual CHANGELOG Edits

**To improve auto-generated changelog:**

```bash
# Edit CHANGELOG in release PR before merging
gh pr checkout <release-pr-number>
vim packages/mindweaver/CHANGELOG.md

# Improve grouping, add context
git add packages/mindweaver/CHANGELOG.md
git commit -m "docs: improve CHANGELOG formatting"
git push

# Or after release
vim packages/mindweaver/CHANGELOG.md
git add packages/mindweaver/CHANGELOG.md
git commit -m "docs(mindweaver): improve CHANGELOG for v0.10.0"
git push origin main
```

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

### Release PR Not Created

**Check:**
1. Was PR merged to `main`?
2. Does commit message follow conventional commits?
3. Does commit use correct scope (`mindweaver`, `neoweaver`)?
4. Is commit type `feat:` or `fix:`? (others don't trigger releases)
5. Check `.github/workflows/release-please.yml` logs

**Common issues:**
```bash
# Wrong - no scope or wrong scope
git commit -m "feat: add feature"  # No release PR (missing scope)
git commit -m "feat(nvim): add feature"  # No release PR (wrong scope - use 'neoweaver')

# Wrong - commit type doesn't trigger release
git commit -m "chore(mindweaver): update code"  # No release PR

# Correct
git commit -m "feat(mindweaver): add feature"  # Creates mindweaver release PR
git commit -m "feat(neoweaver): add feature"  # Creates neoweaver release PR
```

**Fix:**
```bash
# Create empty commit with correct format
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

### Adding New Components

To add a new independently-versioned component:

1. Add entry to `release-please-config.json` with appropriate `release-type` (go, node, simple)
2. Add entry to `.release-please-manifest.json` starting at version `0.0.0`
3. Create `CHANGELOG.md` in component directory
4. Create component-specific release workflow following the pattern in `.github/workflows/release-mindweaver.yml` or `.github/workflows/release-neoweaver.yml`
5. Use component name as scope in commits (e.g., `feat(imex): add feature`)

See existing workflows and configs for reference patterns.

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
