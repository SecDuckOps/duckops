// Package skills implements the Agent Skills open standard.
// See https://agentskills.io for the specification.
package skills

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/SecDuckOps/duckops/internal/home"
	"github.com/SecDuckOps/duckops/internal/pubsub"
	"github.com/charlievieth/fastwalk"
	"gopkg.in/yaml.v3"
)

const (
	SkillFileName          = "SKILL.md"
	MaxNameLength          = 64
	MaxDescriptionLength   = 1024
	MaxCompatibilityLength = 500
)

var (
	namePattern    = regexp.MustCompile(`^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$`)
	promptReplacer = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&apos;")
)

// Skill represents a parsed SKILL.md file.
type Skill struct {
	Name          string            `yaml:"name" json:"name"`
	Description   string            `yaml:"description" json:"description"`
	License       string            `yaml:"license,omitempty" json:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	UserInvocable bool              `yaml:"user-invocable" json:"user_invocable"`
	Label         string            `yaml:"label,omitempty" json:"label,omitempty"`
	Instructions  string            `yaml:"-" json:"instructions"`
	Path          string            `yaml:"-" json:"path"`
	SkillFilePath string            `yaml:"-" json:"skill_file_path"`
	Builtin       bool              `yaml:"-" json:"builtin"`
}

// DiscoveryState represents the outcome of discovering a single skill file.
type DiscoveryState int

const (
	// StateNormal indicates the skill was parsed and validated successfully.
	StateNormal DiscoveryState = iota
	// StateError indicates discovery encountered a scan/parse/validate error.
	StateError
)

// SkillState represents the latest discovery status of a skill file.
type SkillState struct {
	Name  string
	Path  string
	State DiscoveryState
	Err   error
}

// Event is published when skill discovery completes.
type Event struct {
	States []*SkillState
}

var broker = pubsub.NewBroker[Event]()

// SubscribeEvents returns a channel that receives events when skill discovery state changes.
func SubscribeEvents(ctx context.Context) <-chan pubsub.Event[Event] {
	return broker.Subscribe(ctx)
}

var (
	latestStatesMu sync.RWMutex
	latestStates   []*SkillState
)

// GetLatestStates returns the package-wide latest skill discovery states.
func GetLatestStates() []*SkillState {
	latestStatesMu.RLock()
	defer latestStatesMu.RUnlock()
	return append([]*SkillState(nil), latestStates...)
}

// SetLatestStates updates the package-wide latest skill discovery states.
func SetLatestStates(states []*SkillState) {
	latestStatesMu.Lock()
	latestStates = append([]*SkillState(nil), states...)
	latestStatesMu.Unlock()
}

// Validate checks if the skill meets spec requirements.
func (s *Skill) Validate() error {
	var errs []error

	if s.Name == "" {
		errs = append(errs, errors.New("name is required"))
	} else {
		if len(s.Name) > MaxNameLength {
			errs = append(errs, fmt.Errorf("name exceeds %d characters", MaxNameLength))
		}
		if !namePattern.MatchString(s.Name) {
			errs = append(errs, errors.New("name must be alphanumeric with hyphens, no leading/trailing/consecutive hyphens"))
		}
		if s.Path != "" && !strings.EqualFold(filepath.Base(s.Path), s.Name) {
			errs = append(errs, fmt.Errorf("name %q must match directory %q", s.Name, filepath.Base(s.Path)))
		}
	}

	if s.Description == "" {
		errs = append(errs, errors.New("description is required"))
	} else if len(s.Description) > MaxDescriptionLength {
		errs = append(errs, fmt.Errorf("description exceeds %d characters", MaxDescriptionLength))
	}

	if len(s.Compatibility) > MaxCompatibilityLength {
		errs = append(errs, fmt.Errorf("compatibility exceeds %d characters", MaxCompatibilityLength))
	}

	return errors.Join(errs...)
}

// Parse parses a SKILL.md file from disk.
func Parse(path string) (*Skill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	skill, err := ParseContent(content)
	if err != nil {
		return nil, err
	}

	skill.Path = filepath.Dir(path)
	skill.SkillFilePath = path

	return skill, nil
}

// ParseContent parses a SKILL.md from raw bytes.
func ParseContent(content []byte) (*Skill, error) {
	frontmatter, body, err := splitFrontmatter(string(content))
	if err != nil {
		return nil, err
	}

	var skill Skill
	if err := yaml.Unmarshal([]byte(frontmatter), &skill); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	skill.Instructions = strings.TrimSpace(body)

	return &skill, nil
}

// splitFrontmatter extracts YAML frontmatter and body from markdown content.
func splitFrontmatter(content string) (frontmatter, body string, err error) {
	// Strip UTF-8 BOM for compatibility with editors that include it.
	content = strings.TrimPrefix(content, "\uFEFF")
	// Normalize line endings to \n for consistent parsing.
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	lines := strings.Split(content, "\n")
	start := slices.IndexFunc(lines, func(line string) bool {
		return strings.TrimSpace(line) != ""
	})
	if start == -1 || strings.TrimSpace(lines[start]) != "---" {
		return "", "", errors.New("no YAML frontmatter found")
	}

	endOffset := slices.IndexFunc(lines[start+1:], func(line string) bool {
		return strings.TrimSpace(line) == "---"
	})
	if endOffset == -1 {
		return "", "", errors.New("unclosed frontmatter")
	}
	end := start + 1 + endOffset

	frontmatter = strings.Join(lines[start+1:end], "\n")
	body = strings.Join(lines[end+1:], "\n")
	return frontmatter, body, nil
}

// Discover finds all valid skills in the given paths.
func Discover(paths []string) []*Skill {
	skills, _ := DiscoverWithStates(paths)
	return skills
}

// DiscoverWithStates finds all valid skills in the given paths and also
// returns a per-file state slice describing parse/validation outcomes. Useful
// for diagnostics and UI reporting.
func DiscoverWithStates(paths []string) ([]*Skill, []*SkillState) {
	var skills []*Skill
	var states []*SkillState
	var mu sync.Mutex
	seen := make(map[string]bool)
	addState := func(name, path string, state DiscoveryState, err error) {
		mu.Lock()
		states = append(states, &SkillState{
			Name:  name,
			Path:  path,
			State: state,
			Err:   err,
		})
		mu.Unlock()
	}

	for _, base := range paths {
		// We use fastwalk with Follow: true instead of filepath.WalkDir because
		// WalkDir doesn't follow symlinked directories at any depth—only entry
		// points. This ensures skills in symlinked subdirectories are discovered.
		// fastwalk is concurrent, so we protect shared state (seen, skills) with mu.
		conf := fastwalk.Config{
			Follow:  true,
			ToSlash: fastwalk.DefaultToSlash(),
		}
		err := fastwalk.Walk(&conf, base, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				slog.Warn("Failed to walk skills path entry", "base", base, "path", path, "error", err)
				addState("", path, StateError, err)
				return nil
			}
			if d.IsDir() || d.Name() != SkillFileName {
				return nil
			}
			mu.Lock()
			if seen[path] {
				mu.Unlock()
				return nil
			}
			seen[path] = true
			mu.Unlock()
			skill, err := Parse(path)
			if err != nil {
				slog.Warn("Failed to parse skill file", "path", path, "error", err)
				addState("", path, StateError, err)
				return nil
			}
			if err := skill.Validate(); err != nil {
				slog.Warn("Skill validation failed", "path", path, "error", err)
				addState(skill.Name, path, StateError, err)
				return nil
			}
			slog.Debug("Successfully loaded skill", "name", skill.Name, "path", path)
			mu.Lock()
			skills = append(skills, skill)
			mu.Unlock()
			addState(skill.Name, path, StateNormal, nil)
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			slog.Warn("Failed to walk skills path", "path", base, "error", err)
		}
	}

	// fastwalk traversal order is non-deterministic, so sort for stable output.
	sort.SliceStable(skills, func(i, j int) bool {
		left := strings.ToLower(skills[i].SkillFilePath)
		right := strings.ToLower(skills[j].SkillFilePath)
		if left == right {
			return skills[i].SkillFilePath < skills[j].SkillFilePath
		}
		return left < right
	})

	broker.Publish(pubsub.UpdatedEvent, Event{States: states})
	return skills, states
}

// ToPromptXML generates XML for injection into the system prompt.
func ToPromptXML(skills []*Skill) string {
	if len(skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<available_skills>\n")
	for _, s := range skills {
		sb.WriteString("  <skill>\n")
		fmt.Fprintf(&sb, "    <name>%s</name>\n", escape(s.Name))
		fmt.Fprintf(&sb, "    <description>%s</description>\n", escape(s.Description))
		fmt.Fprintf(&sb, "    <location>%s</location>\n", escape(s.SkillFilePath))
		if s.Builtin {
			sb.WriteString("    <type>builtin</type>\n")
		}
		sb.WriteString("  </skill>\n")
	}
	sb.WriteString("</available_skills>")
	return sb.String()
}

func escape(s string) string {
	return promptReplacer.Replace(s)
}

// Deduplicate removes duplicate skills by name. When duplicates exist, the
// last occurrence wins. This means user skills (appended after builtins)
// override builtin skills with the same name.
func Deduplicate(all []*Skill) []*Skill {
	seen := make(map[string]int, len(all))
	for i, s := range all {
		seen[s.Name] = i
	}

	result := make([]*Skill, 0, len(seen))
	for i, s := range all {
		if seen[s.Name] == i {
			result = append(result, s)
		}
	}
	return result
}

// ApproxTokenCount returns a rough estimate of how many tokens a string
// occupies when sent to an LLM. Uses the common ~4-chars-per-token heuristic
// that approximates GPT/Claude tokenizers well enough for diagnostic logging.
func ApproxTokenCount(s string) int {
	if s == "" {
		return 0
	}
	return (len(s) + 3) / 4
}

// Filter removes skills whose names appear in the disabled list.
func Filter(all []*Skill, disabled []string) []*Skill {
	if len(disabled) == 0 {
		return all
	}

	disabledSet := make(map[string]bool, len(disabled))
	for _, name := range disabled {
		disabledSet[name] = true
	}

	result := make([]*Skill, 0, len(all))
	for _, s := range all {
		if !disabledSet[s.Name] {
			result = append(result, s)
		}
	}
	return result
}

// SourceType identifies where a skill comes from.
type SourceType string

const (
	SourceTypeUser    SourceType = "user"
	SourceTypeBuiltin SourceType = "builtin"
)

// SkillReadResult holds metadata about a skill returned alongside its
// content.
type SkillReadResult struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Source      SourceType `json:"source"`
	Builtin     bool       `json:"builtin"`
}

// CatalogEntry represents a skill entry in the skill catalog database/list.
type CatalogEntry struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Label         string     `json:"label"`
	Source        SourceType `json:"source"`
	UserInvocable bool       `json:"user_invocable"`
}

// Option configures the skills manager.
type Option func(*Manager)

// Manager holds the pre-discovered skills lists.
type Manager struct {
	mu            sync.RWMutex
	allSkills     []*Skill
	activeSkills  []*Skill
	states        []*SkillState
	resolvedPaths []string
	workingDir    string
	globalMirror  bool
}

// NewManager creates a new skills Manager with the given lists of skills.
func NewManager(all []*Skill, active []*Skill, states []*SkillState, opts ...Option) *Manager {
	m := &Manager{
		allSkills:    all,
		activeSkills: active,
		states:       append([]*SkillState(nil), states...),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}
	if m.globalMirror {
		SetLatestStates(m.states)
	}
	return m
}

// WithResolvedPaths configures the manager with the resolved skill discovery paths.
func WithResolvedPaths(paths []string) Option {
	return func(m *Manager) {
		m.resolvedPaths = append([]string(nil), paths...)
	}
}

// WithWorkingDir configures the manager with the workspace working directory.
func WithWorkingDir(workingDir string) Option {
	return func(m *Manager) {
		m.workingDir = workingDir
	}
}

// WithGlobalMirror keeps the package-level skills cache in sync with this manager.
func WithGlobalMirror() Option {
	return func(m *Manager) {
		m.globalMirror = true
	}
}

// AllSkills returns all pre-discovered skills.
func (m *Manager) AllSkills() []*Skill {
	return m.allSkills
}

// ActiveSkills returns all active skills.
func (m *Manager) ActiveSkills() []*Skill {
	return m.activeSkills
}

// ResolvedPaths returns the resolved skill discovery paths for the manager.
func (m *Manager) ResolvedPaths() []string {
	return append([]string(nil), m.resolvedPaths...)
}

// WorkingDir returns the manager's workspace working directory.
func (m *Manager) WorkingDir() string {
	return m.workingDir
}

// States returns the most recent skill discovery states for the manager.
func (m *Manager) States() []*SkillState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]*SkillState(nil), m.states...)
}

// SetLatestStates updates the manager state cache and packages it to
// the global shared mirror if configured.
func (m *Manager) SetLatestStates(states []*SkillState) {
	m.mu.Lock()
	m.states = append([]*SkillState(nil), states...)
	m.mu.Unlock()
	if m.globalMirror {
		SetLatestStates(states)
	}
}

// PublishStates updates the manager state cache and publishes an update
// event to subscribers.
func (m *Manager) PublishStates(states []*SkillState) {
	m.SetLatestStates(states)
	broker.Publish(pubsub.UpdatedEvent, Event{States: states})
}

// DiscoveryConfig holds options for the skills discovery process.
type DiscoveryConfig struct {
	SkillsPaths    []string
	DisabledSkills []string
	Resolver       func(string) (string, error)
	ResolvedPaths  []string
	WorkingDir     string
}

// ResolvedPaths returns the resolved skill discovery paths for this config.
func (cfg DiscoveryConfig) ResolvePaths() []string {
	if len(cfg.ResolvedPaths) > 0 {
		return append([]string(nil), cfg.ResolvedPaths...)
	}
	if len(cfg.SkillsPaths) == 0 {
		return nil
	}

	resolved := make([]string, 0, len(cfg.SkillsPaths))
	for _, pth := range cfg.SkillsPaths {
		expanded := home.Long(pth)
		if strings.HasPrefix(expanded, "$") && cfg.Resolver != nil {
			if resolvedPath, err := cfg.Resolver(expanded); err == nil {
				expanded = resolvedPath
			}
		}
		resolved = append(resolved, expanded)
	}
	return resolved
}

// Catalog converts active skills into a frontend-friendly catalog.
func Catalog(skills []*Skill, resolvedPaths []string, workingDir string) []CatalogEntry {
	entries := make([]CatalogEntry, 0, len(skills))
	for _, s := range skills {
		source := SourceTypeUser
		if s.Builtin {
			source = SourceTypeBuiltin
		}
		entries = append(entries, CatalogEntry{
			ID:            s.SkillFilePath,
			Name:          s.Name,
			Description:   s.Description,
			Label:         s.Label,
			Source:        source,
			UserInvocable: s.UserInvocable,
		})
	}
	return entries
}

// ReadContent reads a skill's SKILL.md content and returns metadata.
func ReadContent(activeSkills []*Skill, resolvedPaths []string, workingDir, skillID string) ([]byte, SkillReadResult, error) {
	if skillID == "" {
		return nil, SkillReadResult{}, errors.New("skill id is required")
	}

	var selected *Skill
	for _, s := range activeSkills {
		if s.SkillFilePath == skillID || s.Name == skillID {
			selected = s
			break
		}
	}

	if strings.HasPrefix(skillID, BuiltinPrefix) {
		builtinPath := strings.TrimPrefix(skillID, BuiltinPrefix)
		builtinPath = filepath.ToSlash(strings.TrimPrefix(builtinPath, "/"))
		data, err := BuiltinFS().ReadFile(filepath.ToSlash(filepath.Join("builtin", builtinPath)))
		if err != nil {
			return nil, SkillReadResult{}, err
		}
		result := SkillReadResult{Source: SourceTypeBuiltin, Builtin: true}
		if selected != nil {
			result.Name = selected.Name
			result.Description = selected.Description
		}
		return data, result, nil
	}

	if selected != nil {
		if selected.Builtin {
			return ReadContent(activeSkills, resolvedPaths, workingDir, selected.SkillFilePath)
		}
		data, err := os.ReadFile(selected.SkillFilePath)
		if err != nil {
			return nil, SkillReadResult{}, err
		}
		return data, SkillReadResult{
			Name:        selected.Name,
			Description: selected.Description,
			Source:      SourceTypeUser,
			Builtin:     false,
		}, nil
	}

	candidate := skillID
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(workingDir, candidate)
	}
	data, err := os.ReadFile(candidate)
	if err != nil {
		return nil, SkillReadResult{}, err
	}

	result := SkillReadResult{Source: SourceTypeUser, Builtin: false}
	if parsed, err := ParseContent(data); err == nil {
		result.Name = parsed.Name
		result.Description = parsed.Description
	}
	return data, result, nil
}

// DiscoverFromConfig performs skills discovery based on the provided configuration.
func DiscoverFromConfig(cfg DiscoveryConfig) (allSkills, activeSkills []*Skill, states []*SkillState) {
	builtin, builtinStates := DiscoverBuiltinWithStates()

	// Ensure builtin states path starts with "builtin/" for logDiscoveryStats
	for _, s := range builtinStates {
		if !strings.HasPrefix(s.Path, "builtin/") {
			// If it has BuiltinPrefix, replace it with "builtin/"
			if strings.HasPrefix(s.Path, BuiltinPrefix) {
				s.Path = "builtin/" + strings.TrimPrefix(s.Path, BuiltinPrefix)
			} else {
				s.Path = "builtin/" + s.Path
			}
		}
	}

	discovered := append([]*Skill(nil), builtin...)
	states = append([]*SkillState(nil), builtinStates...)

	var userStates []*SkillState
	var userPaths []string

	if len(cfg.SkillsPaths) > 0 {
		userPaths = make([]string, 0, len(cfg.SkillsPaths))
		for _, pth := range cfg.SkillsPaths {
			expanded := home.Long(pth)
			if strings.HasPrefix(expanded, "$") && cfg.Resolver != nil {
				if resolved, err := cfg.Resolver(expanded); err == nil {
					expanded = resolved
				}
			}
			userPaths = append(userPaths, expanded)
		}
		var userSkills []*Skill
		userSkills, userStates = DiscoverWithStates(userPaths)
		discovered = append(discovered, userSkills...)
		states = append(states, userStates...)
	}

	allSkills = Deduplicate(discovered)
	activeSkills = Filter(allSkills, cfg.DisabledSkills)
	return allSkills, activeSkills, states
}
