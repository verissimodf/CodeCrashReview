// CodeCrashReview - Svelte Stores
// Centralized state management using Svelte's built-in stores

import { writable, derived, get } from 'svelte/store';
import { browser } from '$app/environment';
import type {
	User,
	AuthTokens,
	Match,
	MatchParticipant,
	AgentState,
	AgentPersona,
	AgentStatus,
	CodeChange,
	Submission,
	TestResult,
	ExecutionResult,
	LeaderboardEntry,
	UserStats,
	Challenge,
	Snippet,
	Notification,
	Toast,
	UIState,
	WSMessage,
	WSMessageType,
	AppConfig,
	Language,
	ViewMode
} from '$lib/types';

// ============================================
// Auth Store
// ============================================

interface AuthState {
	user: User | null;
	tokens: AuthTokens | null;
	isAuthenticated: boolean;
	isLoading: boolean;
	error: string | null;
}

function createAuthStore() {
	const { subscribe, set, update } = writable<AuthState>({
		user: null,
		tokens: null,
		isAuthenticated: false,
		isLoading: true,
		error: null
	});

	// Initialize from localStorage on client
	if (browser) {
		const storedTokens = localStorage.getItem('ccr_tokens');
		const storedUser = localStorage.getItem('ccr_user');
		if (storedTokens && storedUser) {
			try {
				const tokens = JSON.parse(storedTokens);
				const user = JSON.parse(storedUser);
				// Check if token is expired
				const payload = JSON.parse(atob(tokens.access_token.split('.')[1]));
				if (payload.exp * 1000 > Date.now()) {
					set({
						user,
						tokens,
						isAuthenticated: true,
						isLoading: false,
						error: null
					});
				} else {
					// Token expired, try refresh
					refreshToken(tokens.refresh_token);
				}
			} catch {
				clearAuth();
			}
		} else {
			update(state => ({ ...state, isLoading: false }));
		}
	}

	async function refreshToken(refreshToken: string) {
		try {
			const response = await fetch(`${get(apiConfig).api_url}/auth/refresh`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ refresh_token: refreshToken })
			});
			if (response.ok) {
				const data = await response.json();
				setAuth(data.user, data.tokens);
			} else {
				clearAuth();
			}
		} catch {
			clearAuth();
		}
	}

	function setAuth(user: User, tokens: AuthTokens) {
		if (browser) {
			localStorage.setItem('ccr_user', JSON.stringify(user));
			localStorage.setItem('ccr_tokens', JSON.stringify(tokens));
		}
		set({
			user,
			tokens,
			isAuthenticated: true,
			isLoading: false,
			error: null
		});
	}

	function clearAuth() {
		if (browser) {
			localStorage.removeItem('ccr_user');
			localStorage.removeItem('ccr_tokens');
		}
		set({
			user: null,
			tokens: null,
			isAuthenticated: false,
			isLoading: false,
			error: null
		});
	}

	function setLoading(loading: boolean) {
		update(state => ({ ...state, isLoading: loading }));
	}

	function setError(error: string | null) {
		update(state => ({ ...state, error }));
	}

	function updateUser(updates: Partial<User>) {
		update(state => {
			if (!state.user) return state;
			const user = { ...state.user, ...updates };
			if (browser) localStorage.setItem('ccr_user', JSON.stringify(user));
			return { ...state, user };
		});
	}

	return {
		subscribe,
		setAuth,
		clearAuth,
		setLoading,
		setError,
		updateUser,
		refreshToken
	};
}

export const auth = createAuthStore();

// ============================================
// WebSocket Store
// ============================================

interface WSState {
	connected: boolean;
	connecting: boolean;
	socket: WebSocket | null;
	matchId: string | null;
	messageQueue: WSMessage[];
	lastMessage: WSMessage | null;
	error: string | null;
	reconnectAttempts: number;
	maxReconnectAttempts: number;
	reconnectDelay: number;
}

function createWSStore() {
	const { subscribe, set, update } = writable<WSState>({
		connected: false,
		connecting: false,
		socket: null,
		matchId: null,
		messageQueue: [],
		lastMessage: null,
		error: null,
		reconnectAttempts: 0,
		maxReconnectAttempts: 5,
		reconnectDelay: 1000
	});

	let pingInterval: ReturnType<typeof setInterval> | null = null;

	function connect(matchId?: string) {
		update(state => {
			if (state.connecting || state.connected) return state;
			return { ...state, connecting: true, error: null };
		});

		const config = get(apiConfig);
		const authState = get(auth);
		const token = authState.tokens?.access_token;

		if (!token) {
			update(state => ({ ...state, connecting: false, error: 'No auth token' }));
			return;
		}

		const wsUrl = `${config.ws_url}?token=${encodeURIComponent(token)}${matchId ? `&match=${matchId}` : ''}`;
		const socket = new WebSocket(wsUrl);

		socket.onopen = () => {
			update(state => ({
				...state,
				connected: true,
				connecting: false,
				socket,
				matchId: matchId || null,
				reconnectAttempts: 0,
				error: null
			}));
			startPing();
			// Flush queued messages
			update(state => {
				state.messageQueue.forEach(msg => send(msg));
				return { ...state, messageQueue: [] };
			});
		};

		socket.onclose = (event) => {
			stopPing();
			update(state => ({
				...state,
				connected: false,
				connecting: false,
				socket: null,
				error: event.wasClean ? null : 'Connection closed unexpectedly'
			}));
			attemptReconnect(matchId);
		};

		socket.onerror = (error) => {
			update(state => ({ ...state, error: 'WebSocket error' }));
		};

		socket.onmessage = (event) => {
			try {
				const message: WSMessage = JSON.parse(event.data);
				update(state => ({ ...state, lastMessage: message }));
				handleMessage(message);
			} catch (e) {
				console.error('Failed to parse WS message:', e);
			}
		};

		update(state => ({ ...state, socket }));
	}

	function disconnect() {
		update(state => {
			if (state.socket) {
				state.socket.close(1000, 'Client disconnect');
			}
			stopPing();
			return {
				...state,
				connected: false,
				connecting: false,
				socket: null,
				matchId: null
			};
		});
	}

	function attemptReconnect(matchId?: string) {
		update(state => {
			if (state.reconnectAttempts >= state.maxReconnectAttempts) {
				return { ...state, error: 'Max reconnect attempts reached' };
			}
			const delay = state.reconnectDelay * Math.pow(2, state.reconnectAttempts);
			setTimeout(() => connect(matchId), delay);
			return { ...state, reconnectAttempts: state.reconnectAttempts + 1 };
		});
	}

	function send(message: WSMessage) {
		update(state => {
			if (state.socket?.readyState === WebSocket.OPEN) {
				state.socket.send(JSON.stringify(message));
			} else {
				state.messageQueue.push(message);
			}
			return state;
		});
	}

	function startPing() {
		pingInterval = setInterval(() => {
			send({
				type: 'ping',
				payload: { timestamp: new Date().toISOString() },
				timestamp: new Date().toISOString(),
				sequence: Date.now()
			});
		}, 30000);
	}

	function stopPing() {
		if (pingInterval) {
			clearInterval(pingInterval);
			pingInterval = null;
		}
	}

	function handleMessage(message: WSMessage) {
		// Message handling is done via subscribers in components
		// This store just relays messages
	}

	return {
		subscribe,
		connect,
		disconnect,
		send,
		handleMessage
	};
}

export const ws = createWSStore();

// ============================================
// Match/Arena Store
// ============================================

interface MatchState {
	currentMatch: Match | null;
	participants: Map<string, MatchParticipant>;
	agentStates: Map<AgentPersona, AgentState>;
	codeChanges: CodeChange[];
	submissions: Submission[];
	currentTurn: number;
	timeRemaining: number;
	isSpectating: boolean;
	spectatorCount: number;
	replayData: unknown | null;
}

function createMatchStore() {
	const { subscribe, set, update } = writable<MatchState>({
		currentMatch: null,
		participants: new Map(),
		agentStates: new Map(),
		codeChanges: [],
		submissions: [],
		currentTurn: 0,
		timeRemaining: 0,
		isSpectating: false,
		spectatorCount: 0,
		replayData: null
	});

	function setMatch(match: Match, isSpectating = false) {
		const participants = new Map(match.participants.map(p => [p.user_id, p]));
		const agentStates = new Map<AgentPersona, AgentState>();
		match.participants.forEach(p => {
			agentStates.set(p.agent_persona, {
				persona: p.agent_persona,
				status: 'idle',
				progress: 0,
				tokens_used: 0,
				latency_ms: 0,
				last_action: '',
				thoughts: [],
				proposed_changes: []
			});
		});

		set({
			currentMatch: match,
			participants,
			agentStates,
			codeChanges: [],
			submissions: [],
			currentTurn: match.current_turn,
			timeRemaining: match.time_remaining_seconds,
			isSpectating,
			spectatorCount: match.spectators.length,
			replayData: null
		});
	}

	function updateMatch(updates: Partial<Match>) {
		update(state => {
			if (!state.currentMatch) return state;
			return { ...state, currentMatch: { ...state.currentMatch, ...updates } };
		});
	}

	function updateParticipant(userId: string, updates: Partial<MatchParticipant>) {
		update(state => {
			const participant = state.participants.get(userId);
			if (!participant) return state;
			const newParticipants = new Map(state.participants);
			newParticipants.set(userId, { ...participant, ...updates });
			return { ...state, participants: newParticipants };
		});
	}

	function updateAgentState(persona: AgentPersona, updates: Partial<AgentState>) {
		update(state => {
			const agentState = state.agentStates.get(persona);
			if (!agentState) return state;
			const newAgentStates = new Map(state.agentStates);
			newAgentStates.set(persona, { ...agentState, ...updates });
			return { ...state, agentStates: newAgentStates };
		});
	}

	function addCodeChange(change: CodeChange) {
		update(state => ({ ...state, codeChanges: [...state.codeChanges, change] }));
	}

	function addSubmission(submission: Submission) {
		update(state => ({ ...state, submissions: [...state.submissions, submission] }));
	}

	function setTurn(turn: number, timeRemaining: number) {
		update(state => ({ ...state, currentTurn: turn, timeRemaining }));
	}

	function updateTimer(seconds: number) {
		update(state => ({ ...state, timeRemaining: seconds }));
	}

	function setSpectatorCount(count: number) {
		update(state => ({ ...state, spectatorCount: count }));
	}

	function clearMatch() {
		set({
			currentMatch: null,
			participants: new Map(),
			agentStates: new Map(),
			codeChanges: [],
			submissions: [],
			currentTurn: 0,
			timeRemaining: 0,
			isSpectating: false,
			spectatorCount: 0,
			replayData: null
		});
	}

	function setReplayData(data: unknown) {
		update(state => ({ ...state, replayData: data }));
	}

	return {
		subscribe,
		setMatch,
		updateMatch,
		updateParticipant,
		updateAgentState,
		addCodeChange,
		addSubmission,
		setTurn,
		updateTimer,
		setSpectatorCount,
		clearMatch,
		setReplayData
	};
}

export const match = createMatchStore();

// ============================================
// Leaderboard Store
// ============================================

interface LeaderboardState {
	global: LeaderboardEntry[];
	weekly: LeaderboardEntry[];
	byLanguage: Record<Language, LeaderboardEntry[]>;
	byAgent: Record<AgentPersona, LeaderboardEntry[]>;
	loading: boolean;
	error: string | null;
	lastUpdated: number | null;
}

function createLeaderboardStore() {
	const { subscribe, set, update } = writable<LeaderboardState>({
		global: [],
		weekly: [],
		byLanguage: {} as Record<Language, LeaderboardEntry[]>,
		byAgent: {} as Record<AgentPersona, LeaderboardEntry[]>,
		loading: false,
		error: null,
		lastUpdated: null
	});

	async function fetchLeaderboard(type: 'global' | 'weekly' | 'language' | 'agent', params?: { language?: Language; agent?: AgentPersona }) {
		update(state => ({ ...state, loading: true, error: null }));
		try {
			const config = get(apiConfig);
			let url = `${config.api_url}/leaderboard/${type}`;
			if (params?.language) url += `?language=${params.language}`;
			if (params?.agent) url += `${params.language ? '&' : '?'}agent=${params.agent}`;

			const response = await fetch(url, {
				headers: { Authorization: `Bearer ${get(auth).tokens?.access_token}` }
			});

			if (!response.ok) throw new Error('Failed to fetch leaderboard');
			const data = await response.json();

			update(state => {
				const newState = { ...state, loading: false, lastUpdated: Date.now() };
				if (type === 'global') newState.global = data.data;
				else if (type === 'weekly') newState.weekly = data.data;
				else if (type === 'language' && params.language) newState.byLanguage[params.language] = data.data;
				else if (type === 'agent' && params.agent) newState.byAgent[params.agent] = data.data;
				return newState;
			});
		} catch (error) {
			update(state => ({ ...state, loading: false, error: String(error) }));
		}
	}

	function refreshAll() {
		fetchLeaderboard('global');
		fetchLeaderboard('weekly');
		// Languages and agents fetched on demand
	}

	return {
		subscribe,
		fetchLeaderboard,
		refreshAll
	};
}

export const leaderboard = createLeaderboardStore();

// ============================================
// Challenges Store
// ============================================

interface ChallengesState {
	challenges: Challenge[];
	templates: ChallengeTemplate[];
	currentChallenge: Challenge | null;
	filters: {
		category?: ChallengeCategory;
		severity?: string;
		search?: string;
		language?: Language;
	};
	loading: boolean;
	error: string | null;
}

import type { ChallengeTemplate } from '$lib/types';

function createChallengesStore() {
	const { subscribe, set, update } = writable<ChallengesState>({
		challenges: [],
		templates: [],
		currentChallenge: null,
		filters: {},
		loading: false,
		error: null
	});

	async function fetchChallenges(filters = {}) {
		update(state => ({ ...state, loading: true, error: null, filters: { ...state.filters, ...filters } }));
		try {
			const config = get(apiConfig);
			const queryParams = new URLSearchParams(filters as Record<string, string>).toString();
			const response = await fetch(`${config.api_url}/challenges?${queryParams}`, {
				headers: { Authorization: `Bearer ${get(auth).tokens?.access_token}` }
			});
			if (!response.ok) throw new Error('Failed to fetch challenges');
			const data = await response.json();
			update(state => ({ ...state, challenges: data.data, loading: false }));
		} catch (error) {
			update(state => ({ ...state, loading: false, error: String(error) }));
		}
	}

	async function fetchTemplates() {
		try {
			const config = get(apiConfig);
			const response = await fetch(`${config.api_url}/challenges/templates`, {
				headers: { Authorization: `Bearer ${get(auth).tokens?.access_token}` }
			});
			if (!response.ok) throw new Error('Failed to fetch templates');
			const data = await response.json();
			update(state => ({ ...state, templates: data.data }));
		} catch (error) {
			console.error('Failed to fetch templates:', error);
		}
	}

	async function createChallenge(challenge: Challenge) {
		update(state => ({ ...state, loading: true }));
		try {
			const config = get(apiConfig);
			const response = await fetch(`${config.api_url}/challenges`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${get(auth).tokens?.access_token}`
				},
				body: JSON.stringify(challenge)
			});
			if (!response.ok) throw new Error('Failed to create challenge');
			const data = await response.json();
			update(state => ({ ...state, challenges: [data.data, ...state.challenges], loading: false }));
			return data.data;
		} catch (error) {
			update(state => ({ ...state, loading: false, error: String(error) }));
			throw error;
		}
	}

	function setCurrentChallenge(challenge: Challenge | null) {
		update(state => ({ ...state, currentChallenge: challenge }));
	}

	return {
		subscribe,
		fetchChallenges,
		fetchTemplates,
		createChallenge,
		setCurrentChallenge
	};
}

export const challenges = createChallengesStore();

// ============================================
// Snippets Store
// ============================================

interface SnippetsState {
	snippets: Snippet[];
	collections: SnippetCollection[];
	filters: {
		language?: Language;
		tag?: string;
		complexity?: string;
		search?: string;
	};
	loading: boolean;
	error: string | null;
}

function createSnippetsStore() {
	const { subscribe, set, update } = writable<SnippetsState>({
		snippets: [],
		collections: [],
		filters: {},
		loading: false,
		error: null
	});

	async function fetchSnippets(filters = {}) {
		update(state => ({ ...state, loading: true, error: null, filters: { ...state.filters, ...filters } }));
		try {
			const config = get(apiConfig);
			const queryParams = new URLSearchParams(filters as Record<string, string>).toString();
			const response = await fetch(`${config.api_url}/snippets?${queryParams}`, {
				headers: { Authorization: `Bearer ${get(auth).tokens?.access_token}` }
			});
			if (!response.ok) throw new Error('Failed to fetch snippets');
			const data = await response.json();
			update(state => ({ ...state, snippets: data.data, loading: false }));
		} catch (error) {
			update(state => ({ ...state, loading: false, error: String(error) }));
		}
	}

	async function importSnippets(source: string) {
		update(state => ({ ...state, loading: true }));
		try {
			const config = get(apiConfig);
			const response = await fetch(`${config.api_url}/snippets/import`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${get(auth).tokens?.access_token}`
				},
				body: JSON.stringify({ source })
			});
			if (!response.ok) throw new Error('Failed to import snippets');
			const data = await response.json();
			update(state => ({ ...state, snippets: [...data.data, ...state.snippets], loading: false }));
		} catch (error) {
			update(state => ({ ...state, loading: false, error: String(error) }));
		}
	}

	return {
		subscribe,
		fetchSnippets,
		importSnippets
	};
}

export const snippets = createSnippetsStore();

// ============================================
// UI Store
// ============================================

function createUIStore() {
	const initialState: UIState = {
		theme: 'dark',
		sidebar_open: true,
		active_modal: null,
		notifications: [],
		loading: {},
		toasts: []
	};

	const { subscribe, set, update } = writable<UIState>(
		browser ? JSON.parse(localStorage.getItem('ccr_ui') || JSON.stringify(initialState)) : initialState
	);

	if (browser) {
		subscribe(state => {
			localStorage.setItem('ccr_ui', JSON.stringify(state));
			document.documentElement.setAttribute('data-theme', state.theme);
		});
	}

	function toggleSidebar() {
		update(state => ({ ...state, sidebar_open: !state.sidebar_open }));
	}

	function setTheme(theme: 'dark' | 'light') {
		update(state => ({ ...state, theme }));
	}

	function openModal(modal: string) {
		update(state => ({ ...state, active_modal: modal }));
	}

	function closeModal() {
		update(state => ({ ...state, active_modal: null }));
	}

	function addNotification(notification: Omit<Notification, 'id' | 'read' | 'created_at'>) {
		const newNotification: Notification = {
			...notification,
			id: crypto.randomUUID(),
			read: false,
			created_at: new Date().toISOString()
		};
		update(state => ({ ...state, notifications: [newNotification, ...state.notifications].slice(0, 50) }));
	}

	function markNotificationRead(id: string) {
		update(state => ({
			...state,
			notifications: state.notifications.map(n => n.id === id ? { ...n, read: true } : n)
		}));
	}

	function setLoading(key: string, loading: boolean) {
		update(state => ({ ...state, loading: { ...state.loading, [key]: loading } }));
	}

	function showToast(toast: Omit<Toast, 'id'>) {
		const newToast: Toast = { ...toast, id: crypto.randomUUID() };
		update(state => ({ ...state, toasts: [...state.toasts, newToast] }));
		if (toast.duration !== 0) {
			setTimeout(() => dismissToast(newToast.id), toast.duration || 5000);
		}
	}

	function dismissToast(id: string) {
		update(state => ({ ...state, toasts: state.toasts.filter(t => t.id !== id) }));
	}

	return {
		subscribe,
		toggleSidebar,
		setTheme,
		openModal,
		closeModal,
		addNotification,
		markNotificationRead,
		setLoading,
		showToast,
		dismissToast
	};
}

export const ui = createUIStore();

// ============================================
// Config Store
// ============================================

interface ConfigState {
	config: AppConfig | null;
	loading: boolean;
}

function createConfigStore() {
	const { subscribe, set, update } = writable<ConfigState>({
		config: null,
		loading: true
	});

	async function loadConfig() {
		update(state => ({ ...state, loading: true }));
		try {
			const response = await fetch('/api/config');
			if (!response.ok) throw new Error('Failed to load config');
			const config = await response.json();
			set({ config, loading: false });
		} catch (error) {
			console.error('Failed to load config:', error);
			set({ config: null, loading: false });
		}
	}

	return {
		subscribe,
		loadConfig
	};
}

export const apiConfig = createConfigStore();

// ============================================
// User Stats Store
// ============================================

interface StatsState {
	stats: UserStats | null;
	loading: boolean;
	error: string | null;
}

function createStatsStore() {
	const { subscribe, set, update } = writable<StatsState>({
		stats: null,
		loading: false,
		error: null
	});

	async function fetchStats(userId?: string) {
		update(state => ({ ...state, loading: true, error: null }));
		try {
			const config = get(apiConfig);
			const url = userId
				? `${config.api_url}/users/${userId}/stats`
				: `${config.api_url}/stats/me`;
			const response = await fetch(url, {
				headers: { Authorization: `Bearer ${get(auth).tokens?.access_token}` }
			});
			if (!response.ok) throw new Error('Failed to fetch stats');
			const data = await response.json();
			update(state => ({ ...state, stats: data.data, loading: false }));
		} catch (error) {
			update(state => ({ ...state, loading: false, error: String(error) }));
		}
	}

	return {
		subscribe,
		fetchStats
	};
}

export const stats = createStatsStore();

// ============================================
// Derived Stores
// ============================================

export const currentUser = derived(auth, $auth => $auth.user);
export const isAuthenticated = derived(auth, $auth => $auth.isAuthenticated);
export const currentMatch = derived(match, $match => $match.currentMatch);
export const currentTurn = derived(match, $match => $match.currentTurn);
export const timeRemaining = derived(match, $match => $match.timeRemaining);
export const agentStates = derived(match, $match => $match.agentStates);
export const matchParticipants = derived(match, $match => Array.from($match.participants.values()));
export const myParticipant = derived(
	[match, auth],
	([$match, $auth]) => {
		if (!$match.currentMatch || !$auth.user) return null;
		return $match.participants.get($auth.user.id) || null;
	}
);
export const isMyTurn = derived(
	[match, auth],
	([$match, $auth]) => {
		if (!$match.currentMatch || !$auth.user) return false;
		const participant = $match.participants.get($auth.user.id);
		return participant?.agent_persona === $match.currentMatch?.participants[$match.currentTurn - 1]?.agent_persona;
	}
);

// ============================================
// Utility Functions
// ============================================

export function formatTime(seconds: number): string {
	const mins = Math.floor(seconds / 60);
	const secs = seconds % 60;
	return `${mins}:${secs.toString().padStart(2, '0')}`;
}

export function formatRating(rating: number): string {
	return rating.toLocaleString();
}

export function getRatingTier(rating: number): { name: string; color: string } {
	if (rating >= 2400) return { name: 'Grandmaster', color: '#ff3366' };
	if (rating >= 2200) return { name: 'Master', color: '#ff6b6b' };
	if (rating >= 2000) return { name: 'Diamond', color: '#b9f2ff' };
	if (rating >= 1800) return { name: 'Platinum', color: '#e5e4e2' };
	if (rating >= 1600) return { name: 'Gold', color: '#ffd700' };
	if (rating >= 1400) return { name: 'Silver', color: '#c0c0c0' };
	if (rating >= 1200) return { name: 'Bronze', color: '#cd7f32' };
	return { name: 'Unranked', color: '#8b949e' };
}

export function getAgentDisplayName(persona: AgentPersona): string {
	const names: Record<AgentPersona, string> = {
		sentinel: 'Sentinel',
		optimizer: 'Optimizer',
		refactor: 'Refactor',
		guardian: 'Guardian',
		docubot: 'DocuBot',
		polyglot: 'Polyglot'
	};
	return names[persona];
}

export function getLanguageDisplayName(language: Language): string {
	const names: Record<Language, string> = {
		cpp: 'C++',
		python: 'Python',
		java: 'Java',
		javascript: 'JavaScript',
		typescript: 'TypeScript',
		rust: 'Rust',
		shell: 'Shell',
		html: 'HTML',
		gherkin: 'Gherkin'
	};
	return names[language];
}