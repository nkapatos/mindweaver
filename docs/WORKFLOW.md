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
fix(mind): resolve collection hierarchy bug
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
- `mindweaver`: Main server binary or combined changes
- `mind`: Mind service specifically
- `brain`: Brain service specifically
- `imex`: Import/export tool (future)
- `lsp`: LSP server (future)
- `nvim`: Neovim client (future)
- `web`: Web client (future)
- `desktop`: Desktop client (future)
- `api`: API-level changes
- `deps`: Dependency updates
- `ci`: CI/CD specific

**Scope is optional** but recommended for clarity.

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

### Single Version for All Components

**Architecture:**
```
mindweaver binary v0.9.0
├─ Mind service (included)
└─ Brain service (included)
```

**One version number** for the entire mindweaver binary that includes both services.

**CHANGELOG sections:**
- `### Mind Service` - Changes to Mind service
- `### Brain Service` - Changes to Brain service
- `### Combined` - Changes affecting both or the binary itself

**Future Components:**
- `imex`, `lsp`, `nvim`, `web`, `desktop` will have independent versions
- Each with own release workflow
- Each documents "requires mindweaver >= X.Y.Z"

### Version Source of Truth

```
cmd/mindweaver/VERSION
```

- Plain text file with version number
- Read by automation
- Updated automatically on release
- Git tags: `mindweaver/vX.Y.Z`

## Release Process

### Automatic Release (Default)

**Trigger:** Merge to `main` branch

**Process:**
1. GitHub Action detects merge to main
2. Analyzes commit message for type (`feat`, `fix`, etc.)
3. Calculates version bump
4. Updates `cmd/mindweaver/VERSION`
5. Creates git tag `mindweaver/vX.Y.Z`
6. Builds binaries with goreleaser:
   - `mindweaver_X.Y.Z_darwin_arm64.tar.gz`
   - `mindweaver_X.Y.Z_darwin_amd64.tar.gz`
   - `mindweaver_X.Y.Z_linux_amd64.tar.gz`
   - `mindweaver_X.Y.Z_linux_arm64.tar.gz`
7. Creates GitHub Release (pre-release for 0.x.x)
8. Updates CHANGELOG.md
9. Commits changes back to main

**Only triggers on changes to:**
- `cmd/mindweaver/**`
- `internal/mind/**`
- `internal/brain/**`
- `pkg/**`
- `migrations/**`
- `go.mod` / `go.sum`

**Skip Release:**
- Commits with `[skip ci]` in message
- Commits that don't match `feat:`, `fix:`, or `BREAKING CHANGE`

### Manual Release (When Needed)

To manually create a release or transition versions:

```bash
# Create empty commit with version override
git commit --allow-empty -m "chore(mindweaver): release v1.0.0

RELEASE-AS: 1.0.0"

git push origin main
```

This triggers the workflow with explicit version.

### Release Artifacts

Each release includes:
- Compiled binaries for multiple platforms
- `README.md`
- `docs/DEVELOPMENT.md`
- `.env.example`
- SHA256 checksums
- Auto-generated release notes from commits

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

### When Adding New Components

**Example: Adding `imex` CLI tool**

1. Create `cmd/imex/VERSION` file
2. Create `.github/workflows/release-imex.yml`
3. Configure path triggers for `cmd/imex/**`
4. Independent versioning: `imex/vX.Y.Z`
5. Document in `cmd/imex/README.md`: "Requires mindweaver >= X.Y.Z"

### When Moving to 1.0.0

- Manual release trigger
- Announcement plan
- Documentation review
- Consider: Auto-release only patches? Manual for minor/major?

## Quick Reference

**Development:**
```bash
git checkout -b feature
# work, commit freely
git push origin feature
# Create PR with conventional title
```

**Version Bumps:**
- `feat:` → 0.9.0 → 0.10.0
- `fix:` → 0.9.0 → 0.9.1
- `feat!:` → 0.9.0 → 0.10.0 (pre-1.0)
- `feat!:` → 1.0.0 → 2.0.0 (post-1.0)

**Tags:**
- `mindweaver/v0.9.0`
- `mindweaver/v0.10.0`
- `mindweaver/v1.0.0`

**Merge Strategy:**
- Messy commits → Squash
- Clean commits → Rebase
- When in doubt → Squash

---

**Last Updated:** 2025-12-12  
**Version:** 0.9.0
