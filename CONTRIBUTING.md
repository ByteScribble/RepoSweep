# Contributing

RepoSweep is early and intentionally opinionated. Contributions should strengthen the local, evidence-first repository intelligence core.

## Start here

1. Read [docs/concepts/evidence.md](docs/concepts/evidence.md). It is the constraint most contributions are judged against, and the one most easily violated by accident.
2. Skim [ARCHITECTURE.md](ARCHITECTURE.md) for how the shipping binary is built, and the [documentation index](docs/README.md) for where the project is going.
3. Open an issue before adding a new language, framework, model provider, or public record type. Changes to a public contract need an [RFC](docs/rfcs/README.md) first.
4. Keep the executable local-first and functional without an AI provider.
5. Add tests for classifiers, event contracts, concurrency boundaries, and failure behavior.

## Development

```sh
go test ./...
go vet ./...
go build ./cmd/reposweep
```

Run the dashboard against this repository:

```sh
go run ./cmd/reposweep watch .
```

## Contribution tracks

- **Watcher:** platform behavior, ignore rules, journals, Git evidence.
- **Analyzer:** language and framework facts with fixtures and provenance. See [writing an analyzer](docs/analyzers/writing-an-analyzer.md); framework rule sets are the best first contribution.
- **Semantic:** flow/change records that never overstate confidence.
- **Agent:** hooks and event adapters for coding tools.
- **UI:** accessible projections of the evidence model.
- **Performance:** benchmarks, allocation reduction, and incremental algorithms.

Small, independently testable changes are preferred. Generated dependencies, cloud-only core features, opaque model claims, and any view whose size scales with the repository are out of scope.
