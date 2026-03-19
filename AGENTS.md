# Repository Guidelines

## Project Structure & Module Organization
Core CLI entrypoint lives at `cmd/search/main.go` (flags, routing, command flow). Internal modules are organized by concern:
- `internal/engine/`: search engine implementations and shared engine interfaces/types.
- `internal/reader/`: article fetch/extract logic and fallback behavior.
- `internal/ui/`: terminal output formatting.
- `internal/util/`: setup helpers (for example, Lightpanda setup).

Repository root contains operational files such as `install.sh`, `README.md`, and `CHANGELOG.md`. Release automation is in `.github/workflows/release.yml` (build + publish on `v*` tags).

## Build, Test, and Development Commands
- `go build -o search cmd/search/main.go`: build local CLI binary.
- `./search "golang generics"`: quick smoke test for default engine flow.
- `./search -read "https://go.dev/blog/go1.22" -save`: validate reader + markdown save path.
- `go test ./...`: run all tests.
- `go test ./internal/engine -run TestEngines -v`: run engine-focused integration-style tests.
- `bash install.sh`: test installer behavior against latest GitHub release.

## Coding Style & Naming Conventions
Use standard Go formatting (`gofmt`) before committing. Keep package names lowercase and short (`engine`, `reader`, `ui`, `util`). Follow existing file naming style by feature/provider (for example, `duckduckgo.go`, `hackernews.go`). Prefer explicit error handling and concise CLI messages. Keep flag behavior consistent with existing conventions (`-e`, `-read`, `-save`, `-hn`, `-i`).

## Testing Guidelines
Place tests next to implementation files using `_test.go` and `TestXxx` naming. Prefer table-driven tests for parser or transformation logic. Some engine tests are network-dependent and may be rate-limited; log transient failures clearly and keep deterministic assertions for local logic where possible.

## Commit & Pull Request Guidelines
Follow the existing commit pattern: `<type>: <summary>` (examples: `feat: ...`, `fix: ...`, `chore: ...`). Keep summaries imperative and focused. PRs should include:
- what changed and why,
- commands run for verification,
- sample CLI output when behavior changes,
- release impact notes when touching `install.sh`, workflow files, or tagged builds.
