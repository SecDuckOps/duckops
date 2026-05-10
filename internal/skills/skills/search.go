package skills

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/SecDuckOps/shared/types"
)

var (
	defaultRegistryOnce sync.Once
	defaultRegistry     Registry
	defaultRegistryErr  error
)

// DefaultRegistry returns a cached embedded registry instance for prompt retrieval.
func DefaultRegistry() (Registry, error) {
	defaultRegistryOnce.Do(func() {
		defaultRegistry, defaultRegistryErr = NewEmbeddedRegistry()
	})
	return defaultRegistry, defaultRegistryErr
}

// SearchSkills returns the most relevant embedded skills for the given query.
func (r *embeddedRegistry) SearchSkills(query string, limit int) []Skill {
	query = strings.TrimSpace(query)
	if limit <= 0 {
		limit = 3
	}
	if query == "" {
		skills := r.ListSkills()
		if len(skills) > limit {
			skills = skills[:limit]
		}
		return skills
	}

	queryLower := strings.ToLower(query)
	queryTokens := tokenize(query)
	type scoredSkill struct {
		skill Skill
		score int
	}

	scored := make([]scoredSkill, 0, len(r.skills))
	for _, skill := range r.skills {
		score := skillSearchScore(skill, queryLower, queryTokens)
		if score <= 0 {
			continue
		}
		scored = append(scored, scoredSkill{skill: skill, score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].skill.Name < scored[j].skill.Name
		}
		return scored[i].score > scored[j].score
	})

	if len(scored) > limit {
		scored = scored[:limit]
	}

	result := make([]Skill, 0, len(scored))
	for _, match := range scored {
		result = append(result, match.skill)
	}
	return result
}

// ResolveSkill finds the best-matching skill. Exact matches win; otherwise a single
// strong fuzzy match is accepted automatically.
func ResolveSkill(reg Registry, query string) (*Skill, []Skill, error) {
	reg = withDefaultRegistry(reg)
	if reg == nil {
		return nil, nil, types.New(types.ErrCodeInternal, "skill registry is unavailable")
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil, types.New(types.ErrCodeInvalidInput, "skill name is required")
	}

	if skill, err := reg.GetSkill(query); err == nil {
		return skill, nil, nil
	}

	matches := reg.SearchSkills(query, 5)
	if len(matches) == 1 {
		skill, err := reg.GetSkill(matches[0].Name)
		return skill, matches, err
	}

	normalized := normalizeSkillKey(query)
	if len(matches) > 0 {
		best := matches[0]
		bestKey := normalizeSkillKey(best.Name)
		if bestKey == normalized || strings.HasPrefix(bestKey, normalized) || strings.HasSuffix(bestKey, normalized) {
			skill, err := reg.GetSkill(best.Name)
			return skill, matches, err
		}
	}

	return nil, matches, types.Newf(types.ErrCodeNotFound, "skill %q not found", query)
}

// BuildRelevantSkillsContext renders a small static-RAG section for system prompts.
func BuildRelevantSkillsContext(reg Registry, query string, limit int) string {
	reg = withDefaultRegistry(reg)
	if reg == nil {
		return ""
	}

	matches := reg.SearchSkills(query, limit)
	if len(matches) == 0 {
		return ""
	}

	lines := []string{
		"The following embedded skills appear relevant. Prefer using `load_skill` with one of these exact names before specialized work:",
	}
	for _, skill := range matches {
		line := fmt.Sprintf("- `%s`: %s", skill.Name, strings.TrimSpace(skill.Description))
		if excerpt := skillExcerpt(skill.Content); excerpt != "" {
			line += fmt.Sprintf(" Key context: %s", excerpt)
		}
		lines = append(lines, line)
	}
	return "=== RELEVANT EMBEDDED SKILLS ===\n" + strings.Join(lines, "\n")
}

func withDefaultRegistry(reg Registry) Registry {
	if reg != nil {
		return reg
	}
	fallback, err := DefaultRegistry()
	if err != nil {
		return nil
	}
	return fallback
}

func skillSearchScore(skill Skill, queryLower string, queryTokens []string) int {
	name := strings.ToLower(skill.Name)
	desc := strings.ToLower(skill.Description)
	content := strings.ToLower(skill.Content)

	score := 0
	if strings.Contains(name, queryLower) {
		score += 30
	}
	if strings.Contains(desc, queryLower) {
		score += 18
	}
	if strings.Contains(content, queryLower) {
		score += 8
	}

	nameSegments := tokenize(strings.ReplaceAll(skill.Name, "/", " "))
	for _, token := range queryTokens {
		if token == "" {
			continue
		}
		if containsToken(nameSegments, token) {
			score += 10
		}
		if strings.Contains(desc, token) {
			score += 5
		}
		if strings.Contains(content, token) {
			score += 1
		}
	}

	return score
}

func containsToken(tokens []string, target string) bool {
	for _, token := range tokens {
		if token == target {
			return true
		}
	}
	return false
}

func tokenize(value string) []string {
	var b strings.Builder
	for _, r := range strings.ToLower(value) {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			b.WriteRune(r)
		default:
			b.WriteByte(' ')
		}
	}
	return strings.Fields(b.String())
}

func normalizeSkillKey(value string) string {
	return strings.Join(tokenize(value), "/")
}

func skillExcerpt(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(line) > 180 {
			return line[:177] + "..."
		}
		return line
	}
	return ""
}
