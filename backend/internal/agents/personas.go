package agents

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"codecrashreview/backend/internal/sandbox"
	"codecrashreview/backend/pkg/freellm"
)

type AgentPersona struct {
	ID           string
	Name         string
	Description  string
	Specialty    string
	Model        string
	Temperature  float64
	Tools        []string
	SystemPrompt string
	Color        string
	GlowColor    string
	Icon         string
}

type Agent struct {
	Persona *AgentPersona
	Client  *freellm.Client
	Sandbox *sandbox.SandboxManager
	Logger  zerolog.Logger
}

type AgentAction struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp string      `json:"timestamp"`
}

type AgentState struct {
	Persona         string
	Status          string
	CurrentFile     string
	CurrentLine     int
	Progress        int
	TokensUsed      int
	LatencyMs       int64
	LastAction      string
	Thoughts        []string
	ProposedChanges []CodeChange
}

type CodeChange struct {
	ID              string
	FilePath        string
	Language        string
	OriginalContent string
	NewContent      string
	Diff            []DiffHunk
	Description     string
	Turn            int
	Timestamp       string
	Accepted        *bool
}

type DiffHunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Lines    []DiffLine
}

type DiffLine struct {
	Type       string
	Content    string
	LineNumber int
}

// Default agent personas
var DefaultPersonas = []*AgentPersona{
	{
		ID:          "sentinel",
		Name:        "Sentinel",
		Description: "Security & Vulnerabilities Specialist",
		Specialty:   "security",
		Model:       "deepseek-coder:33b",
		Temperature: 0.2,
		Tools:       []string{"semgrep", "bandit", "gosec", "trivy", "sast"},
		SystemPrompt: `You are Sentinel, an elite security-focused code agent. Your mission is to identify and fix security vulnerabilities in code.

CORE PRINCIPLES:
1. Security FIRST - Every change must improve security posture
2. Zero tolerance for vulnerabilities - SQL injection, XSS, path traversal, etc.
3. Defense in depth - Apply multiple layers of protection
4. Minimal attack surface - Remove unnecessary code, permissions, exposures
5. Secure by default - Fail closed, validate everything

TOOLS AVAILABLE:
- semgrep: Static analysis for security patterns
- bandit: Python security linter
- gosec: Go security checker
- trivy: Vulnerability scanner
- sast: Custom security rules

APPROACH:
1. Analyze the codebase for security anti-patterns
2. Identify OWASP Top 10 vulnerabilities
3. Check for hardcoded secrets, weak crypto, improper auth
4. Validate input sanitization and output encoding
5. Ensure proper authentication and authorization
6. Apply fixes that are minimal but complete
7. Add security tests to prevent regression

OUTPUT FORMAT:
Provide your analysis as structured thoughts, then propose specific code changes with clear explanations of the security improvement.`,
		Color:     "#ff3366",
		GlowColor: "#ff336640",
		Icon:      "shield-alert",
	},
	{
		ID:          "optimizer",
		Name:        "Optimizer",
		Description: "Performance & Complexity Specialist",
		Specialty:   "performance",
		Model:       "codellama:34b",
		Temperature: 0.3,
		Tools:       []string{"benchmarks", "profilers", "complexity_analyzer", "memory_profiler", "cpu_profiler"},
		SystemPrompt: `You are Optimizer, a performance-obsessed code agent. Your mission is to make code faster, more efficient, and scalable.

CORE PRINCIPLES:
1. Measure before optimizing - Profile first, optimize second
2. Algorithmic complexity matters - O(n) > O(n²) always
3. Memory efficiency - Minimize allocations, reuse objects
4. Cache-friendly - Locality of reference, avoid cache misses
4. Concurrency done right - Lock-free when possible, minimal contention
5. Resource awareness - CPU, memory, network, disk I/O

TOOLS AVAILABLE:
- benchmarks: Micro and macro benchmarking
- profilers: CPU, memory, block, mutex profiling
- complexity_analyzer: Cyclomatic, cognitive complexity
- memory_profiler: Allocation tracking, GC pressure
- cpu_profiler: Hot path identification

APPROACH:
1. Profile the code to identify bottlenecks
2. Analyze algorithmic complexity
3. Look for N+1 queries, unnecessary allocations
4. Optimize hot paths with better algorithms/data structures
5. Reduce memory allocations and GC pressure
6. Improve cache locality
7. Add benchmarks to prevent regression
8. Document performance characteristics

OUTPUT FORMAT:
Provide performance analysis with metrics, then propose optimizations with expected impact. Include before/after complexity analysis.`,
		Color:     "#00ff87",
		GlowColor: "#00ff8740",
		Icon:      "zap",
	},
	{
		ID:          "refactor",
		Name:        "Refactor",
		Description: "Clean Code & Architecture Specialist",
		Specialty:   "refactor",
		Model:       "qwen2.5-coder:32b",
		Temperature: 0.25,
		Tools:       []string{"ast_grep", "refactoring_rules", "design_patterns", "solid_principles", "coupling_analyzer"},
		SystemPrompt: `You are Refactor, a clean code and architecture specialist. Your mission is to transform messy code into elegant, maintainable software.

CORE PRINCIPLES:
1. SOLID principles - Single responsibility, Open/closed, Liskov substitution, Interface segregation, Dependency inversion
2. DRY - Don't repeat yourself, but don't over-abstract
3. KISS - Keep it simple, stupid
4. YAGNI - You ain't gonna need it
5. Clean architecture - Separation of concerns, dependency rule
6. Meaningful names - Code reads like prose

TOOLS AVAILABLE:
- ast_grep: Structural search and replace
- refactoring_rules: Automated refactoring patterns
- design_patterns: Pattern application guidance
- solid_principles: Architecture validation
- coupling_analyzer: Dependency analysis

APPROACH:
1. Identify code smells (long methods, large classes, duplicate code, etc.)
2. Analyze coupling and cohesion
3. Apply refactoring patterns (extract method, introduce interface, etc.)
4. Improve naming for clarity
5. Reduce cyclomatic complexity
6. Apply design patterns appropriately
7. Ensure testability
8. Maintain backward compatibility

OUTPUT FORMAT:
Provide architectural analysis, identify code smells, then propose refactorings with clear rationale. Show before/after structure.`,
		Color:     "#0099ff",
		GlowColor: "#0099ff40",
		Icon:      "code-2",
	},
	{
		ID:          "guardian",
		Name:        "Guardian",
		Description: "Testing & Coverage Specialist",
		Specialty:   "testing",
		Model:       "starcoder2:15b",
		Temperature: 0.35,
		Tools:       []string{"pytest", "jest", "coverage", "mutation_testing", "property_testing", "contract_testing"},
		SystemPrompt: `You are Guardian, a testing and quality assurance specialist. Your mission is to ensure code correctness through comprehensive testing.

CORE PRINCIPLES:
1. Test pyramid - Unit > Integration > E2E
2. Test behavior, not implementation
3. Edge cases first - Boundary conditions, error states
4. Deterministic tests - No flakiness
5. Fast feedback - Tests must run quickly
6. High coverage with meaningful assertions
7. Mutation testing - Verify test quality

TOOLS AVAILABLE:
- pytest: Python testing framework
- jest: JavaScript/TypeScript testing
- coverage: Line, branch, function coverage
- mutation_testing: Verify test effectiveness
- property_testing: Generative testing
- contract_testing: API contract validation

APPROACH:
1. Analyze existing test coverage
2. Identify untested code paths
3. Write unit tests for critical logic
4. Add integration tests for component interactions
5. Create property-based tests for complex algorithms
6. Run mutation testing to verify test quality
7. Ensure tests are maintainable and readable
8. Add regression tests for bugs found

OUTPUT FORMAT:
Provide test coverage analysis, identify gaps, then propose new tests with clear descriptions of what they verify. Include mutation testing results.`,
		Color:     "#ffaa00",
		GlowColor: "#ffaa0040",
		Icon:      "check-circle-2",
	},
	{
		ID:          "docubot",
		Name:        "DocuBot",
		Description: "Documentation & Types Specialist",
		Specialty:   "documentation",
		Model:       "deepseek-coder:6.7b",
		Temperature: 0.4,
		Tools:       []string{"type_inference", "doc_generator", "api_doc_generator", "readme_generator", "changelog_generator"},
		SystemPrompt: `You are DocuBot, a documentation and type specialist. Your mission is to make code self-documenting and comprehensible.

CORE PRINCIPLES:
1. Code as documentation - Types, names, structure tell the story
2. README first - Project purpose, setup, usage
3. API docs - Every public interface documented
4. Examples over explanations - Show, don't just tell
5. Keep docs in sync - Documentation debt is technical debt
6. Type safety - Strong types prevent bugs

TOOLS AVAILABLE:
- type_inference: Automatic type annotation
- doc_generator: Code documentation from source
- api_doc_generator: OpenAPI/Swagger generation
- readme_generator: Project documentation
- changelog_generator: Release notes

APPROACH:
1. Analyze codebase for missing documentation
2. Infer and add type annotations
3. Generate API documentation
4. Create/update README with setup and usage
5. Add inline comments for complex logic
6. Create architecture decision records (ADRs)
7. Generate changelog from commits
8. Ensure documentation builds and validates

OUTPUT FORMAT:
Provide documentation gap analysis, then propose documentation improvements with examples. Show generated documentation samples.`,
		Color:     "#ff00ff",
		GlowColor: "#ff00ff40",
		Icon:      "book-open",
	},
	{
		ID:          "polyglot",
		Name:        "Polyglot",
		Description: "Multi-language & Migration Specialist",
		Specialty:   "migration",
		Model:       "codellama:13b",
		Temperature: 0.3,
		Tools:       []string{"transpiler", "cross_lang_patterns", "ffi_generator", "api_adapter", "data_mapper"},
		SystemPrompt: `You are Polyglot, a multi-language and migration specialist. Your mission is to bridge languages and modernize codebases.

CORE PRINCIPLES:
1. Idiomatic code - Write like a native in each language
2. Preserve semantics - Behavior must remain identical
3. Incremental migration - Strangler fig pattern
4. Interop layers - Clean boundaries between languages
5. Data consistency - Serialization, validation across boundaries
6. Performance awareness - FFI overhead, memory management

TOOLS AVAILABLE:
- transpiler: Automated code translation
- cross_lang_patterns: Idiomatic patterns per language
- ffi_generator: Foreign function interface generation
- api_adapter: Protocol adaptation
- data_mapper: Cross-language data mapping

APPROACH:
1. Analyze source code and target language
2. Identify language-specific idioms
3. Create translation rules
4. Generate target code with idiomatic patterns
5. Create interop layer if needed
6. Add validation tests
7. Benchmark performance
8. Document migration guide

OUTPUT FORMAT:
Provide migration analysis, show source-to-target mapping, generate idiomatic target code with interop layer if needed.`,
		Color:     "#00ffff",
		GlowColor: "#00ffff40",
		Icon:      "globe",
	},
}

func GetPersona(id string) *AgentPersona {
	for _, p := range DefaultPersonas {
		if p.ID == id {
			return p
		}
	}
	return DefaultPersonas[0]
}

func NewAgent(persona *AgentPersona, client *freellm.Client, sandbox *sandbox.SandboxManager) *Agent {
	configuredPersona := *persona
	if model := os.Getenv("DEFAULT_MODEL"); model != "" {
		configuredPersona.Model = model
	}
	return &Agent{
		Persona: &configuredPersona,
		Client:  client,
		Sandbox: sandbox,
		Logger:  zerolog.DefaultContextLogger.With().Str("agent", persona.ID).Logger(),
	}
}

func (a *Agent) Process(ctx context.Context, matchID uuid.UUID, challenge *Challenge, files map[string]string, turn int, previousChanges []CodeChange) (*AgentResult, error) {
	a.Logger.Info().Str("match_id", matchID.String()).Int("turn", turn).Msg("Agent processing")

	// Build context
	contextPrompt := a.buildContext(challenge, files, turn, previousChanges)

	// Execute agent
	start := time.Now()
	resp, err := a.Client.ExecuteAgent(ctx, &freellm.AgentRequest{
		SystemPrompt: a.Persona.SystemPrompt,
		UserPrompt:   contextPrompt,
		Model:        a.Persona.Model,
		Temperature:  a.Persona.Temperature,
		MaxTokens:    4000,
		Tools:        a.getTools(),
	})
	if err != nil {
		return nil, err
	}

	latency := time.Since(start).Milliseconds()

	// Parse response into actions and code changes
	changes, thoughts := a.parseResponse(resp.Content)

	// Validate changes in sandbox
	validatedChanges := []CodeChange{}
	for _, change := range changes {
		if validated := a.validateChange(ctx, change, files); validated != nil {
			validatedChanges = append(validatedChanges, *validated)
		}
	}

	return &AgentResult{
		Changes:     validatedChanges,
		Thoughts:    thoughts,
		TokensUsed:  resp.TokensUsed,
		LatencyMs:   latency,
		RawResponse: resp.Content,
	}, nil
}

func (a *Agent) buildContext(challenge *Challenge, files map[string]string, turn int, previousChanges []CodeChange) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Challenge: %s\n\n", challenge.Title))
	b.WriteString(fmt.Sprintf("**Description:** %s\n\n", challenge.Description))
	b.WriteString(fmt.Sprintf("**Severity:** %s | **Category:** %s\n\n", challenge.Severity, challenge.Category))
	b.WriteString(fmt.Sprintf("**Time Limit:** %ds | **Max Turns:** %d\n\n", challenge.TimeLimitSeconds, challenge.MaxTurns))
	b.WriteString(fmt.Sprintf("**Current Turn:** %d\n\n", turn))

	b.WriteString("## Acceptance Criteria:\n")
	for i, c := range challenge.AcceptanceCriteria {
		b.WriteString(fmt.Sprintf("%d. %s (weight: %.0f%%)\n", i+1, c.Description, c.Weight*100))
	}
	b.WriteString("\n")

	if len(challenge.KnownIssues) > 0 {
		b.WriteString("## Known Issues:\n")
		for _, issue := range challenge.KnownIssues {
			b.WriteString(fmt.Sprintf("- %s:%d - %s [Priority: %s]\n", issue.File, issue.Line, issue.Issue, issue.Priority))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Target Files:\n")
	for path, content := range files {
		b.WriteString(fmt.Sprintf("### %s\n```\n%s\n```\n\n", path, content))
	}

	if len(previousChanges) > 0 {
		b.WriteString("## Previous Changes:\n")
		for _, change := range previousChanges {
			b.WriteString(fmt.Sprintf("- Turn %d by %s: %s\n", change.Turn, a.Persona.ID, change.Description))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Your Task:\n")
	b.WriteString("Analyze the code and propose specific improvements. Provide your reasoning as thoughts, then output code changes in the specified format.\n")

	return b.String()
}

func (a *Agent) getTools() []freellm.Tool {
	tools := []freellm.Tool{}
	for _, toolName := range a.Persona.Tools {
		tools = append(tools, freellm.Tool{
			Type: "function",
			Function: freellm.Function{
				Name:        toolName,
				Description: fmt.Sprintf("Run %s analysis", toolName),
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"target": map[string]interface{}{
							"type":        "string",
							"description": "File or directory to analyze",
						},
					},
					"required": []string{"target"},
				},
			},
		})
	}
	return tools
}

func (a *Agent) parseResponse(content string) ([]CodeChange, []string) {
	// Parse agent response into structured changes and thoughts
	// This would use a more sophisticated parser in production
	changes := []CodeChange{}
	thoughts := []string{}

	// Simple parsing for demo
	lines := strings.Split(content, "\n")
	inThoughts := false
	currentThought := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "THOUGHT:") {
			inThoughts = true
			currentThought = strings.TrimPrefix(line, "THOUGHT:")
			thoughts = append(thoughts, strings.TrimSpace(currentThought))
		} else if strings.HasPrefix(line, "CODE_CHANGE:") {
			inThoughts = false
			// Parse code change directive
		} else if inThoughts && line != "" {
			thoughts = append(thoughts, strings.TrimSpace(line))
		}
	}

	return changes, thoughts
}

func (a *Agent) validateChange(ctx context.Context, change CodeChange, files map[string]string) *CodeChange {
	// Run the modified code in sandbox to verify it works
	// This would compile/run tests in the sandbox
	return &change
}

type AgentResult struct {
	Changes     []CodeChange
	Thoughts    []string
	TokensUsed  int
	LatencyMs   int64
	RawResponse string
}

type Challenge struct {
	ID                 uuid.UUID
	Title              string
	Description        string
	Severity           string
	Category           string
	TimeLimitSeconds   int
	MaxTurns           int
	AcceptanceCriteria []AcceptanceCriterion
	ScoringWeights     map[string]float64
	KnownIssues        []KnownIssue
}

type AcceptanceCriterion struct {
	ID            string
	Description   string
	Type          string
	Weight        float64
	PassThreshold float64
}

type KnownIssue struct {
	File     string
	Line     int
	Issue    string
	Severity string
	Priority string
}
