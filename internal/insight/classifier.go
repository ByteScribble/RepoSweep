package insight

import (
	"path/filepath"
	"strings"

	"github.com/ByteScribble/RepoSweep/internal/events"
)

func Classify(path, operation string) events.Event {
	clean := filepath.ToSlash(path)
	lower := strings.ToLower(clean)
	base := filepath.Base(clean)
	title := humanOperation(operation) + " " + base
	summary := "Repository file " + operation + "."
	risk := "low"
	confidence := "observed"
	evidence := []string{clean, "filesystem:" + operation}

	switch {
	case strings.Contains(lower, "migration") || strings.Contains(lower, "schema"):
		title = "Data shape may have changed"
		summary = base + " changed in a schema or migration path. Review compatibility and rollback behavior."
		risk = "high"
		confidence = "inferred"
	case isDependencyManifest(lower):
		title = "Dependency surface changed"
		summary = base + " changed. Build inputs or resolved dependencies may differ."
		risk = "high"
		confidence = "observed"
	case strings.Contains(lower, "auth") || strings.Contains(lower, "permission") || strings.Contains(lower, "policy"):
		title = "Security-sensitive code changed"
		summary = base + " changed in a path associated with authorization or policy."
		risk = "high"
		confidence = "inferred"
	case strings.Contains(lower, "route") || strings.Contains(lower, "controller") || strings.Contains(lower, "handler") || strings.Contains(lower, "/api/"):
		title = "Entry point changed"
		summary = base + " changed in an API or request-handling path."
		risk = "medium"
		confidence = "inferred"
	case isTest(lower):
		title = "Test evidence changed"
		summary = base + " changed. This affects verification evidence, not necessarily product behavior."
		risk = "low"
		confidence = "inferred"
	case isConfig(lower):
		title = "Configuration changed"
		summary = base + " changed and may alter build or runtime behavior."
		risk = "medium"
		confidence = "inferred"
	case strings.HasSuffix(lower, ".md"):
		title = "Documentation changed"
		summary = base + " changed. No runtime impact is claimed."
		risk = "low"
		confidence = "observed"
	}

	return events.Event{
		Kind: "change", Status: "observed", Path: clean, Title: title, Summary: summary,
		Risk: risk, Confidence: confidence, Evidence: evidence,
	}
}

func humanOperation(operation string) string {
	switch operation {
	case "created":
		return "Created"
	case "removed":
		return "Removed"
	case "renamed":
		return "Renamed"
	default:
		return "Modified"
	}
}

func isDependencyManifest(path string) bool {
	base := filepath.Base(path)
	return base == "go.mod" || base == "go.sum" || base == "package.json" || base == "package-lock.json" ||
		base == "pnpm-lock.yaml" || base == "yarn.lock" || base == "cargo.toml" || base == "cargo.lock" ||
		base == "requirements.txt" || base == "poetry.lock" || base == "pyproject.toml"
}

func isTest(path string) bool {
	return strings.Contains(path, "/test/") || strings.Contains(path, "/tests/") || strings.Contains(path, "_test.") ||
		strings.Contains(path, ".test.") || strings.Contains(path, ".spec.")
}

func isConfig(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, ".yaml") || strings.HasSuffix(base, ".yml") || strings.HasSuffix(base, ".toml") ||
		strings.HasSuffix(base, ".ini") || strings.HasSuffix(base, ".env") || strings.Contains(path, "/config/")
}
