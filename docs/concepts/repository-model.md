# The repository model

RepoSweep maintains three models of a repository. They answer different
questions, change at different rates, and are built by different parts of the
system. Keeping them separate is what stops the project from becoming one large
undifferentiated graph.

```text
  structural          what exists              rebuilt from source
  semantic            what it does             derived from structural
  evolution           what is happening        appended over time
```

## Structural model

Deterministic facts extracted from source by analyzers.

```text
module · package · file · type · interface · function · method · field
imports · calls · implementations · entry points · tables · topics
```

This model has no opinions. It does not know that a change is risky or that a
service is important. It knows what is declared and what refers to what, and
every fact carries a file and line.

Built by: analyzer plugins.
Rate of change: on every file write, incrementally.
Confidence: mostly `observed`, `inferred` where resolution is required.

## Semantic model

Projections over the structural model that a person can read.

```text
entry-point inventory · flows · blast radius · boundaries · failure modes
```

Nothing here is stored as a separate truth. A flow is a walk over structural
edges; a blast radius is a reverse walk. If the structural model changes, these
change with it, and they can always be recomputed from scratch.

This is where the project's presentation rule applies: these projections render
as ordered text, never as a node-edge picture. See
[RFC 0002](../rfcs/0002-graph-model.md).

Built by: the graph engine.
Rate of change: whenever the structural model changes.
Confidence: never higher than the weakest structural fact it traverses.

## Evolution model

An append-only record of what happened, in order.

```text
filesystem event · agent intent · test result · build result
commit · branch · semantic change
```

This is the model RepoSweep v0.1 ships today, in `internal/events`. It is the
one that makes the tool live rather than a report generator, and it is the one
that answers the question the project exists for: *an agent just changed
thirty-seven files — what happened?*

Built by: the watcher and the event API.
Rate of change: continuously.
Confidence: `observed` for measurements, `reported` for integrations.

## How they connect

```text
   file write  ──►  evolution: event recorded immediately
                             │
                             ▼
                    structural: reparse that file, patch the graph
                             │
                             ▼
                    semantic: recompute affected projections
                             │
                             ▼
   evolution: semantic change recorded, linked to the originating event
```

The evolution model is written twice for a single change: once immediately, with
only what was observed, and once after analysis, with what it meant. The first
write is what keeps the interface responsive while analysis is still running. The
second is what makes it useful.

This ordering is deliberate. A tool that waits for full analysis before showing
anything feels broken during exactly the moments it is needed.

## Why not one graph

Collapsing these into a single graph is tempting and wrong for three reasons.

**They have different lifetimes.** Structural facts are true until the code
changes. Evolution facts are true forever. Storing a filesystem event as a graph
node means either garbage-collecting history or growing without bound.

**They have different confidence semantics.** An observed file write and an
inferred call edge are both `observed`-class facts in isolation, but a projection
over them must not average or combine those judgements.

**They are queried differently.** The structural model is traversed. The
evolution model is scanned in time order. One storage shape serves neither well.
