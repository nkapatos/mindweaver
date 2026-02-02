> [!IMPORTANT]
> Imex is still under heavy development and a lot of v1 functionality hasn't been migrated yet.

# imex CLI

Imex is the import/export tool for Mindweaver data. It is a standalone binary designed to interact with Mindweaver server via the v3 API proto contracts under /proto/mind/v3/.

## Status
Work in progress, basic project scaffold only.

## Configuration
See `config-example.yaml` for standalone configuration format.

## Developer Notes
- imex requires up-to-date generated Go client code under `/gen`, matching the proto contracts in `/proto/mind/v3`.
- If proto files are changed, run `task mw:proto:generate` from the repo root to regenerate shared code.
- In CI, proto generation is handled globally before any app/binary builds.
- imex does not run code generation on its own; it consumes generated code only.
