# Version Management Rules

  ## NEVER Assume or Change Versions
ALWAYS check project configuration files before using any tool or dependency version.
 Version Sources (in order of precedence)
1. `go.mod` - Go version and Go dependencies
2. `.mise.toml` - Tool versions (Go, Task, sqlc, buf, etc.)
3. `package.json` - Node/npm dependencies (if applicable)
  ## Before Any Operation
- Read the relevant config file
- Use the EXACT version specified
- Do NOT substitute or "correct" versions
- Do NOT assume latest versions
  ## Examples
- Writing CI workflows → Check `go.mod` and `.mise.toml` first
- Running tools → Verify tool version matches `.mise.toml`
- Adding dependencies → Check existing version patterns in `go.mod`
  ## On Version Conflicts
- STOP and report the conflict
- Do NOT choose a version yourself
- Wait for user guidanceODO: Add version management rules (e.g., always check go.mod and .mise.toml, never assume versions)
