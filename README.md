# RepoSweep

**Live, local repository intelligence for humans working with coding agents.**

An agent changes 37 files. Git can show `+3,291 -1,804`; RepoSweep shows the live sequence, the agent's reported intent, the paths touched, deterministic risk signals, and the evidence behind every claim.

RepoSweep is an early open-source prototype. It runs as one local executable, watches a repository through native filesystem events, and serves an embedded dashboard. Source code is not uploaded and a model provider is not required.

## What works in v0.1

- Native recursive repository watching with noisy build/vendor paths excluded.
- Debounced live changes over Server-Sent Events—no browser refresh.
- Agent intent and status ingestion through a small CLI/API contract.
- Deterministic change classification with `observed`, `reported`, and `inferred` confidence labels.
- Active-agent attribution for file changes during a reported session.
- Bounded in-memory history and a responsive dashboard embedded in the binary.
- Secure-by-default loopback binding.

RepoSweep deliberately does **not** claim that a filename heuristic proves application behavior. Flow extraction, Git semantic diffs, and AST-backed analyzers are next-stage work.

The current binary is intentionally lean: the dashboard is embedded, uses no JavaScript framework, and adds no runtime process beyond RepoSweep itself.

## Build and run

Requires Go 1.22 or newer.

```sh
go build -o reposweep ./cmd/reposweep
./reposweep watch /path/to/repository
```

Open [http://127.0.0.1:7331](http://127.0.0.1:7331).

From an agent hook or another terminal, report intent:

```sh
./reposweep emit \
  --agent codex \
  --kind intent \
  --message "Add authorization checks to the account API"
```

Any file events that follow are attributed to that active agent. End the session explicitly:

```sh
./reposweep emit --agent codex --kind session --status done --message "Implementation complete"
```

The HTTP equivalent is intentionally simple:

```sh
curl -X POST http://127.0.0.1:7331/api/events \
  -H 'content-type: application/json' \
  -d '{"agent":"claude","kind":"test","status":"failed","message":"4 auth tests failed"}'
```

## Design principles

1. **Local is the default.** The core has no cloud dependency.
2. **Evidence before explanation.** Every insight exposes where it came from.
3. **Deterministic before probabilistic.** Optional models may enrich facts, never replace them.
4. **Flows over giant graphs.** The eventual semantic model is projected as human-readable execution narratives.
5. **One useful binary.** Installation and removal should stay boring.

## Project layout

```text
cmd/reposweep/       CLI entry point
internal/events/     append-only event model and fan-out
internal/insight/    deterministic change classification
internal/watcher/    native recursive repository watcher
internal/server/     local API and SSE transport
web/                 embedded dependency-free dashboard
docs/                concepts, analyzer contract, and RFCs
```

## Documentation

Start with [docs/concepts/evidence.md](docs/concepts/evidence.md) — the
confidence vocabulary and the boundary a language model may not cross. Most
other decisions in the project follow from it.

- [Documentation index](docs/README.md)
- [RFC 0001 — Repository Intelligence Record](docs/rfcs/0001-repository-intelligence-record.md), the analyzer contract
- [RFC 0002 — Graph model and projections](docs/rfcs/0002-graph-model.md)
- [Writing an analyzer](docs/analyzers/writing-an-analyzer.md)

See also [ARCHITECTURE.md](ARCHITECTURE.md), [ROADMAP.md](ROADMAP.md), and [CONTRIBUTING.md](CONTRIBUTING.md).

## Status

Prototype quality. The event and analyzer contracts will evolve before the first tagged release. See [SECURITY.md](SECURITY.md) before binding the server to a non-loopback interface.

## License

MIT. See [LICENSE](LICENSE).
