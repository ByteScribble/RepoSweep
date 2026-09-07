# Flows

A flow is an ordered narrative of what happens when something triggers the
system. It is the project's main unit of navigation, in place of a directory
tree.

## Why flows instead of folders

A directory tree tells you how the code was filed. It does not tell you how it
behaves. Two engineers can know every file in a repository and still disagree
about what happens when a request arrives.

```text
   instead of              show
   ──────────              ────

   src/                    Order creation
     controller/           Checkout
     service/              User signup
     repository/           Refund processing
     client/               Nightly reconciliation
```

The second list is what an engineer actually holds in their head. The first is
an implementation detail of the filesystem.

## Shape

A flow starts at an entry point and proceeds in order:

```text
ORDER CREATION

Entry point
POST /orders

  Handler.Create
       ↓
  Service.Create
       ↓
  Inventory.Reserve
       ↓
  Payment.Authorize          ⚠ inferred — interface, 1 implementation
       ↓
  orders table               WRITES_TO
       ↓
  order.created              PUBLISHES
```

Each step names a real symbol and links to its evidence. A step is expandable to
show its callers, its own outbound edges, its failure modes, and its recent
changes.

## Flows are derived, never authored

A flow is a traversal of the graph, not a document someone writes. Nobody
maintains a list of flows; they fall out of entry-point detection plus edge
traversal. See [RFC 0002](../rfcs/0002-graph-model.md) for the traversal rules.

Two consequences follow:

**A flow is only as trustworthy as its weakest edge.** If one step resolved
through an interface with three candidate implementations, the flow says so at
that step and its overall confidence drops to match. It does not quietly pick
one.

**A flow changes when the code changes, without anyone updating it.** This is the
entire point. Hand-maintained architecture documentation is wrong within weeks,
and it is wrong silently. A derived flow is either correct or visibly uncertain.

## Naming

Flows are named after the entry point's observable trigger, not after the code
that implements it: *Order creation*, not *OrderServiceImpl*. When a good name
cannot be derived, the entry point itself is the name — `POST /orders` is a
worse title than *Order creation* but an honest one, and better than a guess.

An optional one-sentence `purpose` may be written by a model enrichment layer.
It is advisory, it is labelled as such, and removing it does not change the
flow. See [evidence](evidence.md).

## What a flow is not

- **Not a call graph rendering.** A call graph shows every edge; a flow shows one
  ordered path from a trigger, which is what a person can read.
- **Not a runtime trace.** Flows are static. Correlating them with real traces is
  future work and will be a separate record type, because a trace proves a path
  was taken while a flow only shows one exists.
- **Not exhaustive.** A flow follows the primary path. Error handling, retries,
  and rarely-taken branches are shown on the step, not inlined into the
  narrative, or the narrative stops being readable.
