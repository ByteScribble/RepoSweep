# Documentation

## Concepts

Read these in order. They build on each other.

- [Evidence](concepts/evidence.md) — the confidence vocabulary, resolution
  chains, and the boundary a language model may not cross. Start here; the rest
  of the project follows from it.
- [The repository model](concepts/repository-model.md) — the structural,
  semantic, and evolution models, and why they stay separate.
- [Flows](concepts/flow.md) — the unit of navigation that replaces the
  directory tree.
- [Semantic change](concepts/semantic-change.md) — architectural diff, blast
  radius, and what replaces the current path-based risk heuristic.
- [Interface](concepts/interface.md) — semantic zoom, the four screens, and
  where a language model sits in the interface.

## Design

- [V0 screen designs](design/canvas/README.md) — artboards for the four screens
  and two overlays, and how to rebuild them.

## Contributing analyzers

- [Writing an analyzer](analyzers/writing-an-analyzer.md) — the extraction
  contract, the extract/resolve split, and framework rule sets.

## Design decisions

- [RFC index](rfcs/README.md) — when an RFC is required and how one moves.
- [RFC 0001 — Repository Intelligence Record](rfcs/0001-repository-intelligence-record.md)
- [RFC 0002 — Graph model and projections](rfcs/0002-graph-model.md)

## Elsewhere in the repository

- [ARCHITECTURE.md](../ARCHITECTURE.md) — how the shipping v0.1 binary is built.
- [ROADMAP.md](../ROADMAP.md) — what is planned and what is deliberately not.
- [CONTRIBUTING.md](../CONTRIBUTING.md) — tracks, workflow, and scope.
- [SECURITY.md](../SECURITY.md) — the local trust boundary.

## A note on status

The concepts and RFCs describe the model RepoSweep is being built toward. The
shipping binary implements the evolution model and a path-based classifier, and
nothing more. Where a document describes something not yet built, it says so.
That gap is deliberate and stated rather than papered over.
