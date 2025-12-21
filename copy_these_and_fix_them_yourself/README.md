# Broken Neoweaver Sync Workflows

These workflows are broken and need to be fixed properly.

## Files

- `release-neoweaver.yml.original` - The simple working version before all the PR/docs generation complexity
- `release-neoweaver.yml.broken` - The overcomplicated PR-based version that doesn't work
- `release-neoweaver.yml.current_broken` - The latest attempt with curl instead of gh CLI (still broken)

## What's Wrong

All the "smart" approaches tried to:
1. Clone the mirror repo
2. Copy files manually
3. Create a PR
4. Wait for docs to generate
5. Auto-merge

But they all fail because:
- Deleting `.github/` from mirror repo removes the workflow that generates docs
- Overcomplicated token handling
- Manual file copying is error-prone

## What You Actually Need

Figure out the right approach for syncing to the mirror repo that:
1. Preserves the mirror repo's `.github/workflows/generate-docs.yml`
2. Updates all source files from `clients/neoweaver/`
3. Either waits for docs or doesn't - decide which

Fix it yourself when you have time.
