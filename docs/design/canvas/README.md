# V0 screen designs

Design artboards for the four screens and two overlays described in
[docs/concepts/interface.md](../../concepts/interface.md).

| Artboard | Screen |
|---|---|
| `Main.dc.html` | Repository overview |
| `Flow.dc.html` | Execution flow |
| `Inspector.dc.html` | Component inspector |
| `Diff.dc.html` | Architecture diff |
| `Search.dc.html` | Search overlay (⌘K) |
| `Ask.dc.html` | Ask overlay (⌘J) |

These are static mockups at 1440×900. They are not the shipping dashboard —
that lives in [`web/`](../../../web) and is deliberately much smaller.

## Structure

```text
_tokens.css              palette, type scale, shared chrome
_topbar.html             top bar, shared by every artboard
_rail.html               left navigation rail, shared

part-main.*              repository overview
part-flow.*              execution flow
part-inspector.*         component inspector
part-diff.*              architecture diff
overlay-search.css       search overlay styling
overlay-ask.css          ask overlay styling
part-search-overlay.html search overlay markup
part-ask-overlay.html    ask overlay markup

build.sh                 assembles artboards from the parts above
canvas.json              artboard layout
```

Edit the parts, never the artboards. `*.dc.html` is generated:

```sh
sh docs/design/canvas/build.sh
```

The search and ask screens are derived from the overview and inspector parts
respectively, so those two screens have exactly one source of truth. Their
intermediate `part-search.*` and `part-ask.*` files are generated and git-ignored.

## Design tokens

The palette, spacing base, and type families follow
[docs/concepts/interface.md](../../concepts/interface.md). Changing a token here
without changing that document leaves the two out of sync — update both, or
neither.
