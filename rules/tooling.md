# Tooling Rules

 ## Task CLI Usage
ALWAYS use Task commands for all development operations.
 Never Run Directly
- `sqlc generate` → use appropriate task command
- `go build` → use `task mw:build` or equivalent
- `go test` → currently allowed directly, but check for task alternatives
- Database operations → use `task mw:db:*` commands
## Discovery
- List available tasks: `task --list` or `task -l`
- If unsure which task to use: STOP and ASK
## On Task Failure
- STOP immediately (see global blocking conditions)
- Do NOT run commands directly as workaround
- Do NOT attempt alternative approaches
- Report failure and wait for instructions
## Why This Matters
- Task commands ensure correct build output locations (e.g., `/bin/` directory)
- Direct commands may create artifacts in wrong locations
- Prevents accidental commits of build artifacts
- Maintains consistency with team workflow
