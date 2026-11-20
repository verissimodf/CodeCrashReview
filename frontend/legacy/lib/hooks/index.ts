// CodeCrashReview - Custom Svelte Hooks
// Reusable logic for components

import { onMount, onDestroy } from 'svelte';
import { writable, derived, get } from 'svelte/store';
import { browser } from '$app/environment';
import { ws } from '$lib/stores';
import type { WSMessage, WSMessageType, AgentStatus } from '$lib/types';

// ============================================
// WebSocket Hook
// ============================================

export function useWebSocket() {
	const handlers = new Map<WSMessageType, Set<(payload: unknown) => void>>();

	function on<T>(type: WSMessageType, handler: (payload: T) => void) {
		if (!handlers.has(type)) handlers.set(type, new Set());
		handlers.get(type)!.add(handler as (payload: unknown) => void);

		return () => {
			const set = handlers.get(type);
			if (set) {
				set.delete(handler as (payload: unknown) => void);
				if (set.size === 0) handlers.delete(type);
			}
		};
	}

	function off(type: WSMessageType, handler?: (payload: unknown) => void) {
		const set = handlers.get(type);
		if (!set) return;
		if (handler) {
			set.delete(handler);
		} else {
			set.clear();
		}
		if (set.size === 0) handlers.delete(type);
	}

	// Subscribe to WS store messages
	const unsubscribe = ws.subscribe(state => {
		if (state.lastMessage) {
			const set = handlers.get(state.lastMessage.type);
			if (set) {
				set.forEach(handler => handler(state.lastMessage!.payload));
			}
		}
	});

	onDestroy(() => {
		unsubscribe();
		handlers.clear();
	});

	return { on, off };
}

// ============================================
// Match Hook
// ============================================

export function useMatch(matchId: string, asSpectator = false) {
	const { on, off } = useWebSocket();

	onMount(() => {
		get(ws).connect(matchId);
	});

	onDestroy(() => {
		get(ws).disconnect();
	});

	return {
		send: (message: WSMessage) => get(ws).send(message),
		onAgentStatus: (handler: (payload: { participant_id: string; agent_persona: string; status: AgentStatus; progress: number }) => void) =>
			on('agent_status', handler),
		onCodeChange: (handler: (payload: { participant_id: string; agent_persona: string; change: any }) => void) =>
			on('code_change', handler),
		onSubmission: (handler: (payload: any) => void) =>
			on('submission', handler),
		onTurnChange: (handler: (payload: { turn: number; time_remaining_seconds: number }) => void) =>
			on('turn_change', handler),
		onTimerUpdate: (handler: (payload: { time_remaining_seconds: number; warning: boolean }) => void) =>
			on('timer_update', handler),
		onMatchEnd: (handler: (payload: any) => void) =>
			on('match_end', handler)
	};
}

// ============================================
// Timer Hook
// ============================================

export function useTimer(initialSeconds: number, options: { onWarning?: () => void; onDanger?: () => void; onEnd?: () => void } = {}) {
	const { onWarning, onDanger, onEnd } = options;
	const time = writable(initialSeconds);
	const isRunning = writable(false);
	let interval: ReturnType<typeof setInterval> | null = null;
	let warningTriggered = false;
	let dangerTriggered = false;

	function start() {
		if (get(isRunning)) return;
		isRunning.set(true);
		warningTriggered = false;
		dangerTriggered = false;
		interval = setInterval(() => {
			time.update(t => {
				if (t <= 1) {
					stop();
					onEnd?.();
					return 0;
				}
				const newTime = t - 1;
				if (newTime <= 10 && !dangerTriggered) {
					dangerTriggered = true;
					onDanger?.();
				} else if (newTime <= 30 && !warningTriggered) {
					warningTriggered = true;
					onWarning?.();
				}
				return newTime;
			});
		}, 1000);
	}

	function stop() {
		if (interval) {
			clearInterval(interval);
			interval = null;
		}
		isRunning.set(false);
	}

	function reset(seconds: number) {
		stop();
		time.set(seconds);
		warningTriggered = false;
		dangerTriggered = false;
	}

	function pause() {
		stop();
	}

	function resume() {
		start();
	}

	onDestroy(() => stop());

	return {
		time,
		isRunning,
		start,
		stop,
		reset,
		pause,
		resume
	};
}

// ============================================
// Debounce Hook
// ============================================

export function useDebounce<T>(value: T, delay: number): T {
	const debounced = writable(value);

	onMount(() => {
		let timeout: ReturnType<typeof setTimeout>;
		const unsubscribe = value.subscribe((v: T) => {
			clearTimeout(timeout);
			timeout = setTimeout(() => debounced.set(v), delay);
		});
		return () => {
			unsubscribe();
			clearTimeout(timeout);
		};
	});

	return get(debounced);
}

// ============================================
// LocalStorage Hook
// ============================================

export function useLocalStorage<T>(key: string, initialValue: T) {
	const stored = browser ? localStorage.getItem(key) : null;
	const value = writable<T>(stored ? JSON.parse(stored) : initialValue);

	if (browser) {
		value.subscribe(v => {
			localStorage.setItem(key, JSON.stringify(v));
		});
	}

	return value;
}

// ============================================
// Media Query Hook
// ============================================

export function useMediaQuery(query: string): boolean {
	const matches = writable(false);

	onMount(() => {
		if (!browser) return;
		const media = window.matchMedia(query);
		matches.set(media.matches);
		const handler = (e: MediaQueryListEvent) => matches.set(e.matches);
		media.addEventListener('change', handler);
		return () => media.removeEventListener('change', handler);
	});

	return get(matches);
}

// ============================================
// Intersection Observer Hook
// ============================================

export function useIntersectionObserver(
	element: HTMLElement | null,
	options: IntersectionObserverInit = {}
) {
	const isIntersecting = writable(false);

	onMount(() => {
		if (!element || !browser) return;
		const observer = new IntersectionObserver(([entry]) => {
			isIntersecting.set(entry.isIntersecting);
		}, options);
		observer.observe(element);
		return () => observer.disconnect();
	}, [element]);

	return isIntersecting;
}

// ============================================
// Keyboard Shortcuts Hook
// ============================================

export function useKeyboardShortcuts(shortcuts: Record<string, (event: KeyboardEvent) => void>) {
	onMount(() => {
		if (!browser) return;
		const handler = (event: KeyboardEvent) => {
			const key = [
				event.ctrlKey ? 'ctrl' : '',
				event.metaKey ? 'meta' : '',
				event.shiftKey ? 'shift' : '',
				event.altKey ? 'alt' : '',
				event.key.toLowerCase()
			].filter(Boolean).join('+');

			if (shortcuts[key]) {
				event.preventDefault();
				shortcuts[key](event);
			}
		};
		window.addEventListener('keydown', handler);
		return () => window.removeEventListener('keydown', handler);
	});
}

// ============================================
// Click Outside Hook
// ============================================

export function useClickOutside(element: HTMLElement | null, handler: (event: MouseEvent) => void) {
	onMount(() => {
		if (!browser || !element) return;
		const handleClick = (event: MouseEvent) => {
			if (!element.contains(event.target as Node)) {
				handler(event);
			}
		};
		document.addEventListener('mousedown', handleClick);
		return () => document.removeEventListener('mousedown', handleClick);
	}, [element]);
}

// ============================================
// Animation Frame Hook
// ============================================

export function useAnimationFrame(callback: (deltaTime: number) => void) {
	let frameId: number;
	let lastTime = 0;

	onMount(() => {
		if (!browser) return;
		const loop = (time: number) => {
			const deltaTime = time - lastTime;
			lastTime = time;
			callback(deltaTime);
			frameId = requestAnimationFrame(loop);
		};
		frameId = requestAnimationFrame(loop);
		return () => cancelAnimationFrame(frameId);
	});
}

// ============================================
// Copy to Clipboard Hook
// ============================================

export function useClipboard() {
	const copied = writable(false);

	async function copy(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			copied.set(true);
			setTimeout(() => copied.set(false), 2000);
			return true;
		} catch {
			copied.set(false);
			return false;
		}
	}

	return { copied, copy };
}

// ============================================
// Fetch Hook
// ============================================

export function useFetch<T>(url: string | null, options: RequestInit = {}) {
	const data = writable<T | null>(null);
	const loading = writable(false);
	const error = writable<string | null>(null);

	async function execute() {
		if (!url) return;
		loading.set(true);
		error.set(null);
		try {
			const response = await fetch(url, {
				...options,
				headers: {
					'Content-Type': 'application/json',
					...options.headers
				}
			});
			if (!response.ok) throw new Error(`HTTP ${response.status}`);
			const json = await response.json();
			data.set(json);
		} catch (e) {
			error.set(String(e));
			data.set(null);
		} finally {
			loading.set(false);
		}
	}

	onMount(() => {
		execute();
	}, [url]);

	return { data, loading, error, refetch: execute };
}

// ============================================
// WebSocket Connection Status Hook
// ============================================

export function useConnectionStatus() {
	const status = derived(ws, $ws => ({
		connected: $ws.connected,
		connecting: $ws.connecting,
		error: $ws.error,
		reconnectAttempts: $ws.reconnectAttempts
	}));

	return status;
}