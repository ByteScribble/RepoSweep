# RFCs

Design decisions that affect a public contract are written down before they are
implemented. An RFC exists so a contributor can disagree with a decision by
reading one document instead of reverse-engineering the code.

## Index

| RFC | Title | Status |
|---|---|---|
| [0001](0001-repository-intelligence-record.md) | Repository Intelligence Record | Draft |
| [0002](0002-graph-model.md) | Graph model and projections | Draft |

## When an RFC is required

- Changing the Repository Intelligence Record.
- Adding or changing an edge kind.
- Adding a language, framework, or model provider to the core.
- Changing the confidence vocabulary or the evidence contract.
- Adding a public HTTP, CLI, or MCP surface.

Bug fixes, new tests, UI work within existing projections, and analyzer rules
that emit only existing record kinds do not need one.

## Status values

- **Draft** — proposed, open to fundamental disagreement. May change without notice.
- **Accepted** — agreed. Changes require an amendment.
- **Implemented** — shipped and matching the code.
- **Superseded** — replaced. The header links to the replacement.

## Process

1. Copy an existing RFC as a starting point and number it sequentially.
2. Open a pull request with status Draft.
3. Argue about it in the pull request, not in the code.
4. A maintainer marks it Accepted, or it closes with the reasoning recorded.

An RFC that turned out to be wrong stays in the tree, marked Superseded. The
history of a rejected decision is worth more than a clean directory.
