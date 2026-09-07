package web

import "embed"

// Files contains the complete local dashboard and is embedded into the binary.
//
//go:embed index.html app.css app.js
var Files embed.FS
