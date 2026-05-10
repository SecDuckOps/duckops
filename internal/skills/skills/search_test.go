package skills

import (
	"strings"
	"testing"
)

func TestResolveSkillSupportsApproximateNames(t *testing.T) {
	reg, err := NewEmbeddedRegistry()
	if err != nil {
		t.Fatalf("NewEmbeddedRegistry() failed: %v", err)
	}

	skill, matches, err := ResolveSkill(reg, "command injection")
	if err != nil {
		t.Fatalf("ResolveSkill() failed: %v (matches=%v)", err, matches)
	}
	if skill == nil {
		t.Fatal("expected a resolved skill")
	}
	if skill.Name != "vulnerabilities/command_injection" {
		t.Fatalf("expected command injection skill, got %q", skill.Name)
	}
}

func TestBuildRelevantSkillsContextIncludesExactSkillNames(t *testing.T) {
	reg, err := NewEmbeddedRegistry()
	if err != nil {
		t.Fatalf("NewEmbeddedRegistry() failed: %v", err)
	}

	context := BuildRelevantSkillsContext(reg, "nextjs performance and routing", 3)
	if strings.TrimSpace(context) == "" {
		t.Fatal("expected non-empty skills context")
	}
	if !strings.Contains(context, "`frameworks/nextjs`") {
		t.Fatalf("expected nextjs skill hint, got: %s", context)
	}
}
