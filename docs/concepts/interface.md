# Interface

> **Status: mostly unbuilt.** RepoSweep v0.1 ships a single event timeline. This
> document records the interface decisions the project is designed toward, so
> that view work does not have to relitigate them. Where it describes something
> not built, it says so.
>
> Static mockups of the four screens and both overlays live in
> [docs/design/canvas](../design/canvas/README.md).

## The governing question

Most tools in this category make the same mistake:

> We have a graph, therefore let us visualise the graph.

The correct framing:

> The engineer has a question. Show the smallest part of the graph that answers
> it.

That single decision separates a tool people use from a technically impressive
demo. It is the interface expression of the bounds rule in
[RFC 0002](../rfcs/0002-graph-model.md).

Navigation follows the question, not the storage:

```text
   yes    repository → domain → flow → component → method → source
   no     repository → giant graph → pan around until something makes sense
```

## Semantic zoom

Detail is disclosed on request, never by default. Each level is a bounded
diagram — a handful of nodes, sized by the question rather than by the
repository.

**Level 1 — architecture.** What this repository exposes and what it talks to.

```text
   APIs  →  Controllers  →  Services  →  Database · Events · External
```

**Level 2 — domain.** One area, expanded.

```text
   PaymentController → PaymentService → Fraud · DB · Gateway
```

**Level 3 — component.** One symbol and its immediate neighbours.

**Level 4 — method.** One function's ordered steps.

**Level 5 — source.** The file, at the line the evidence names.

An engineer should never be shown detail they did not ask for. This is what keeps
a diagram from becoming the mess the project exists to eliminate.

## The four screens

V0 is four states done well, not a suite. Each answers one of the questions the
project was built for.

| Screen | Question |
|---|---|
| **Repository overview** | What is this repository? |
| **Execution flow** | How does this behaviour actually work? |
| **Component inspector** | What is this thing and what touches it? |
| **Architecture diff** | What structurally changed, and what does it affect? |

Search and ask are overlays, not screens. If those four interactions are
excellent, that is a compelling product; adding a fifth before they are is how
the surface area gets away from you.

The default screen is the overview, and it answers with an inventory rather than
a picture of everything:

```text
payments-service
Payment processing and transaction orchestration

  12 APIs · 18 services · 7 repositories · 4 consumers · 3 external
```

No charts, no metric cards, no activity feed competing for attention, no
400-node hairball. The repository is the interface.

## Navigation shell

Keep the global shell small. Five destinations at most:

```text
   Explore · Flows · Changes · Dependencies · Settings
```

Files, classes, methods, controllers, services, endpoints, tables and topics are
**ways of viewing the graph**, not product areas. Putting them in global
navigation is how a focused tool becomes a file browser with extra steps.

Breadcrumbs are not optional once a view is more than two levels deep:

```text
   payments-service / Payments / PaymentService / process()
```

Every segment navigable. In an interface that moves between zoom levels, the
engineer must always be able to answer "where am I?".

## Inspection preserves context

Selecting a node opens a contextual inspector — roughly 340–380px — rather than
navigating to a new page. The surrounding context stays visible behind it.

An inspector shows, in this order: what the thing is, who calls it, what it
calls, what it touches (data, events, external systems), its members, and its
location. Location is last and always present, because a claim without a file
and line is not a claim. See [evidence](evidence.md).

Hover answers "what is this?" in under three seconds — name, one line, and the
caller and call counts. Basic understanding should not require a click.

## Rendering relationships

Where a bounded diagram is drawn, three rules keep it readable:

**Dim the irrelevant.** On selection, drop unrelated nodes to 20–30% prominence
and highlight only the selected node's callers, calls, and effects. This is what
makes even a hundred-node view comprehensible, and it does more for legibility
than any layout algorithm.

**Label edges on demand.** Every edge labelled at all times is clutter:

```text
   default    A ───── B ───── C ───── D
   selected   A ─── CALLS ──▶ B
```

**Distinguish edge weight, not edge colour.** Solid for the focused path, thin
for default, dashed for secondary. Colour is reserved for meaning.

## Visual language

Aim for Linear × GitHub × Figma × IDE, not Datadog. Neutral surfaces, one
accent, semantic colour only where it carries information — HTTP, database,
event, external, error, success. Do not colour every node; the eye should land
on relationships, not decoration.

A 4px spacing base (4, 8, 12, 16, 20, 24, 32, 40, 48), a proportional family for
UI and a monospace family for symbols, paths, and routes. Not everything needs
to be a rounded card — whitespace separates information perfectly well.

The one hard constraint the shipping dashboard already meets: the embedded UI
uses no framework and adds no runtime dependency. Any view work must keep that
true, or it stops being one boring binary.

## Where the model sits

A language model does not get a permanent third of the screen. The
graph-plus-chat-sidebar layout is becoming an anti-pattern: it occupies the most
valuable space with the least reliable content.

Instead, a collapsible composer. An answer cites its evidence and offers to
navigate:

```text
   Ask about this codebase...                                    ⌘J

   PaymentService calls FraudService.check() before gateway
   authorization in PaymentService.process().

   Evidence
     PaymentService.java:72
     FraudService.java:31

   [ Show path ]  [ Open source ]
```

"Show path" makes the interface itself the answer. That is a far stronger
interaction than paragraphs of prose, and it holds the model to the same
evidence standard as everything else — it describes the graph, it does not
invent it. See [evidence](evidence.md).

## Search

Search spans semantic entities, not filenames: symbols, routes, events, tables,
flows. Exact deterministic search is always available and always first. Natural
language querying can sit alongside it, never in place of it.
