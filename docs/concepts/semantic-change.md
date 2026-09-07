# Semantic change

A git diff tells you which lines moved. A semantic change tells you what the
system now does differently. With coding agents producing large diffs quickly,
the second question is the one a reviewer cannot answer by reading.

## The problem

```text
git diff --stat

  43 files changed, 2812 insertions(+), 1143 deletions(-)
```

This is a measurement of typing, not of behaviour. Reviewing it means
reconstructing intent from hunks, which is exactly the work that does not scale
when an agent produces several such diffs in an afternoon.

What a reviewer needs:

```text
ARCHITECTURAL DIFF

  + 1 entry point        POST /orders/validate
  + 2 edges              Service.Create → Fraud.Check
                         Fraud.Check → risk_profiles
  - 1 edge               Service.Create → LegacyGateway.Charge

  Flow reordered         Order creation

    before   Create → Payment.Authorize → Fraud.Check
    after    Create → Fraud.Check → Payment.Authorize

  Blast radius           2 entry points · 9 symbols · 4 tests
  Internal refactors     31 edges changed within packages, not enumerated
```

Every line is derived by graph comparison. Nothing is summarised by a model.

## How it is computed

A semantic change is the delta between two snapshots of the structural model:

```text
snapshot A          snapshot B
    │                   │
    └────── diff ───────┘
              │
     symbol set delta       added, removed, signature changed
     entry-point delta      added, removed
     edge set delta         added, removed
              │
              ▼
     flow reordering        derived from edge deltas on flow paths
     blast radius           reverse traversal from changed symbols
```

The rules that keep the output readable:

- **Filter to the observable surface.** Edge changes where both endpoints are
  unexported and live in the same package are counted, not listed. They are real
  but they are not what a reviewer is deciding about.
- **Report reordering, not just membership.** A flow whose steps are the same set
  in a different order is the change most worth surfacing and the one a file diff
  hides best.
- **Never exceed the confidence of the underlying edges.** A change derived from
  an inferred edge is an inferred change, and it says which edge.

See [RFC 0001](../rfcs/0001-repository-intelligence-record.md) for the `change`
record shape and the full `changeKind` list.

## Blast radius

For a changed symbol, reverse-traverse inbound edges and collect what depends on
it:

```text
UserRepository.FindByID  changed

Reached by
  2 entry points     POST /login · GET /profile
  3 symbols          AuthService.Login · ProfileService.Get · UserService.Load
  4 tests

Confidence          inferred at 1 step
                    AuthService.users resolves through interface UserStore
```

This is deterministic traversal. No model is involved, and the result is
checkable: every hop names a file and line.

The honest limitation is that blast radius is bounded by resolution quality. If
an interface call could not be resolved, the radius under-reports. A view that
shows a radius must also show how many unresolved edges were encountered, or it
reads as more complete than it is.

## Risk

RepoSweep v0.1 assigns risk from file paths — a change under a directory named
`auth` is marked security-sensitive. That is a placeholder and it is labelled
`inferred` for good reason.

Real risk signals are structural and are the intended replacement:

```text
an entry point was added or removed
a public signature changed
a dependency edge crossed a module boundary for the first time
a write to a table gained a new caller
a flow reordered around an authorization step
```

Each of these is a graph fact with evidence, not a keyword match on a path.
Until analyzers land, the path heuristic stays, clearly marked as an inference,
because overstating it would be worse than its being shallow.
