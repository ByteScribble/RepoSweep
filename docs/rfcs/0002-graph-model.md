# RFC 0002 — Graph model and projections

- **Status:** Draft
- **Created:** 2026-09-07
- **Depends on:** [RFC 0001](0001-repository-intelligence-record.md)

## Summary

RepoSweep stores a typed graph and shows bounded answers. This RFC defines the
edge vocabulary that makes the graph queryable, and states the rule that keeps
the interface usable: **no view renders more of the graph than the question
requires.**

## Motivation

Two things are true at once, and the project needs both.

A graph is the right way to *store* code structure. Blast radius is reverse
traversal. An architectural diff is an edge-set delta. A flow is a walk from an
entry point. None of these are expressible over a pile of per-file summaries.

A whole graph is the wrong thing to *show*. A force-directed rendering of a real
repository is an illegible hairball, and it gets worse as the repository gets
more valuable to understand. Every adjacent tool in this space has learned this
the expensive way.

So: build the whole graph, show a bounded piece of it.

## The rule

```text
     typed graph            projections
     ───────────            ───────────
     symbols                flow narratives
     entry points     ──►   change timelines
     typed edges            blast-radius lists
     evidence               entry-point inventories
```

The rule is about **bounds**, not about pictures:

> No view renders more of the graph than the question requires.

Diagrams are allowed and useful. What is prohibited is rendering the graph
*because it exists* — the whole-repository node-edge view that every tool in this
category eventually ships and that no engineer can read.

**Allowed.** A bounded diagram answering one question: the steps of one flow, one
symbol and its immediate neighbours, the before-and-after of one architectural
change. These stay legible because their size is set by the question, not by the
repository.

**Prohibited.** Any view whose node count scales with repository size. A
force-directed layout of every symbol. A dependency diagram of all packages at
once. "Zoom out to see everything."

Two consequences follow for anyone building a view:

1. **Every projection must also be readable as ordered text.** A list of steps, a
   table, a count with a list. If a view is only comprehensible as a picture,
   it cannot be shown in the CLI, quoted in a review, or read by an agent —
   and it is probably unbounded.
2. **Detail is disclosed, never defaulted.** Start at the smallest useful
   answer and expand on request. See [interface](../concepts/interface.md) for
   the semantic zoom levels this implies.

## Edge vocabulary

An untyped `CONNECTED_TO` relation is worthless: it cannot answer "what writes to
this table" or "which entry points reach this function". Every edge is typed.

### v0 edge kinds

| Kind | From | To | Typical confidence | Meaning |
|---|---|---|---|---|
| `CALLS` | function, method | function, method | observed / inferred | Invocation. Inferred when the receiver is an interface. |
| `IMPORTS` | module, file | module, package | observed | Static import or require. |
| `IMPLEMENTS` | type | interface | observed / inferred | Explicit in Java and TypeScript; structural in Go. |
| `EXTENDS` | type | type | observed | Inheritance. |
| `INJECTS` | type, function | type, interface | inferred | Dependency injection wiring. Always carries a `resolution` chain. |
| `READS_FROM` | function, method | table, collection, cache key | observed / inferred | Data read. |
| `WRITES_TO` | function, method | table, collection, cache key | observed / inferred | Data write. |
| `PUBLISHES` | function, method | topic, queue, event | observed / inferred | Emits a message or event. |
| `CONSUMES` | entrypoint | topic, queue, event | observed | Subscribes to a message or event. |
| `EXPOSES` | module, type | entrypoint | observed | Registers an externally reachable trigger. |
| `TRIGGERS` | schedule, event | entrypoint | observed | Timer or event source that starts a flow. |

### Deliberately excluded

- **`DEPENDS_ON`** — derivable from the typed edges above. Admitting it lets an
  analyzer avoid the work of typing a relation, which is exactly the work that
  makes the graph useful.
- **`RETURNS` and `CREATES`** — these describe a symbol's type relationships and
  belong on the `symbol` record's `inputs` and `outputs`, not as graph edges.
  Promoting them to edges doubles the edge count without enabling a query.
- **`CONTAINS`** — containment is expressed by the `container` field on a symbol
  record. Duplicating it as an edge creates two sources of truth for the tree.

Adding an edge kind requires an RFC amendment. The vocabulary staying small is
what lets a projection author reason about the whole graph.

## Edge confidence

The distinction that matters most in practice is a direct call versus an injected
one.

```text
s.payments.Authorize(ctx, order)

  payments has concrete type *payment.Client
  → CALLS, observed
```

```text
s.gateway.Charge(ctx, amount)

  gateway has interface type payment.Gateway
  one implementation registered in wire.go:31
  → CALLS, inferred
  → resolution: ["s.gateway is payment.Gateway",
                 "single binding at internal/app/wire.go:31 provides *StripeGateway"]
```

```text
s.gateway.Charge(ctx, amount)

  gateway has interface type payment.Gateway
  three implementations, no binding determines which
  → CALLS, inferred, to: null
  → unresolved: { expression: "s.gateway.Charge",
                  reason: "interface with 3 implementations" }
```

The third case is the one that separates a trustworthy tool from a plausible
one. An analyzer that picks a likely implementation to avoid an empty result
produces a graph nobody can rely on, and the failure is silent.

## Projections

Each projection is a defined traversal. Naming them here keeps analyzer authors
aware of what their edges will be used for.

### Flow

Walk forward from an entry point along `CALLS`, `PUBLISHES`, `READS_FROM`, and
`WRITES_TO`, depth-first, recording the first visit to each symbol. Stop at
repository boundaries and at unresolved edges. Render as an ordered list of
steps, never as a picture.

Flow confidence is the minimum confidence of the edges walked.

### Blast radius

Reverse traversal from a symbol along inbound `CALLS`, `IMPLEMENTS`, and
`INJECTS` edges, collecting the entry points reached. Rendered as counts plus a
list, ordered by distance:

```text
UserRepository.FindByID

Reached by
  2 entry points     POST /login, GET /profile
  3 symbols          AuthService.Login, ProfileService.Get, UserService.Load
  4 tests
```

### Architectural diff

The symmetric difference of two edge sets, filtered to edges whose endpoints are
entry points or exported symbols. Internal refactors that add and remove edges
within one package are reported as a count, not enumerated, because they are
noise at review time.

### Entry-point inventory

A grouped count of `entrypoint` records. This is the repository overview: what
this repository exposes, before any question about how it is organised.

## Storage

v0 keeps the graph in memory, rebuilt on start, sized for repositories up to
roughly 100k symbols. That is enough to build and validate every projection.

Persistence, incremental invalidation keyed by content hash, and any embedded
store are deferred until the projections are proven. Choosing a database before
knowing the query patterns is how this kind of project acquires a dependency it
cannot remove.

The incremental path this is designed for:

```text
file change → reparse one file → changed symbols → invalidate their edges
            → recompute edges for those symbols only → patch graph → diff → event
```

Only edges whose `from` endpoint lives in a changed file need recomputation.
Inbound edges to a changed symbol are found by reverse index and do not require
reparsing their source files unless the symbol's identity changed.

## Open questions

1. Should an unresolved edge with N candidates be one edge with a candidate list,
   or N edges each marked low confidence? The first is honest; the second makes
   blast radius over-report rather than under-report, which may be the safer
   failure for a review tool.
2. Do generated files belong in the graph, marked as generated, or outside it?
   Excluding them breaks call resolution; including them floods the diff.
3. Should `READS_FROM` and `WRITES_TO` targets be modelled as first-class symbol
   records (a `table` symbol kind) or as opaque string identifiers?
4. At what repository size does the in-memory graph stop being viable, and what
   measurement decides that?
