package insight

import "testing"

func TestClassifierDoesNotOverclaimDocumentationImpact(t *testing.T) {
	event := Classify("docs/architecture.md", "modified")
	if event.Title != "Documentation changed" || event.Confidence != "observed" || event.Risk != "low" {
		t.Fatalf("unexpected classification: %#v", event)
	}
}

func TestClassifierFlagsSchemaAsHighRiskInference(t *testing.T) {
	event := Classify("db/migrations/004_users.sql", "created")
	if event.Risk != "high" || event.Confidence != "inferred" {
		t.Fatalf("unexpected classification: %#v", event)
	}
}
