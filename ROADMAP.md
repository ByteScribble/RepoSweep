# Roadmap

## v0.1 — Live repository pulse

- [x] Single local executable
- [x] Native recursive watcher
- [x] Evidence-labeled change timeline
- [x] Agent event ingestion
- [x] Embedded live dashboard
- [ ] Gitignore-compatible filtering
- [ ] Durable local event journal
- [ ] Release builds for macOS, Linux, and Windows

## v0.2 — Structural accuracy

- Versioned Repository Intelligence Record ([RFC 0001](docs/rfcs/0001-repository-intelligence-record.md))
- Typed graph and projections ([RFC 0002](docs/rfcs/0002-graph-model.md))
- Analyzer host with the extract/resolve split
- Go packages, symbols, interfaces, and HTTP entry points — **first analyzer**
- TypeScript/JavaScript imports, exports, routes, and call evidence
- Git status, diff hunks, branch, and commit correlation
- Incremental analysis keyed by content hash

**Why Go first.** RepoSweep is written in Go, so the first analyzer can be
dogfooded against this repository from the day it compiles, and `go/packages`
provides real type-checked symbol resolution instead of hand-rolled inference.
An earlier proposal argued for Java and Spring on the grounds of framework
richness; that remains a strong second target, but validating the record format
against a language we run daily is worth more than breadth at this stage.

## v0.3 — Semantic change

- Before/after flow projections
- Blast-radius candidates with explicit confidence
- Test/build event adapters
- Codex, Claude Code, and generic hook integrations
- MCP read interface for agent context

## Later

- OpenTelemetry correlation
- Java/Spring, Python/FastAPI/Django, and Rust analyzers
- Optional local and remote model enrichment behind a provider-neutral interface
- Plugin SDK and compatibility test suite

Roadmap items are proposals, not promises. Evidence integrity and incremental performance take priority over feature count.

Decisions that affect a public contract are recorded as [RFCs](docs/rfcs/README.md) before they are implemented.
