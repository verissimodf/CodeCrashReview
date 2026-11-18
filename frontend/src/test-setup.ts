// Vitest test setup
import { vi } from 'vitest';

// Mock WebSocket
global.WebSocket = vi.fn().mockImplementation(() => ({
	send: vi.fn(),
	close: vi.fn(),
	addEventListener: vi.fn(),
	removeEventListener: vi.fn(),
	readyState: 1,
})) as any;

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
	observe: vi.fn(),
	unobserve: vi.fn(),
	disconnect: vi.fn(),
}));

// Mock IntersectionObserver
global.IntersectionObserver = vi.fn().mockImplementation(() => ({
	observe: vi.fn(),
	unobserve: vi.fn(),
	disconnect: vi.fn(),
}));

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
	writable: true,
	value: vi.fn().mockImplementation(query => ({
		matches: false,
		media: query,
		onchange: null,
		addListener: vi.fn(),
		removeListener: vi.fn(),
		addEventListener: vi.fn(),
		removeEventListener: vi.fn(),
		dispatchEvent: vi.fn(),
	})),
});

// Mock localStorage
const localStorageMock = {
	getItem: vi.fn(),
	setItem: vi.fn(),
	removeItem: vi.fn(),
	clear: vi.fn(),
};
global.localStorage = localStorageMock as any;

// Mock crypto.randomUUID
Object.defineProperty(global, 'crypto', {
	value: {
		randomUUID: () => 'test-uuid-' + Math.random().toString(36).substr(2, 9),
	},
});

// Suppress console.error in tests (optional)
const originalError = console.error;
beforeAll(() => {
	console.error = (...args) => {
		if (args[0]?.includes?.('Warning:')) return;
		originalError.call(console, ...args);
	};
});

afterAll(() => {
	console.error = originalError;
});