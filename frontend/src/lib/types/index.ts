// CodeCrashReview - Shared TypeScript Types

// ============================================
// Core Domain Types
// ============================================

export type Language =
	| 'cpp'
	| 'python'
	| 'java'
	| 'javascript'
	| 'typescript'
	| 'rust'
	| 'shell'
	| 'html'
	| 'gherkin';

export type AgentPersona =
	| 'sentinel'
	| 'optimizer'
	| 'refactor'
	| 'guardian'
	| 'docubot'
	| 'polyglot';

export type AgentStatus =
	| 'idle'
	| 'thinking'
	| 'editing'
	| 'testing'
	| 'submitting'
	| 'waiting'
	| 'error'
	| 'completed';

export type MatchStatus =
	| 'pending'
	| 'matching'
	| 'preparing'
	| 'active'
	| 'paused'
	| 'completed'
	| 'cancelled'
	| 'error';

export type MatchMode = 'ranked' | 'private' | 'tournament' | 'practice';

export type ChallengeSeverity = 'low' | 'medium' | 'high' | 'critical';

export type ChallengeCategory =
	| 'security'
	| 'performance'
	| 'refactor'
	| 'testing'
	| 'documentation'
	| 'migration'
	| 'bugfix'
	| 'feature';

export type SubmissionStatus =
	| 'pending'
	| 'running'
	| 'passed'
	| 'failed'
	| 'timeout'
	| 'error';

// ============================================
// User & Auth Types
// ============================================

export interface User {
	id: string;
	username: string;
	email: string;
	avatar_url?: string;
	rating: number;
	rating_deviation: number; // Glicko-2 RD
	rating_volatility: number; // Glicko-2 sigma
	peak_rating: number;
	total_matches: number;
	wins: number;
	losses: number;
	draws: number;
	preferred_agents: AgentPersona[];
	preferred_languages: Language[];
	created_at: string;
	updated_at: string;
}

export interface AuthTokens {
	access_token: string;
	refresh_token: string;
	expires_in: number;
	token_type: 'Bearer';
}

export interface LoginRequest {
	email: string;
	password: string;
}

export interface RegisterRequest {
	username: string;
	email: string;
	password: string;
}

export interface JWTPayload {
	sub: string; // user_id
	username: string;
	iat: number;
	exp: number;
	type: 'access' | 'refresh';
}

// ============================================
// Agent Types
// ============================================

export interface AgentConfig {
	id: AgentPersona;
	name: string;
	description: string;
	specialty: string;
	model: string;
	temperature: number;
	tools: string[];
	system_prompt: string;
	color: string;
	glow_color: string;
	icon: string;
}

export interface AgentState {
	persona: AgentPersona;
	status: AgentStatus;
	current_file?: string;
	current_line?: number;
	progress: number; // 0-100
	tokens_used: number;
	latency_ms: number;
	last_action: string;
	thoughts: string[];
	proposed_changes: CodeChange[];
	test_results?: TestResult;
	error?: string;
}

export interface AgentAction {
	type: 'think' | 'edit' | 'test' | 'submit' | 'wait';
	payload: unknown;
	timestamp: string;
	duration_ms?: number;
}

// ============================================
// Code & Sandbox Types
// ============================================

export interface CodeFile {
	path: string;
	language: Language;
	content: string;
	original_content?: string;
	encoding: 'utf-8';
	size_bytes: number;
	lines: number;
}

export interface CodeChange {
	id: string;
	file_path: string;
	language: Language;
	original_content: string;
	new_content: string;
	diff: DiffHunk[];
	description: string;
	agent_persona: AgentPersona;
	turn: number;
	timestamp: string;
	accepted: boolean | null;
}

export interface DiffHunk {
	old_start: number;
	old_lines: number;
	new_start: number;
	new_lines: number;
	lines: DiffLine[];
}

export interface DiffLine {
	type: 'context' | 'add' | 'remove';
	content: string;
	line_number?: number;
}

export interface SandboxConfig {
	language: Language;
	image: string;
	cpu_limit: number;
	memory_limit: string;
	pids_limit: number;
	network_disabled: boolean;
	readonly_rootfs: boolean;
	seccomp_profile: string;
	timeout_seconds: number;
	env_vars: Record<string, string>;
	working_dir: string;
}

export interface ExecutionRequest {
	files: CodeFile[];
	command: string;
	args: string[];
	stdin?: string;
	timeout_seconds: number;
	resource_limits: {
		cpu_seconds: number;
		memory_bytes: number;
		output_bytes: number;
	};
}

export interface ExecutionResult {
	exit_code: number;
	stdout: string;
	stderr: string;
	duration_ms: number;
	memory_used_bytes: number;
	timed_out: boolean;
	oom_killed: boolean;
}

export interface TestResult {
	passed: number;
	failed: number;
	skipped: number;
	total: number;
	duration_ms: number
	suites: TestSuite[];
	coverage?: CoverageReport;
}

export interface TestSuite {
	name: string;
	tests: TestCase[];
	duration_ms: number;
}

export interface TestCase {
	name: string;
	status: 'passed' | 'failed' | 'skipped';
	duration_ms: number;
	error?: string;
	stdout?: string;
	stderr?: string;
}

export interface CoverageReport {
	lines: { total: number; covered: number; percentage: number };
	functions: { total: number; covered: number; percentage: number };
	branches: { total: number; covered: number; percentage: number };
	files: FileCoverage[];
}

export interface FileCoverage {
	path: string;
	lines: { total: number; covered: number; percentage: number };
}

// ============================================
// Match & Arena Types
// ============================================

export interface Match {
	id: string;
	mode: MatchMode;
	status: MatchStatus;
	challenge_id: string;
	challenge?: Challenge;
	participants: MatchParticipant[];
	spectators: string[];
	current_turn: number;
	max_turns: number;
	time_limit_seconds: number;
	time_remaining_seconds: number;
	started_at?: string;
	ended_at?: string;
	winner_id?: string;
	winner_persona?: AgentPersona;
	scoring: MatchScoring;
	replay_url?: string;
	created_at: string;
	updated_at: string;
}

export interface MatchParticipant {
	user_id: string;
	username: string;
	avatar_url?: string;
	agent_persona: AgentPersona;
	rating_before: number;
	rating_after?: number;
	rating_change?: number;
	score: number;
	submissions: Submission[];
	status: AgentStatus;
	connected: boolean;
}

export interface MatchScoring {
	criteria: ScoringCriterion[];
	weights: Record<string, number>;
	agent_scores: Record<AgentPersona, CriterionScore[]>;
	final_scores: Record<AgentPersona, number>;
}

export interface ScoringCriterion {
	id: string;
	name: string;
	description: string;
	weight: number;
	max_score: number;
	category: 'correctness' | 'security' | 'performance' | 'style' | 'testing' | 'innovation';
}

export interface CriterionScore {
	criterion_id: string;
	score: number;
	max_score: number;
	details: string;
	evidence: string[];
}

export interface Submission {
	id: string;
	match_id: string;
	participant_id: string;
	agent_persona: AgentPersona;
	turn: number;
	files: CodeFile[];
	changes: CodeChange[];
	test_result?: TestResult;
	execution_result?: ExecutionResult;
	score?: number;
	status: SubmissionStatus;
	submitted_at: string;
	evaluated_at?: string;
	feedback?: string;
}

// ============================================
// Challenge Types
// ============================================

export interface Challenge {
	id: string;
	title: string;
	description: string;
	repository_url: string;
	repository_branch: string;
	target_paths: string[];
	exclude_paths: string[];
	severity: ChallengeSeverity;
	category: ChallengeCategory;
	time_limit_seconds: number;
	max_turns: number
	acceptance_criteria: AcceptanceCriterion[];
	scoring_weights: Record<string, number>;
	known_issues: KnownIssue[];
	tags: string[];
	template_id?: string;
	is_public: boolean;
	created_by: string;
	created_at: string;
	updated_at: string;
	play_count: number;
	average_rating: number;
}

export interface AcceptanceCriterion {
	id: string;
	description: string;
	type: 'test' | 'static_analysis' | 'metric' | 'manual';
	weight: number;
	pass_threshold: number;
	config?: Record<string, unknown>;
}

export interface KnownIssue {
	file: string;
	line: number;
	column?: number;
	issue: string;
	severity: ChallengeSeverity;
	priority: 'low' | 'medium' | 'high' | 'critical';
	cwe_id?: string;
	description?: string;
}

export interface ChallengeTemplate {
	id: string;
	name: string;
	description: string;
	category: ChallengeCategory;
	default_severity: ChallengeSeverity;
	default_time_limit: number
	default_max_turns: number
	acceptance_criteria_template: AcceptanceCriterion[];
	scoring_weights_template: Record<string, number>;
	known_issues_template: KnownIssue[];
	example_repository?: string;
	created_at: string;
	updated_at: string;
}

export interface CreateChallengeRequest {
	title: string;
	description: string;
	repository_url: string;
	repository_branch?: string;
	target_paths: string[];
	exclude_paths?: string[];
	severity: ChallengeSeverity;
	category: ChallengeCategory;
	time_limit_seconds?: number;
	max_turns?: number;
	acceptance_criteria: AcceptanceCriterion[];
	scoring_weights?: Record<string, number>;
	known_issues?: KnownIssue[];
	tags?: string[];
	template_id?: string;
	is_public?: boolean;
}

// ============================================
// Snippet Types
// ============================================

export interface Snippet {
	id: string;
	title: string;
	description: string;
	language: Language;
	code: string;
	tags: string[];
	complexity: 'basic' | 'intermediate' | 'advanced';
	author: string;
	author_id?: string;
	source_url?: string;
	rating: number;
	usage_count: number;
	is_official: boolean;
	created_at: string;
	updated_at: string;
}

export interface SnippetCollection {
	id: string;
	name: string;
	description: string;
	snippet_ids: string[];
	tags: string[];
	is_curated: boolean;
	created_at: string;
}

// ============================================
// Leaderboard & Stats Types
// ============================================

export interface LeaderboardEntry {
	rank: number;
	user_id: string;
	username: string;
	avatar_url?: string;
	rating: number;
	rating_deviation: number;
	peak_rating: number;
	total_matches: number;
	wins: number;
	losses: number;
	win_rate: number;
	preferred_agent?: AgentPersona;
	preferred_language?: Language;
	last_active: string;
	trend: 'up' | 'down' | 'stable';
	rating_history: RatingPoint[];
}

export interface RatingPoint {
	timestamp: string;
	rating: number;
	rating_deviation: number;
	match_id: string;
	change: number;
}

export interface AgentStats {
	persona: AgentPersona;
	total_matches: number;
	wins: number;
	losses: number;
	draws: number;
	win_rate: number;
	average_score: number;
	average_turns: number;
	average_time_seconds: number
	favorite_language: Language;
	strongest_against: AgentPersona;
	weakest_against: AgentPersona;
	rating_distribution: Record<string, number>; // rating range -> count
}

export interface UserStats {
	user_id: string;
	overall: LeaderboardEntry;
	by_agent: Record<AgentPersona, AgentStats>;
	by_language: Record<Language, { matches: number; wins: number; win_rate: number }>;
	by_category: Record<ChallengeCategory, { matches: number; wins: number; win_rate: number }>;
	recent_matches: MatchSummary[];
	streaks: {
		current: number;
		longest_win: number;
		longest_loss: number;
	};
	achievements: Achievement[];
}

export interface MatchSummary {
	id: string;
	mode: MatchMode;
	challenge_title: string;
	opponent: string;
	opponent_agent: AgentPersona;
	my_agent: AgentPersona;
	result: 'win' | 'loss' | 'draw';
	rating_change: number;
	score: number;
	opponent_score: number;
	duration_seconds: number;
	played_at: string;
}

export interface Achievement {
	id: string;
	name: string;
	description: string;
	icon: string;
	color: string;
	unlocked_at?: string;
	progress?: number;
	max_progress?: number;
}

// ============================================
// WebSocket Message Types
// ============================================

export type WSMessageType =
	| 'auth'
	| 'auth_response'
	| 'join_match'
	| 'leave_match'
	| 'match_state'
	| 'match_update'
	| 'agent_action'
	| 'agent_status'
	| 'code_change'
	| 'submission'
	| 'submission_result'
	| 'test_result'
	| 'execution_result'
	| 'turn_change'
	| 'timer_update'
	| 'match_end'
	| 'leaderboard_update'
	| 'queue_status'
	| 'error'
	| 'ping'
	| 'pong';

export interface WSMessage<T = unknown> {
	type: WSMessageType;
	payload: T;
	timestamp: string;
	sequence: number;
}

export interface WSAuthPayload {
	token: string;
}

export interface WSAuthResponsePayload {
	success: boolean;
	user?: User;
	error?: string;
}

export interface WSJoinMatchPayload {
	match_id: string;
	as_spectator?: boolean;
}

export interface WSMatchStatePayload {
	match: Match;
	your_participant_id?: string;
}

export interface WSMatchUpdatePayload {
	match: Partial<Match>;
	participants?: Partial<MatchParticipant>[];
}

export interface WSAgentActionPayload {
	participant_id: string;
	agent_persona: AgentPersona;
	action: AgentAction;
}

export interface WSAgentStatusPayload {
	participant_id: string;
	agent_persona: AgentPersona;
	status: AgentStatus;
	progress: number;
	current_file?: string;
	current_line?: number;
	thoughts: string[];
}

export interface WSCodeChangePayload {
	participant_id: string;
	agent_persona: AgentPersona;
	change: CodeChange;
}

export interface WSSubmissionPayload {
	participant_id: string;
	agent_persona: AgentPersona;
	submission: Submission;
}

export interface WSSubmissionResultPayload {
	submission_id: string;
	score: number;
	criterion_scores: CriterionScore[];
	feedback: string;
}

export interface WSTestResultPayload {
	participant_id: string;
	agent_persona: AgentPersona;
	test_result: TestResult;
}

export interface WSExecutionResultPayload {
	participant_id: string;
	agent_persona: AgentPersona;
	result: ExecutionResult;
}

export interface WSTurnChangePayload {
	turn: number;
	active_participant_id: string;
	time_remaining_seconds: number;
}

export interface WSTimerUpdatePayload {
	time_remaining_seconds: number;
	warning: boolean;
}

export interface WSMatchEndPayload {
	winner_id: string | null;
	winner_persona: AgentPersona | null;
	final_scores: Record<AgentPersona, number>;
	rating_changes: Record<string, number>;
	summary: MatchSummary;
	replay_url: string;
}

export interface WSQueueStatusPayload {
	position: number;
	estimated_wait_seconds: number;
	players_in_queue: number;
}

export interface WSErrorPayload {
	code: string;
	message: string;
	details?: unknown;
}

export interface WSPingPayload {
	timestamp: string;
}

export interface WSPongPayload {
	client_timestamp: string;
	server_timestamp: string;
}

// ============================================
// API Request/Response Types
// ============================================

export interface ApiResponse<T = unknown> {
	success: boolean;
	data?: T;
	error?: ApiError;
	meta?: ResponseMeta;
}

export interface ApiError {
	code: string;
	message: string;
	details?: unknown;
	status_code: number;
}

export interface ResponseMeta {
	page?: number;
	per_page?: number;
	total?: number;
	total_pages?: number;
}

export interface PaginatedResponse<T> extends ApiResponse<T[]> {
	meta: ResponseMeta;
}

// ============================================
// UI State Types
// ============================================

export interface UIState {
	theme: 'dark' | 'light';
	sidebar_open: boolean;
	active_modal: string | null;
	notifications: Notification[];
	loading: Record<string, boolean>;
	toasts: Toast[];
}

export interface Notification {
	id: string;
	type: 'info' | 'success' | 'warning' | 'error';
	title: string;
	message?: string;
	action?: { label: string; href: string };
	read: boolean;
	created_at: string;
}

export interface Toast {
	id: string;
	type: 'info' | 'success' | 'warning' | 'error';
	title: string;
	message?: string;
	duration?: number;
	action?: { label: string; onClick: () => void };
}

export type ViewMode = 'dashboard' | 'arena' | 'challenges' | 'leaderboard' | 'snippets' | 'profile' | 'settings';

// ============================================
// Configuration Types
// ============================================

export interface AppConfig {
	api_url: string;
	ws_url: string;
	features: {
		spectator_mode: boolean;
		replay_system: boolean;
		challenge_templates: boolean;
		snippet_import: boolean;
		agent_analytics: boolean;
	};
	monaco: {
		theme: string;
		font_size: number;
		line_height: number;
		minimap: boolean;
		word_wrap: 'on' | 'off' | 'bounded';
	};
	arena: {
		auto_refresh_interval: number;
		max_thoughts_display: number;
		show_timeline: boolean;
	};
}

// ============================================
// Type Guards
// ============================================

export function isAgentPersona(value: string): value is AgentPersona {
	return ['sentinel', 'optimizer', 'refactor', 'guardian', 'docubot', 'polyglot'].includes(value);
}

export function isLanguage(value: string): value is Language {
	return ['cpp', 'python', 'java', 'javascript', 'typescript', 'rust', 'shell', 'html', 'gherkin'].includes(value);
}

export function isMatchStatus(value: string): value is MatchStatus {
	return ['pending', 'matching', 'preparing', 'active', 'paused', 'completed', 'cancelled', 'error'].includes(value);
}

export function getAgentColor(persona: AgentPersona): string {
	const colors: Record<AgentPersona, string> = {
		sentinel: '#ff3366',
		optimizer: '#00ff87',
		refactor: '#0099ff',
		guardian: '#ffaa00',
		docubot: '#ff00ff',
		polyglot: '#00ffff',
	};
	return colors[persona];
}

export function getLanguageExtension(language: Language): string {
	const extensions: Record<Language, string> = {
		cpp: 'cpp',
		python: 'py',
		java: 'java',
		javascript: 'js',
		typescript: 'ts',
		rust: 'rs',
		shell: 'sh',
		html: 'html',
		gherkin: 'feature',
	};
	return extensions[language];
}

export function getMonacoLanguageId(language: Language): string {
	const ids: Record<Language, string> = {
		cpp: 'cpp',
		python: 'python',
		java: 'java',
		javascript: 'javascript',
		typescript: 'typescript',
		rust: 'rust',
		shell: 'shell',
		html: 'html',
		gherkin: 'gherkin',
	};
	return ids[language];
}