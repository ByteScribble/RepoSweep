# Writing an analyzer

> **Status: the analyzer host does not exist yet.** This document specifies the
> contract that [RFC 0001](../rfcs/0001-repository-intelligence-record.md)
> implies, so the interface can be argued about before it is built. Do not start
> a large analyzer against it — the shape will move. Small framework rule sets
> and fixtures are useful now.

An analyzer turns source files into Repository Intelligence Records. It is the
main extension point of the project and the place most contributions belong.

## What an analyzer is responsible for

**In scope:**

- Detecting whether it applies to a repository.
- Extracting symbols, entry points, and typed edges from files it understands.
- Resolving references as far as the language allows, and stating how.
- Attaching a file and line to every record it emits.

**Out of scope — these belong to the engine, not to you:**

- Deciding that a change is risky.
- Building flows, computing blast radius, or diffing snapshots.
- Ranking importance.
- Calling a language model. Analyzers are deterministic. See
  [evidence](../concepts/evidence.md).

An analyzer reports facts. Everything interpretive happens downstream, once,
where it can be reasoned about across languages.

## The contract

```go
// Analyzer extracts Repository Intelligence Records from source files.
// Implementations must be deterministic and safe for concurrent use.
type Analyzer interface {
    // Name identifies the analyzer in evidence records, e.g. "analyzer/go".
    Name() string

    // Languages returns the language identifiers this analyzer claims.
    Languages() []string

    // Detect reports whether this analyzer applies to the repository, and
    // with what specificity. A framework analyzer returns a higher score than
    // the base language analyzer for the same repository.
    Detect(repo Repository) (Detection, error)

    // Extract parses one file and returns the records it contains. It must not
    // read other files: cross-file resolution happens in Resolve, after every
    // file has been extracted.
    Extract(file File) ([]Record, error)

    // Resolve upgrades unresolved references once the full symbol table is
    // known. It returns only the records it changed or added, each carrying a
    // resolution chain for anything it marks inferred.
    Resolve(table SymbolTable, pending []Record) ([]Record, error)
}
```

The two-phase split matters. `Extract` is per-file, parallel, and cheap to
re-run when one file changes. `Resolve` is the only phase that needs global
knowledge, and it is the only phase that may produce `inferred` records.

## Extract, then resolve

```text
  file  ──►  Extract  ──►  records with unresolved references
                                        │
      every file extracted              │
                  └─────────────────────┤
                                        ▼
                          Resolve  ──►  resolved edges + resolution chains
```

During `Extract`, an unresolvable call is emitted with `to: null` and an
`unresolved` reason. During `Resolve`, it either gains a target and a resolution
chain, or it stays unresolved and says why. It never gains a guessed target.

## Confidence rules

- A call to a concrete, named target is `observed`.
- A call resolved through an interface, injection, registry, or dynamic
  dispatch is `inferred` and **must** carry a `resolution` chain.
- Ambiguity is reported, not resolved. Three candidate implementations produce
  an unresolved edge listing three candidates, not one edge for the likeliest.

The third rule is the one that gets violated under pressure to produce a
complete-looking graph. A guessed edge is indistinguishable from a real one
downstream, which makes every projection built on it unfalsifiable.

## Framework rules

Framework knowledge does not belong in a language analyzer. A Go analyzer knows
about functions and interfaces; it does not know what `chi` or `gin` mean.

```text
  language analyzer   →   structural records
          │
          ▼
  framework rules     →   entry points and typed edges
```

Rules are pattern matches over the parsed tree, contributed independently:

```text
rules/
  go/
    chi/          r.Post("/orders", h.Create)        → http_route
    gin/          r.POST("/orders", h.Create)        → http_route
    cron/         c.AddFunc("0 3 * * *", job)        → scheduled_job
  typescript/
    express/      router.post("/orders", handler)    → http_route
    nextjs/       app/api/**/route.ts                → http_route
  java/
    spring/       @PostMapping                       → http_route
                  @KafkaListener                     → queue_consumer
                  @Autowired                         → INJECTS edge
```

A framework rule set is the best first contribution: it is self-contained,
testable against a fixture, and does not require understanding the engine.

## Testing

Every analyzer ships fixtures. A fixture is a small repository under
`testdata/` with the records it should produce recorded alongside it.

```text
internal/analyzer/go/testdata/
  chi-routes/
    repo/
      main.go
      handlers.go
    expected.json
```

Tests assert on the full record set, including evidence locations. A change that
moves a reported line number is a change to the contract and should fail
loudly — a wrong line number destroys the reader's ability to check a claim,
which is the only reason the record exists.

Fixtures should include the awkward cases, not only the clean ones: an interface
with several implementations, a route registered in a loop, a handler referenced
by a variable. An analyzer that only has clean fixtures will guess in production.

## Before you start

Open an issue first if you are adding a language, a framework, or a new record
or edge kind. New edge kinds require an RFC amendment — see
[RFC 0002](../rfcs/0002-graph-model.md).

Read [evidence](../concepts/evidence.md) before writing any resolution code. It
is the constraint most likely to be violated by accident, and the one the project
will not compromise on.
