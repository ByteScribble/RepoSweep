# RFC 0001 — Repository Intelligence Record

- **Status:** Draft
- **Created:** 2026-09-07
- **Affects:** analyzer plugins, graph engine, HTTP API, MCP interface

## Summary

The Repository Intelligence Record (RIR) is the language-neutral format every
RepoSweep analyzer emits and every consumer reads. It is the project's one
stable contract: analyzers depend on it, the graph engine is built from it, and
the dashboard, agent API, and future MCP interface are projections over it.

An analyzer author should never need to read the graph engine. A UI author
should never need to read an analyzer. The RIR is the seam between them.

## Motivation

RepoSweep v0.1 classifies changes from file paths alone. That is honest but
shallow: `internal/auth/policy.go` looks security-sensitive because of the word
`auth`, not because anything about the code was understood.

Replacing that with real analysis raises an immediate structural question. If
each language analyzer emits its own shape, the graph engine grows a branch per
language and contributors have to understand the whole system to add anything.
A normalized record solves this the same way an IR solves it for a compiler:

```text
Go analyzer    ─────┐
TypeScript     ─────┤
Python         ─────┼──►  RIR  ──►  Graph engine  ──►  Projections
Java/Spring    ─────┤
Framework rules ────┘
```

## Design principles

1. **Every record carries its evidence.** A record without a file and line is
   not a record. See [evidence](../concepts/evidence.md).
2. **Confidence is categorical, not numeric.** `observed`, `reported`, and
   `inferred` mean something a reader can verify. `0.92` does not.
3. **Relations are records, not string arrays.** An edge needs its own
   evidence and confidence, which a bare string in a `calls: []` list cannot
   carry. See [RFC 0002](0002-graph-model.md).
4. **Analyzers report; they do not interpret.** An analyzer states that a call
   expression resolves to a symbol. It does not decide the change was risky.
5. **Unknown is a valid answer.** An analyzer that cannot resolve a target emits
   an unresolved edge rather than guessing. Guessing poisons every projection
   built on top.

## The envelope

Every record shares one envelope:

```json
{
  "rir": "0",
  "id": "go:github.com/acme/shop/internal/order#Service.Create",
  "kind": "symbol",
  "confidence": "observed",
  "evidence": [
    {
      "source": "analyzer/go",
      "file": "internal/order/service.go",
      "startLine": 42,
      "endLine": 78
    }
  ]
}
```

| Field | Required | Meaning |
|---|---|---|
| `rir` | yes | Format version. `"0"` is unstable and may change without migration. |
| `id` | yes | Stable identity. See [Identity](#identity). |
| `kind` | yes | One of `symbol`, `entrypoint`, `edge`, `flow`, `change`. |
| `confidence` | yes | `observed`, `reported`, or `inferred`. |
| `evidence` | yes | At least one location. Empty arrays are invalid. |

### Identity

IDs must be stable across runs and across unrelated edits, because the change
engine diffs record sets by ID. The scheme is:

```text
<language>:<module>#<symbol-path>
```

```text
go:github.com/acme/shop/internal/order#Service.Create
ts:@acme/web/src/checkout/api.ts#submitOrder
java:com.acme.shop.order#OrderService.create
http:route#POST /orders
kafka:topic#order.created
```

Line numbers are never part of an ID. Moving a function within its file must not
change its identity, or every reformat becomes an architectural change.

### Evidence

```json
{
  "source": "analyzer/go",
  "file": "internal/order/service.go",
  "startLine": 51,
  "endLine": 51,
  "snippet": "s.payments.Authorize(ctx, order)"
}
```

`source` names the analyzer that made the claim, so a wrong record can be traced
to the plugin that produced it. `snippet` is optional and should be a single
line where possible; it exists so a reader can confirm a claim without opening
an editor.

## Record kinds

### `symbol`

A declared thing.

```json
{
  "rir": "0",
  "id": "go:github.com/acme/shop/internal/order#Service.Create",
  "kind": "symbol",
  "symbolKind": "method",
  "language": "go",
  "name": "Create",
  "container": "go:github.com/acme/shop/internal/order#Service",
  "signature": "func (s *Service) Create(ctx context.Context, req CreateRequest) (*Order, error)",
  "inputs": ["context.Context", "CreateRequest"],
  "outputs": ["*Order", "error"],
  "confidence": "observed",
  "evidence": [{ "source": "analyzer/go", "file": "internal/order/service.go", "startLine": 42, "endLine": 78 }]
}
```

`symbolKind` is one of `module`, `package`, `type`, `interface`, `function`,
`method`, `field`, `constant`.

### `entrypoint`

Something the outside world can trigger. Entry points are the roots of every
flow, and the repository overview is built by counting them.

```json
{
  "rir": "0",
  "id": "http:route#POST /orders",
  "kind": "entrypoint",
  "entrypointKind": "http_route",
  "trigger": "POST /orders",
  "handler": "go:github.com/acme/shop/internal/order#Handler.Create",
  "confidence": "inferred",
  "resolution": [
    "chi router registration at internal/order/routes.go:18",
    "handler method value resolves to Handler.Create"
  ],
  "evidence": [{ "source": "analyzer/go/chi", "file": "internal/order/routes.go", "startLine": 18, "endLine": 18, "snippet": "r.Post(\"/orders\", h.Create)" }]
}
```

`entrypointKind` is one of `http_route`, `graphql_resolver`, `cli_command`,
`scheduled_job`, `queue_consumer`, `event_handler`, `websocket_handler`,
`serverless_handler`, `main`.

### `edge`

A typed relation. The vocabulary is defined in [RFC 0002](0002-graph-model.md);
this RFC defines only the record shape.

```json
{
  "rir": "0",
  "id": "edge:calls#go:...order#Service.Create->go:...payment#Client.Authorize",
  "kind": "edge",
  "edgeKind": "CALLS",
  "from": "go:github.com/acme/shop/internal/order#Service.Create",
  "to": "go:github.com/acme/shop/internal/payment#Client.Authorize",
  "confidence": "observed",
  "evidence": [{ "source": "analyzer/go", "file": "internal/order/service.go", "startLine": 51, "endLine": 51, "snippet": "s.payments.Authorize(ctx, order)" }]
}
```

When the target cannot be resolved, say so rather than guessing:

```json
{
  "edgeKind": "CALLS",
  "from": "go:github.com/acme/shop/internal/order#Service.Create",
  "to": null,
  "unresolved": { "expression": "s.gateway.Charge", "reason": "interface with 3 implementations" },
  "confidence": "inferred",
  "resolution": ["s.gateway has interface type payment.Gateway", "3 implementations found, no injection site determines which"]
}
```

`resolution` is required whenever `confidence` is `inferred`. It is the chain a
reader follows to check the analyzer's reasoning, and it is what makes dependency
injection defensible instead of magical.

### `flow`

An ordered narrative from an entry point. A flow is a **projection**, not a
source of truth: it is derived by traversing edges, and it may be recomputed at
any time. It is recorded because the change engine diffs flows between snapshots.

```json
{
  "rir": "0",
  "id": "flow:order-creation",
  "kind": "flow",
  "name": "Order creation",
  "entrypoint": "http:route#POST /orders",
  "steps": [
    { "symbol": "go:...order#Handler.Create", "edge": null },
    { "symbol": "go:...order#Service.Create", "edge": "edge:calls#..." },
    { "symbol": "go:...payment#Client.Authorize", "edge": "edge:calls#..." },
    { "symbol": "kafka:topic#order.created", "edge": "edge:publishes#..." }
  ],
  "confidence": "inferred",
  "evidence": [{ "source": "engine/flow", "file": "internal/order/service.go", "startLine": 42, "endLine": 78 }]
}
```

A flow is never more confident than its least confident edge.

`purpose` — the one-sentence human description of what a flow does — is an
optional field written only by an enrichment layer, never by an analyzer, and
never by traversal. It is the single place a language model may contribute, and
it is advisory. See [evidence](../concepts/evidence.md).

### `change`

A semantic delta between two snapshots of the model.

```json
{
  "rir": "0",
  "id": "change:2026-09-07T08:14:02Z#flow:order-creation",
  "kind": "change",
  "changeKind": "flow_reordered",
  "subject": "flow:order-creation",
  "before": ["...order#Service.Create", "...payment#Client.Authorize", "...fraud#Service.Check"],
  "after": ["...order#Service.Create", "...fraud#Service.Check", "...payment#Client.Authorize"],
  "blastRadius": { "entrypoints": 2, "symbols": 9, "tests": 4 },
  "confidence": "inferred",
  "resolution": ["edge set for flow:order-creation differs between snapshot A and B"],
  "evidence": [{ "source": "engine/change", "file": "internal/order/service.go", "startLine": 51, "endLine": 63 }]
}
```

`changeKind` is one of `symbol_added`, `symbol_removed`, `signature_changed`,
`entrypoint_added`, `entrypoint_removed`, `edge_added`, `edge_removed`,
`flow_reordered`, `dependency_added`, `dependency_removed`.

See [semantic change](../concepts/semantic-change.md).

## Relationship to the v0.1 event stream

RepoSweep already ships an event model in `internal/events`. The two are not
competing:

```text
events.Event   what happened, when, and who did it   — the evolution stream
RIR record     what exists and how it relates        — the model
```

A filesystem write produces an event immediately. Analysis of that write later
produces RIR records, and a `change` record links back to the event sequence
that triggered it via `metadata.event_sequence`. The dashboard timeline stays
event-driven; the flow and architecture views read the model.

The confidence vocabulary is deliberately shared between them, so a reader learns
it once.

## What v0 deliberately excludes

- **Numeric confidence.** `0.92` cannot be verified by a reader and invites
  averaging scores that mean nothing. Categorical classes plus a `resolution`
  chain carry the same information honestly.
- **A generic `DEPENDS_ON` relation.** It is derivable from typed edges and
  its presence encourages analyzers to skip the work of typing a relation.
- **Runtime records.** Traces and spans correlate with this model but have a
  different lifetime and volume. They deserve their own RFC.
- **Cross-repository identity.** IDs are repository-scoped in v0. Federating
  several repositories into one graph changes the identity scheme and is out of
  scope until single-repository analysis is solid.
- **Model-generated topology.** No record may be created by a language model.
  Only `flow.purpose` and similar descriptive fields may be model-written, and
  they are advisory.

## Compatibility

`rir: "0"` is explicitly unstable. Records may change shape without a migration
path until this RFC reaches Accepted and the version becomes `"1"`. Analyzer
authors should expect churn and pin to a RepoSweep commit.

Once `"1"` ships, additive fields are allowed within a major version; removing a
field, changing its meaning, or changing the identity scheme requires a new
major version and a new RFC.

## Open questions

1. Should `flow` records be persisted at all, or always recomputed on read? They
   are derived, but the change engine needs a previous value to diff against.
2. Should an analyzer emit records for vendored and generated code, marked as
   such, or skip them? Skipping breaks call resolution into third-party code.
3. What is the right granularity for `symbol` records in languages with heavy
   closure and callback use, where the meaningful unit is not a named function?
4. How should an unresolved edge with several candidate targets be surfaced —
   one edge with a candidate list, or one low-confidence edge per candidate?
