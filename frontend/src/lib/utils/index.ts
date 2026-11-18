// CodeCrashReview - Utility Functions
// Helper functions, formatters, validators

import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { Language, AgentPersona, DiffLine, CodeChange, TestResult } from '$lib/types';

// ============================================
// Class Name Utilities
// ============================================

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// ============================================
// Date & Time Formatting
// ============================================

export function formatDate(date: string | Date, options: Intl.DateTimeFormatOptions = {}): string {
	const d = typeof date === 'string' ? new Date(date) : date;
	return d.toLocaleDateString('pt-BR', {
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		...options
	});
}

export function formatDateTime(date: string | Date): string {
	return formatDate(date, { hour: '2-digit', minute: '2-digit' });
}

export function formatRelativeTime(date: string | Date): string {
	const d = typeof date === 'string' ? new Date(date) : date;
	const now = new Date();
	const diffMs = now.getTime() - d.getTime();
	const diffSecs = Math.floor(diffMs / 1000);
	const diffMins = Math.floor(diffSecs / 60);
	const diffHours = Math.floor(diffMins / 60);
	const diffDays = Math.floor(diffHours / 24);

	if (diffSecs < 60) return 'agora mesmo';
	if (diffMins < 60) return `${diffMins}m atrás`;
	if (diffHours < 24) return `${diffHours}h atrás`;
	if (diffDays < 7) return `${diffDays}d atrás`;
	return formatDate(d);
}

export function formatDuration(ms: number): string {
	if (ms < 1000) return `${ms}ms`;
	const secs = Math.floor(ms / 1000);
	if (secs < 60) return `${secs}s`;
	const mins = Math.floor(secs / 60);
	const remainingSecs = secs % 60;
	if (mins < 60) return `${mins}m ${remainingSecs}s`;
	const hours = Math.floor(mins / 60);
	const remainingMins = mins % 60;
	return `${hours}h ${remainingMins}m`;
}

export function formatSeconds(seconds: number): string {
	const mins = Math.floor(seconds / 60);
	const secs = seconds % 60;
	return `${mins}:${secs.toString().padStart(2, '0')}`;
}

// ============================================
// Number Formatting
// ============================================

export function formatNumber(num: number): string {
	return new Intl.NumberFormat('pt-BR').format(num);
}

export function formatCompactNumber(num: number): string {
	return new Intl.NumberFormat('pt-BR', { notation: 'compact', maximumFractionDigits: 1 }).format(num);
}

export function formatRating(rating: number): string {
	return rating.toLocaleString('pt-BR');
}

export function formatPercent(value: number, decimals = 1): string {
	return `${value.toFixed(decimals)}%`;
}

export function formatBytes(bytes: number): string {
	if (bytes === 0) return '0 B';
	const k = 1024;
	const sizes = ['B', 'KB', 'MB', 'GB'];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}

// ============================================
// String Utilities
// ============================================

export function truncate(str: string, length: number, suffix = '…'): string {
	if (str.length <= length) return str;
	return str.slice(0, length - suffix.length) + suffix;
}

export function slugify(str: string): string {
	return str
		.toLowerCase()
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/(^-|-$)/g, '');
}

export function capitalize(str: string): string {
	return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
}

export function camelToKebab(str: string): string {
	return str.replace(/([a-z])([A-Z])/g, '$1-$2').toLowerCase();
}

export function kebabToCamel(str: string): string {
	return str.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase());
}

// ============================================
// Code & Diff Utilities
// ============================================

export function getLanguageFromPath(path: string): Language | null {
	const ext = path.split('.').pop()?.toLowerCase();
	const map: Record<string, Language> = {
		cpp: 'cpp', cc: 'cpp', cxx: 'cpp', h: 'cpp', hpp: 'cpp',
		py: 'python', pyw: 'python',
		java: 'java',
		js: 'javascript', mjs: 'javascript', cjs: 'javascript',
		ts: 'typescript', tsx: 'typescript',
		rs: 'rust',
		sh: 'shell', bash: 'shell', zsh: 'shell', fish: 'shell',
		html: 'html', htm: 'html',
		feature: 'gherkin'
	};
	return map[ext || ''] || null;
}

export function getMonacoLanguageId(language: Language): string {
	const map: Record<Language, string> = {
		cpp: 'cpp',
		python: 'python',
		java: 'java',
		javascript: 'javascript',
		typescript: 'typescript',
		rust: 'rust',
		shell: 'shell',
		html: 'html',
		gherkin: 'gherkin'
	};
	return map[language];
}

export function parseDiff(diff: string): DiffLine[] {
	const lines: DiffLine[] = [];
	const diffLines = diff.split('\n');

	for (const line of diffLines) {
		if (line.startsWith('+') && !line.startsWith('+++')) {
			lines.push({ type: 'add', content: line.slice(1) });
		} else if (line.startsWith('-') && !line.startsWith('---')) {
			lines.push({ type: 'remove', content: line.slice(1) });
		} else if (line.startsWith(' ')) {
			lines.push({ type: 'context', content: line.slice(1) });
		}
	}

	return lines;
}

export function generateDiff(oldContent: string, newContent: string): DiffLine[] {
	// Simple line-by-line diff (for production, use a proper diff library)
	const oldLines = oldContent.split('\n');
	const newLines = newContent.split('\n');
	const lines: DiffLine[] = [];

	const maxLen = Math.max(oldLines.length, newLines.length);
	for (let i = 0; i < maxLen; i++) {
		const oldLine = oldLines[i];
		const newLine = newLines[i];

		if (oldLine === newLine) {
			if (oldLine !== undefined) {
				lines.push({ type: 'context', content: oldLine, line_number: i + 1 });
			}
		} else if (oldLine !== undefined && newLine !== undefined) {
			lines.push({ type: 'remove', content: oldLine, line_number: i + 1 });
			lines.push({ type: 'add', content: newLine, line_number: i + 1 });
		} else if (oldLine !== undefined) {
			lines.push({ type: 'remove', content: oldLine, line_number: i + 1 });
		} else if (newLine !== undefined) {
			lines.push({ type: 'add', content: newLine, line_number: i + 1 });
		}
	}

	return lines;
}

export function applyChanges(content: string, changes: CodeChange[]): string {
	let result = content;
	// Sort changes by line number descending to avoid offset issues
	const sortedChanges = [...changes].sort((a, b) => {
		const aLine = a.diff[0]?.old_start || 0;
		const bLine = b.diff[0]?.old_start || 0;
		return bLine - aLine;
	});

	for (const change of sortedChanges) {
		// This is simplified - real implementation would use proper patching
		result = change.new_content;
	}

	return result;
}

// ============================================
// Test Result Utilities
// ============================================

export function getTestSummary(result: TestResult): { passed: number; failed: number; passRate: number } {
	return {
		passed: result.passed,
		failed: result.failed,
		passRate: result.total > 0 ? (result.passed / result.total) * 100 : 0
	};
}

export function getFailedTests(result: TestResult): TestResult['suites'][0]['tests'] {
	const failed: TestResult['suites'][0]['tests'] = [];
	for (const suite of result.suites) {
		for (const test of suite.tests) {
			if (test.status === 'failed') {
				failed.push({ ...test, name: `${suite.name} > ${test.name}` });
			}
		}
	}
	return failed;
}

// ============================================
// Validation
// ============================================

export function validateEmail(email: string): boolean {
	return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export function validateUsername(username: string): boolean {
	return /^[a-zA-Z0-9_-]{3,20}$/.test(username);
}

export function validatePassword(password: string): { valid: boolean; errors: string[] } {
	const errors: string[] = [];
	if (password.length < 8) errors.push('Mínimo 8 caracteres');
	if (!/[A-Z]/.test(password)) errors.push('Pelo menos uma maiúscula');
	if (!/[a-z]/.test(password)) errors.push('Pelo menos uma minúscula');
	if (!/[0-9]/.test(password)) errors.push('Pelo menos um número');
	if (!/[!@#$%^&*]/.test(password)) errors.push('Pelo menos um caractere especial');
	return { valid: errors.length === 0, errors };
}

export function validateUrl(url: string): boolean {
	try {
		new URL(url);
		return true;
	} catch {
		return false;
	}
}

export function validateGitHubUrl(url: string): boolean {
	return /^https:\/\/github\.com\/[\w-]+\/[\w.-]+/.test(url);
}

// ============================================
// Color Utilities
// ============================================

export function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
	const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
	return result ? {
		r: parseInt(result[1], 16),
		g: parseInt(result[2], 16),
		b: parseInt(result[3], 16)
	} : null;
}

export function rgbToHex(r: number, g: number, b: number): string {
	return '#' + [r, g, b].map(x => x.toString(16).padStart(2, '0')).join('');
}

export function getContrastColor(hex: string): string {
	const rgb = hexToRgb(hex);
	if (!rgb) return '#ffffff';
	const luminance = (0.299 * rgb.r + 0.587 * rgb.g + 0.114 * rgb.b) / 255;
	return luminance > 0.5 ? '#0d1117' : '#ffffff';
}

export function lightenColor(hex: string, percent: number): string {
	const rgb = hexToRgb(hex);
	if (!rgb) return hex;
	return rgbToHex(
		Math.min(255, Math.round(rgb.r + (255 - rgb.r) * percent)),
		Math.min(255, Math.round(rgb.g + (255 - rgb.g) * percent)),
		Math.min(255, Math.round(rgb.b + (255 - rgb.b) * percent))
	);
}

export function darkenColor(hex: string, percent: number): string {
	const rgb = hexToRgb(hex);
	if (!rgb) return hex;
	return rgbToHex(
		Math.max(0, Math.round(rgb.r * (1 - percent))),
		Math.max(0, Math.round(rgb.g * (1 - percent))),
		Math.max(0, Math.round(rgb.b * (1 - percent)))
	);
}

// ============================================
// Array & Object Utilities
// ============================================

export function groupBy<T>(array: T[], key: keyof T | ((item: T) => string)): Record<string, T[]> {
	return array.reduce((groups, item) => {
		const groupKey = typeof key === 'function' ? key(item) : String(item[key]);
		(groups[groupKey] = groups[groupKey] || []).push(item);
		return groups;
	}, {} as Record<string, T[]>);
}

export function sortBy<T>(array: T[], key: keyof T | ((item: T) => unknown), order: 'asc' | 'desc' = 'asc'): T[] {
	return [...array].sort((a, b) => {
		const aVal = typeof key === 'function' ? key(a) : a[key];
		const bVal = typeof key === 'function' ? key(b) : b[key];
		if (aVal < bVal) return order === 'asc' ? -1 : 1;
		if (aVal > bVal) return order === 'asc' ? 1 : -1;
		return 0;
	});
}

export function uniqueBy<T>(array: T[], key: keyof T | ((item: T) => unknown)): T[] {
	const seen = new Set();
	return array.filter(item => {
		const val = typeof key === 'function' ? key(item) : item[key];
		if (seen.has(val)) return false;
		seen.add(val);
		return true;
	});
}

export function chunk<T>(array: T[], size: number): T[][] {
	const chunks: T[][] = [];
	for (let i = 0; i < array.length; i += size) {
		chunks.push(array.slice(i, i + size));
	}
	return chunks;
}

// ============================================
// Random & ID Generation
// ============================================

export function generateId(prefix = ''): string {
	return `${prefix}${Date.now().toString(36)}${Math.random().toString(36).slice(2, 9)}`;
}

export function randomInt(min: number, max: number): number {
	return Math.floor(Math.random() * (max - min + 1)) + min;
}

export function randomChoice<T>(array: T[]): T {
	return array[Math.floor(Math.random() * array.length)];
}

export function shuffle<T>(array: T[]): T[] {
	const result = [...array];
	for (let i = result.length - 1; i > 0; i--) {
		const j = Math.floor(Math.random() * (i + 1));
		[result[i], result[j]] = [result[j], result[i]];
	}
	return result;
}

// ============================================
// Async Utilities
// ============================================

export function sleep(ms: number): Promise<void> {
	return new Promise(resolve => setTimeout(resolve, ms));
}

export function retry<T>(fn: () => Promise<T>, retries = 3, delay = 1000): Promise<T> {
	return fn().catch(err => {
		if (retries === 0) throw err;
		return sleep(delay).then(() => retry(fn, retries - 1, delay * 2));
	});
}

export function timeout<T>(promise: Promise<T>, ms: number): Promise<T> {
	return Promise.race([
		promise,
		new Promise<never>((_, reject) => setTimeout(() => reject(new Error('Timeout')), ms))
	]);
}

export function debounce<T extends (...args: unknown[]) => unknown>(fn: T, ms: number): T {
	let timeoutId: ReturnType<typeof setTimeout>;
	return ((...args: unknown[]) => {
		clearTimeout(timeoutId);
		timeoutId = setTimeout(() => fn(...args), ms);
	}) as T;
}

export function throttle<T extends (...args: unknown[]) => unknown>(fn: T, ms: number): T {
	let lastCall = 0;
	return ((...args: unknown[]) => {
		const now = Date.now();
		if (now - lastCall >= ms) {
			lastCall = now;
			fn(...args);
		}
	}) as T;
}

// ============================================
// File Utilities
// ============================================

export function getFileExtension(filename: string): string {
	return filename.slice(((filename.lastIndexOf('.') - 1) >>> 0) + 2);
}

export function getFileNameWithoutExtension(filename: string): string {
	return filename.replace(/\.[^/.]+$/, '');
}

export function formatFileSize(bytes: number): string {
	return formatBytes(bytes);
}

export function isCodeFile(filename: string): boolean {
	const codeExtensions = ['js', 'ts', 'jsx', 'tsx', 'py', 'java', 'cpp', 'c', 'h', 'rs', 'go', 'php', 'rb', 'swift', 'kt', 'cs', 'sh', 'bash', 'zsh', 'fish', 'html', 'css', 'scss', 'json', 'yaml', 'yml', 'xml', 'md', 'feature'];
	const ext = getFileExtension(filename).toLowerCase();
	return codeExtensions.includes(ext);
}

// ============================================
// URL Utilities
// ============================================

export function buildUrl(base: string, params: Record<string, string | number | boolean | undefined>): string {
	const url = new URL(base);
	for (const [key, value] of Object.entries(params)) {
		if (value !== undefined && value !== null) {
			url.searchParams.set(key, String(value));
		}
	}
	return url.toString();
}

export function parseQueryString(query: string): Record<string, string> {
	const params: Record<string, string> = {};
	new URLSearchParams(query).forEach((value, key) => {
		params[key] = value;
	});
	return params;
}

// ============================================
// Rating/Tier Utilities
// ============================================

export function getRatingTier(rating: number): { name: string; color: string; icon: string } {
	if (rating >= 2400) return { name: 'Grandmaster', color: '#ff3366', icon: 'crown' };
	if (rating >= 2200) return { name: 'Master', color: '#ff6b6b', icon: 'medal' };
	if (rating >= 2000) return { name: 'Diamond', color: '#b9f2ff', icon: 'gem' };
	if (rating >= 1800) return { name: 'Platinum', color: '#e5e4e2', icon: 'award' };
	if (rating >= 1600) return { name: 'Gold', color: '#ffd700', icon: 'star' };
	if (rating >= 1400) return { name: 'Silver', color: '#c0c0c0', icon: 'badge' };
	if (rating >= 1200) return { name: 'Bronze', color: '#cd7f32', icon: 'shield' };
	return { name: 'Unranked', color: '#8b949e', icon: 'user' };
}

export function calculateEloChange(ratingA: number, ratingB: number, scoreA: number, kFactor = 32): number {
	const expectedA = 1 / (1 + Math.pow(10, (ratingB - ratingA) / 400));
	return Math.round(kFactor * (scoreA - expectedA));
}

export function calculateGlicko2Rating(
	rating: number,
	deviation: number,
	volatility: number,
	results: { rating: number; deviation: number; score: number }[],
	tau = 0.5
): { rating: number; deviation: number; volatility: number } {
	// Simplified Glicko-2 implementation
	// For production, use a proper library like 'glicko2'
	const q = Math.log(10) / 400;
	let newVolatility = volatility;

	// This is a placeholder - real implementation is complex
	// See: http://www.glicko.net/glicko/glicko2.pdf

	return { rating, deviation, volatility: newVolatility };
}

// ============================================
// Agent Utilities
// ============================================

export function getAgentDisplayInfo(persona: AgentPersona): { name: string; color: string; glow: string; icon: string; description: string } {
	const info: Record<AgentPersona, { name: string; color: string; glow: string; icon: string; description: string }> = {
		sentinel: { name: 'Sentinel', color: '#ff3366', glow: '#ff336640', icon: 'shield-alert', description: 'Segurança e vulnerabilidades' },
		optimizer: { name: 'Optimizer', color: '#00ff87', glow: '#00ff8740', icon: 'zap', description: 'Performance e complexidade' },
		refactor: { name: 'Refactor', color: '#0099ff', glow: '#0099ff40', icon: 'code-2', description: 'Clean code e arquitetura' },
		guardian: { name: 'Guardian', color: '#ffaa00', glow: '#ffaa0040', icon: 'check-circle-2', description: 'Testes e cobertura' },
		docubot: { name: 'DocuBot', color: '#ff00ff', glow: '#ff00ff40', icon: 'book-open', description: 'Documentação e tipos' },
		polyglot: { name: 'Polyglot', color: '#00ffff', glow: '#00ffff40', icon: 'globe', description: 'Multi-linguagem e migração' }
	};
	return info[persona];
}

export function getAgentColor(persona: AgentPersona): string {
	return getAgentDisplayInfo(persona).color;
}

// ============================================
// Challenge Utilities
// ============================================

export function getSeverityColor(severity: string): string {
	const colors: Record<string, string> = {
		low: '#00ff87',
		medium: '#ffaa00',
		high: '#ff8800',
		critical: '#ff3366'
	};
	return colors[severity] || '#8b949e';
}

export function getCategoryIcon(category: string): string {
	const icons: Record<string, string> = {
		security: 'shield',
		performance: 'zap',
		refactor: 'code-2',
		testing: 'check-circle-2',
		documentation: 'book-open',
		migration: 'git-branch',
		bugfix: 'bug',
		feature: 'sparkles'
	};
	return icons[category] || 'file-code';
}

// ============================================
// Export all
// ============================================

export * from './date';
export * from './validation';
export * from './code';