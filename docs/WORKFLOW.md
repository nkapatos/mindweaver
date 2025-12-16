# Workflow Documentation

This document captures the development workflow, versioning strategy, and release process for Mindweaver. These guidelines ensure consistency across contributions and releases.

## Table of Contents

- [Development Workflow](#development-workflow)
- [Branching Strategy](#branching-strategy)
- [Commit Conventions](#commit-conventions)
- [Versioning Strategy](#versioning-strategy)
- [Release Process](#release-process)
- [PR Guidelines](#pr-guidelines)

## Development Workflow

### Local Development

```bash
# Clone repository
git clone https://github.com/nkapatos/mindweaver.git
cd mindweaver

# Create feature branch (any name)
git checkout -b your-feature-name

# Develop with hot reload
task dev

# Run tests
task test

# Commit changes (any format during development)
git commit -m "wip: working on feature"
git commit -m "fix that bug"
git commit -m "more changes"

# Push branch
git push origin your-feature-name

# Create PR (title MUST follow conventional commits)
```

**Key Points:**
- Branch names are freeform - use whatever makes sense to you
- Commit messages during development are freeform
- Only PR titles need to follow conventional commit format
- Hot reload with `task dev` for rapid iteration
- Run tests before creating PR

### For Maintainers

**Merge Strategy:**
- **Squash merge (default)**: For messy commits → single clean commit on main
- **Rebase merge (optional)**: For PRs with clean, logical commits
- Decision made per-PR based on commit quality

**Branch Cleanup:**
- Keep PR branches briefly (30 days) for reference
- Contributors' work remains visible in closed PRs
- Delete branches manually after sufficient time

## Branching Strategy

### Main Branch

- `main` is the primary branch and is **protected**
- Always stable and releasable
- Direct pushes disabled (even for maintainers)
- All changes via Pull Requests

### Feature Branches

- Create from `main` for any new work
- Use any naming convention: `feature/x`, `fix-bug`, `dev`, etc.
- Work, commit, push freely
- Delete after PR is merged

### No Long-Lived Branches

- No `develop` or `staging` branches
- Simpler workflow: feature → main → release
- GitHub Flow model

## Commit Conventions

### During Development (On Feature Branches)

**Anything goes!** Commit messages can be informal:
```bash
git commit -m "wip"
git commit -m "trying something"
git commit -m "fixed the fucking bug"
```

### PR Titles (REQUIRED Format)

PR titles MUST follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description
```

**Examples:**
```
feat(mindweaver): add advanced search operators
feat(neoweaver): add buffer management improvements
fix(mind): resolve collection hierarchy bug
fix(neoweaver): correct note save behavior
docs: update installation instructions
feat(brain)!: breaking change to embedding API
```

**Types:**
- `feat`: New feature (→ MINOR version bump)
- `fix`: Bug fix (→ PATCH version bump)
- `feat!` or `BREAKING CHANGE`: Breaking change (→ MAJOR bump for 1.x+, MINOR for 0.x)
- `docs`: Documentation only
- `chore`: Maintenance tasks (no version bump)
- `refactor`: Code refactoring (no version bump)
- `test`: Test changes (no version bump)
- `ci`: CI/CD changes (no version bump)

**Scopes:**

**Release-triggering scopes (required for component releases):**
- `mindweaver`: Main server binary (packages/mindweaver) - triggers mindweaver release PR
- `neoweaver`: Neovim client (clients/neoweaver) - triggers neoweaver release PR

**Sub-component scopes (included in parent component's release):**
- `mind`: Mind service within mindweaver server
- `brain`: Brain service within mindweaver server
- `api`: API-level changes affecting clients

**General scopes:**
- `pkg`: Shared packages/libraries
- `deps`: Dependency updates
- `ci`: CI/CD specific
- `proto`: Protocol buffer changes
- `docs`: Documentation changes

**Future component scopes:**
- `imex`: Import/export tool (future)
- `lsp`: LSP server (future)
- `web`: Web client (future, clients/web)
- `desktop`: Desktop client (future, clients/desktop)

**Important Notes:**
- **Use `neoweaver` (NOT `nvim`)** for Neovim client changes - this ensures proper release triggering and changelog updates
- **Use `mindweaver` (NOT `server` or `backend`)** for server changes
- Changes with `mind`, `brain`, or `api` scopes are included in the `mindweaver` component release
- Using incorrect scopes (e.g., `nvim` instead of `neoweaver`) will prevent release-please from creating release PRs
- Scope is optional but **strongly recommended** for clarity and proper release targeting

## Versioning Strategy

### Semantic Versioning

Mindweaver follows [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH (e.g., 0.9.0, 1.2.3)
```

**Version Bumps:**
- **MAJOR**: Breaking changes (for 1.x.x+)
- **MINOR**: New features or breaking changes (for 0.x.x)
- **PATCH**: Bug fixes

**Pre-1.0 Semantics:**
- Currently at `0.x.x` (pre-stable)
- `feat` → MINOR bump (0.9.0 → 0.10.0)
- `fix` → PATCH bump (0.9.0 → 0.9.1)
- `feat!` → MINOR bump (breaking changes allowed in 0.x)
- Move to `1.0.0` when ready for stable release

### Independent Component Versioning

The monorepo uses **independent versioning** for each releasable component. Each component evolves at its own pace with separate version numbers.

**Current Components:**

| Component | Path | Version Managed By | Git Tag Format |
|-----------|------|-------------------|----------------|
| **mindweaver** | `packages/mindweaver/` | release-please | `mindweaver/vX.Y.Z` |
| **neoweaver** | `clients/neoweaver/` | release-please | `neoweaver/vX.Y.Z` |

**Server (mindweaver):**
```
mindweaver binary v0.12.5
├─ Mind service (included)
└─ Brain service (included)
```

The mindweaver server is a single binary containing both Mind and Brain services:
- **Single version** for the binary
- **CHANGELOG sections** differentiate service-specific changes:
  - `### Mind Service` - Changes to Mind service
  - `### Brain Service` - Changes to Brain service  
  - `### API` - API-level changes affecting clients

**Clients (neoweaver, future: web, desktop):**
- Each client has its **own independent version**
- Can release without server changes
- Document compatibility: "requires mindweaver >= X.Y.Z"

**Shared Libraries (`packages/pkg`):**
- Currently not independently versioned
- Used internally by mindweaver
- Future: May become releasable Go modules

**Version Source of Truth:**
- `packages/mindweaver/CHANGELOG.md` - mindweaver version
- `clients/neoweaver/CHANGELOG.md` - neoweaver version
- Managed automatically by release-please
- Git tags use component prefix: `mindweaver/v1.2.3`, `neoweaver/v0.5.0`

**Compatibility Matrix:**

Components track compatibility requirements in their README:

```markdown
# neoweaver v0.5.0
Requires: mindweaver >= v0.10.0
```

This allows:
- Server updates without forcing client updates
- Client UI/UX improvements independent of server
- Clear compatibility communication to users

## Release Process

The project uses [release-please](https://github.com/googleapis/release-please) for automated, independent releases of each component.

### How It Works

**1. Development Flow:**
```bash
# Make changes to mindweaver server
git commit -m "feat(mindweaver): add search operators"

# Make changes to neoweaver client  
git commit -m "fix(neoweaver): resolve buffer display bug"

# Changes affecting both
git commit -m "feat(mindweaver): add new API endpoint"
git commit -m "feat(neoweaver): integrate new API endpoint"

# Merge PR to main
```

**2. Automatic Release PR Creation:**

When commits are merged to `main`, release-please:
- Scans commit messages for conventional commit format
- Groups commits by scope (`mindweaver`, `neoweaver`, etc.)
- Creates separate release PRs for each affected component
- Each release PR updates CHANGELOG.md and can be merged independently

**3. Merging Release PRs:**

Release PRs can be merged independently:
- Merge server release PR → creates `mindweaver/v*` tag → triggers build
- Merge client release PR → creates `neoweaver/v*` tag → triggers packaging
- Merge both in any order → creates both releases independently

**4. Release Artifacts:**

After merging a release PR, the component-specific workflow builds and publishes artifacts:
- mindweaver: GoReleaser builds multi-platform binaries
- neoweaver: Packaged as tar.gz and zip archives
- Each creates a separate GitHub Release

### Cross-Component Changes

When a PR affects multiple components (using different scopes in commits), release-please creates separate release PRs for each component. These can be merged independently based on readiness.

### Manual Release Override

To force a specific version (e.g., moving to 1.0.0):

```bash
# For mindweaver
git commit --allow-empty -m "chore(mindweaver): release 1.0.0

RELEASE-AS: 1.0.0"

# For neoweaver  
git commit --allow-empty -m "chore(neoweaver): release 1.0.0

RELEASE-AS: 1.0.0"

git push origin main
```

Release-please will create a release PR with the specified version.

### Skip Release

To skip release PR creation for certain commits:

```bash
# Commits that don't trigger releases:
git commit -m "docs: update README [skip ci]"
git commit -m "chore: code cleanup"
git commit -m "refactor(mindweaver): internal restructure"

# Only feat:, fix:, and BREAKING CHANGE trigger version bumps
```

### Configuration Files

**Release-please configuration:**
- `release-please-config.json` - Component definitions and settings
- `.release-please-manifest.json` - Current version tracking

**Example `.release-please-manifest.json`:**
```json
{
  "packages/mindweaver": "0.11.0",
  "clients/neoweaver": "0.6.0"
}
```

Updated automatically by release-please when release PRs are merged.

## PR Guidelines

### Creating a PR

1. **Branch from main:**
   ```bash
   git checkout main
   git pull
   git checkout -b your-feature
   ```

2. **Develop and commit freely:**
   - Any commit format during development
   - Commit often

3. **Create PR with proper title:**
   ```
   Title: feat(mind): add advanced search operators
   Description: Detailed explanation of changes
   ```

4. **CI checks run automatically:**
   - Tests must pass
   - Linting must pass
   - Build must succeed
   - PR title format validated (warning only)

5. **Code review:**
   - Maintainer reviews code
   - Address feedback
   - Commit changes (any format)

6. **Merge:**
   - Maintainer decides: squash or rebase
   - PR merged to main
   - Branch can be kept or deleted

### For Contributors

**Clean commits (optional but appreciated):**

If comfortable with git, clean up before PR:
```bash
# Interactive rebase to clean commits
git rebase -i main

# Squash "wip" commits
# Reword commit messages to be meaningful
# Use conventional commit format
```

This is **optional** - maintainer can squash merge if needed.

**Required: PR title format**

Only the PR title must follow conventional commits. This becomes the commit message on `main`.

### For Maintainers

**Merge Decision Matrix:**

| PR Commit Quality | Action |
|------------------|---------|
| Messy (wip, fixes, etc.) | **Squash merge** - Creates single clean commit |
| Clean logical commits | **Rebase merge** - Preserves commit history |
| Mix of good and bad | **Squash merge** - Safest option |

**Squash Merge:**
- Combines all commits into one
- PR title becomes commit message
- Individual commits visible in closed PR (for reference)
- Contributors credited automatically

**Rebase Merge:**
- Preserves all commits as-is
- Only if commits are clean and follow conventions
- Less common, case-by-case decision

## CI/CD Pipelines

### CI (Pull Requests)

**Trigger:** Every PR to `main`

**Jobs:**
- `test`: Run all tests with race detection
- `lint`: Run golangci-lint
- `build`: Verify binary builds

**Must pass** before merge (enforced by branch protection).

### PR Title Check

**Trigger:** PR opened, edited, or synchronized

**Action:** Posts warning comment if PR title doesn't follow format

**Non-blocking:** Can still merge, just a reminder

### Release (Main Branch)

**Trigger:** Push to `main` (after PR merge)

**Conditions:**
- Changes to relevant paths
- Commit matches `feat:`, `fix:`, or `BREAKING CHANGE`
- No `[skip ci]` in message

**Output:** New release on GitHub with binaries

## GitHub Repository Settings

### Branch Protection (main)

Required settings:
- ✅ Require pull request before merging
- ✅ Require status checks to pass (CI must pass)
- ✅ Require conversation resolution before merging
- ✅ Allow squash merging
- ✅ Allow rebase merging
- ❌ Allow merge commits (disabled)
- ❌ Allow direct pushes (disabled for everyone)
- ❌ Auto-delete PR branches (disabled - keep for 30 days manually)

### Required Secrets

- `GITHUB_TOKEN`: Automatically provided by GitHub Actions
- `CODECOV_TOKEN`: (Optional) For code coverage reporting

## Future Considerations

### Adding New Components

**Example: Adding `clients/web` or `packages/imex`**

1. Add entry to `release-please-config.json` with appropriate `release-type` and `component` name
2. Add entry to `.release-please-manifest.json` starting at `0.0.0`
3. Create component CHANGELOG.md
4. Create release workflow following the pattern of existing component workflows (see `.github/workflows/release-mindweaver.yml` and `.github/workflows/release-neoweaver.yml`)
5. Document compatibility requirements in component README

### When Moving to 1.0.0

**For each component independently:**

- **Decide readiness:** Server vs. clients may reach 1.0.0 at different times
- **Manual release trigger** with `RELEASE-AS: 1.0.0`
- **Update documentation:** Installation, compatibility matrix
- **Announcement plan:** Blog post, release notes
- **Consider:** Post-1.0 release strategy (auto-release patches only?)

**Example: mindweaver reaches 1.0.0 first**
```bash
# Server is stable
git commit --allow-empty -m "chore(mindweaver): release 1.0.0

RELEASE-AS: 1.0.0"

# Clients remain at 0.x.x until they're ready
# neoweaver v0.8.0 - requires mindweaver >= v1.0.0
```

### Shared Go Modules

**If `packages/pkg` needs independent versioning:**

1. Make it a separate Go module
2. Add to release-please config
3. Use Go module versioning conventions
4. Import as: `github.com/nkapatos/mindweaver/pkg@v1.2.3`

Currently, `packages/pkg` is part of the workspace and doesn't need independent releases.

## Quick Reference

**Development:**
```bash
git checkout -b feature

# Server changes
git commit -m "feat(mindweaver): add feature"

# Client changes  
git commit -m "fix(neoweaver): fix bug"

git push origin feature
# Create PR with conventional title
```

**Version Bumps (per component):**
- `feat(mindweaver):` → mindweaver 0.9.0 → 0.10.0
- `fix(mindweaver):` → mindweaver 0.9.0 → 0.9.1
- `feat(neoweaver):` → neoweaver 0.5.0 → 0.6.0
- `fix(neoweaver):` → neoweaver 0.5.0 → 0.5.1
- `feat!:` → MINOR bump (pre-1.0), MAJOR bump (post-1.0)

**Tags (component-prefixed):**
- `mindweaver/v0.9.0`, `mindweaver/v0.10.0`, `mindweaver/v1.0.0`
- `neoweaver/v0.5.0`, `neoweaver/v0.6.0`, `neoweaver/v1.0.0`

**Release Process:**
1. Merge PR to main
2. release-please creates release PR(s)
3. Merge release PR(s) independently
4. GitHub releases created automatically

**Merge Strategy:**
- Messy commits → Squash
- Clean commits → Rebase  
- When in doubt → Squash

**Cross-component changes:**
- Use correct scope for each commit
- Multiple release PRs created automatically
- Merge release PRs independently based on readiness

---

**Last Updated:** 2025-12-12  
**Version:** 0.9.0
