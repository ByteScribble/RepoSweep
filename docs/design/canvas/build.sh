#!/bin/sh
# Assembles each artboard from the shared chrome snippets + its own content part.
# Regenerate after editing any part:  sh build.sh
set -e
cd "$(dirname "$0")"

emit() {
  name="$1"; active="$2"; css="$3"; content="$4"; overlay="$5"
  out="${name}.dc.html"
  {
    printf '%s\n' '<!doctype html>'
    printf '%s\n' '<html>'
    printf '%s\n' '<head>'
    printf '%s\n' '  <meta charset="utf-8">'
    printf '%s\n' '  <script src="./support.js"></script>'
    printf '%s\n' '</head>'
    printf '%s\n' '<body>'
    printf '%s\n' '<x-dc>'
    printf '%s\n' '<helmet>'
    printf '%s\n' '  <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600;700&amp;family=Geist+Mono:wght@400;500;600&amp;display=swap">'
    printf '%s\n' '  <style>'
    cat _tokens.css
    cat "$css"
    printf '%s\n' '  </style>'
    printf '%s\n' '</helmet>'
    printf '%s\n' '<div class="screen">'
    cat _topbar.html
    printf '%s\n' '  <div class="body">'
    sed "s/ACT${active}/active/; s/ACT[0-9]//g" _rail.html
    cat "$content"
    printf '%s\n' '  </div>'
    if [ -n "$overlay" ]; then cat "$overlay"; fi
    printf '%s\n' '</div>'
    printf '%s\n' '</x-dc>'
    printf '%s\n' '</body>'
    printf '%s\n' '</html>'
  } > "$out"
  echo "built $out"
}


# --- derived parts (single source of truth: the four primary screens) ---
cat part-main.css overlay-search.css > part-search.css
cp part-main.html part-search.html
cat part-inspector.css overlay-ask.css > part-ask.css
python3 - <<'PYEOF'
src = open('part-inspector.html').read()
a = src.index('        <aside class="inspector">')
b = src.index('</aside>', a) + len('</aside>\n')
open('part-ask.html','w').write(src[:a] + src[b:])
PYEOF

emit Main      1 part-main.css      part-main.html      ""
emit Flow      2 part-flow.css      part-flow.html      ""
emit Inspector 1 part-inspector.css part-inspector.html ""
emit Diff      3 part-diff.css      part-diff.html      ""
emit Search    1 part-search.css    part-search.html    part-search-overlay.html
emit Ask       1 part-ask.css       part-ask.html       part-ask-overlay.html
