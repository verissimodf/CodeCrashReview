import { describe, it, expect } from 'vitest';
import {
	formatRating,
	formatCompactNumber,
	formatSeconds,
	formatRelativeTime,
	truncate,
	slugify,
	capitalize,
	validateEmail,
	validateUsername,
	validatePassword,
	validateUrl,
	hexToRgb,
	rgbToHex,
	getContrastColor,
	groupBy,
	sortBy,
	uniqueBy,
	chunk,
	generateId,
	sleep,
	debounce,
	throttle,
	getLanguageFromPath,
	getMonacoLanguageId,
	getAgentDisplayInfo,
	getSeverityColor,
	getCategoryIcon,
	getRatingTier,
	calculateEloChange
} from './index';

describe('Utils', () => {
	describe('formatRating', () => {
		it('formats rating with commas', () => {
			expect(formatRating(1200)).toBe('1,200');
			expect(formatRating(1500000)).toBe('1,500,000');
		});
	});

	describe('formatCompactNumber', () => {
		it('formats compact numbers', () => {
			expect(formatCompactNumber(1000)).toBe('1K');
			expect(formatCompactNumber(1500)).toBe('1.5K');
			expect(formatCompactNumber(1000000)).toBe('1M');
		});
	});

	describe('formatSeconds', () => {
		it('formats seconds to MM:SS', () => {
			expect(formatSeconds(0)).toBe('0:00');
			expect(formatSeconds(30)).toBe('0:30');
			expect(formatSeconds(60)).toBe('1:00');
			expect(formatSeconds(125)).toBe('2:05');
			expect(formatSeconds(3661)).toBe('61:01');
		});
	});

	describe('formatRelativeTime', () => {
		it('formats relative time', () => {
			const now = new Date();
			expect(formatRelativeTime(now)).toBe('agora mesmo');

			const minuteAgo = new Date(now.getTime() - 60000);
			expect(formatRelativeTime(minuteAgo)).toBe('1m atrás');

			const hourAgo = new Date(now.getTime() - 3600000);
			expect(formatRelativeTime(hourAgo)).toBe('1h atrás');

			const dayAgo = new Date(now.getTime() - 86400000);
			expect(formatRelativeTime(dayAgo)).toBe('1d atrás');
		});
	});

	describe('truncate', () => {
		it('truncates long strings', () => {
			expect(truncate('hello world', 8)).toBe('hello…');
			expect(truncate('hi', 10)).toBe('hi');
		});
	});

	describe('slugify', () => {
		it('creates URL-friendly slugs', () => {
			expect(slugify('Hello World')).toBe('hello-world');
			expect(slugify('Test@#$%String')).toBe('test-string');
			expect(slugify('  leading trailing  ')).toBe('leading-trailing');
		});
	});

	describe('capitalize', () => {
		it('capitalizes first letter', () => {
			expect(capitalize('hello')).toBe('Hello');
			expect(capitalize('WORLD')).toBe('World');
		});
	});

	describe('validateEmail', () => {
		it('validates email addresses', () => {
			expect(validateEmail('test@example.com')).toBe(true);
			expect(validateEmail('user.name@domain.org')).toBe(true);
			expect(validateEmail('invalid')).toBe(false);
			expect(validateEmail('missing@domain')).toBe(false);
			expect(validateEmail('@nodomain.com')).toBe(false);
		});
	});

	describe('validateUsername', () => {
		it('validates usernames', () => {
			expect(validateUsername('validuser')).toBe(true);
			expect(validateUsername('user_123')).toBe(true);
			expect(validateUsername('user-name')).toBe(true);
			expect(validateUsername('ab')).toBe(false); // too short
			expect(validateUsername('verylongusernamethatexceedslimit')).toBe(false); // too long
			expect(validateUsername('user@name')).toBe(false); // invalid char
		});
	});

	describe('validatePassword', () => {
		it('validates password strength', () => {
			const strong = validatePassword('StrongPass1!');
			expect(strong.valid).toBe(true);
			expect(strong.errors).toHaveLength(0);

			const weak = validatePassword('weak');
			expect(weak.valid).toBe(false);
			expect(weak.errors.length).toBeGreaterThan(0);
		});
	});

	describe('validateUrl', () => {
		it('validates URLs', () => {
			expect(validateUrl('https://example.com')).toBe(true);
			expect(validateUrl('http://localhost:3000')).toBe(true);
			expect(validateUrl('not-a-url')).toBe(false);
		});
	});

	describe('Color utilities', () => {
		describe('hexToRgb', () => {
			it('converts hex to RGB', () => {
				expect(hexToRgb('#ff0000')).toEqual({ r: 255, g: 0, b: 0 });
				expect(hexToRgb('#00ff00')).toEqual({ r: 0, g: 255, b: 0 });
				expect(hexToRgb('#0000ff')).toEqual({ r: 0, g: 0, b: 255 });
				expect(hexToRgb('invalid')).toBeNull();
			});
		});

		describe('rgbToHex', () => {
			it('converts RGB to hex', () => {
				expect(rgbToHex(255, 0, 0)).toBe('#ff0000');
				expect(rgbToHex(0, 255, 0)).toBe('#00ff00');
				expect(rgbToHex(0, 0, 255)).toBe('#0000ff');
			});
		});

		describe('getContrastColor', () => {
			it('returns appropriate contrast color', () => {
				expect(getContrastColor('#ffffff')).toBe('#0d1117'); // white -> dark
				expect(getContrastColor('#000000')).toBe('#ffffff'); // black -> light
				expect(getContrastColor('#ff0000')).toBe('#ffffff'); // red -> light
			});
		});
	});

	describe('Array utilities', () => {
		describe('groupBy', () => {
			it('groups array by key', () => {
				const arr = [
					{ id: 1, category: 'a' },
					{ id: 2, category: 'b' },
					{ id: 3, category: 'a' }
				];
				const grouped = groupBy(arr, 'category');
				expect(grouped.a).toHaveLength(2);
				expect(grouped.b).toHaveLength(1);
			});

			it('groups by function', () => {
				const arr = [1, 2, 3, 4, 5, 6];
				const grouped = groupBy(arr, x => x % 2 === 0 ? 'even' : 'odd');
				expect(grouped.even).toEqual([2, 4, 6]);
				expect(grouped.odd).toEqual([1, 3, 5]);
			});
		});

		describe('sortBy', () => {
			it('sorts array by key', () => {
				const arr = [
					{ id: 3, name: 'c' },
					{ id: 1, name: 'a' },
					{ id: 2, name: 'b' }
				];
				expect(sortBy(arr, 'id')).toEqual([
					{ id: 1, name: 'a' },
					{ id: 2, name: 'b' },
					{ id: 3, name: 'c' }
				]);
				expect(sortBy(arr, 'id', 'desc')).toEqual([
					{ id: 3, name: 'c' },
					{ id: 2, name: 'b' },
					{ id: 1, name: 'a' }
				]);
			});
		});

		describe('uniqueBy', () => {
			it('removes duplicates by key', () => {
				const arr = [
					{ id: 1, name: 'a' },
					{ id: 2, name: 'b' },
					{ id: 1, name: 'c' } // duplicate id
				];
				const unique = uniqueBy(arr, 'id');
				expect(unique).toHaveLength(2);
			});
		});

		describe('chunk', () => {
			it('splits array into chunks', () => {
				const arr = [1, 2, 3, 4, 5, 6, 7];
				expect(chunk(arr, 3)).toEqual([[1, 2, 3], [4, 5, 6], [7]]);
			});
		});
	});

	describe('generateId', () => {
		it('generates unique IDs', () => {
			const id1 = generateId();
			const id2 = generateId();
			expect(id1).not.toBe(id2);
			expect(id1.startsWith('test-')).toBe(false); // no prefix by default
		});

		it('accepts prefix', () => {
			const id = generateId('prefix-');
			expect(id.startsWith('prefix-')).toBe(true);
		});
	});

	describe('Language utilities', () => {
		describe('getLanguageFromPath', () => {
			it('detects language from file extension', () => {
				expect(getLanguageFromPath('main.py')).toBe('python');
				expect(getLanguageFromPath('Main.java')).toBe('java');
				expect(getLanguageFromPath('script.js')).toBe('javascript');
				expect(getLanguageFromPath('app.ts')).toBe('typescript');
				expect(getLanguageFromPath('main.rs')).toBe('rust');
				expect(getLanguageFromPath('script.sh')).toBe('shell');
				expect(getLanguageFromPath('index.html')).toBe('html');
				expect(getLanguageFromPath('test.feature')).toBe('gherkin');
				expect(getLanguageFromPath('unknown.xyz')).toBeNull();
			});
		});

		describe('getMonacoLanguageId', () => {
			it('returns Monaco editor language IDs', () => {
				expect(getMonacoLanguageId('python')).toBe('python');
				expect(getMonacoLanguageId('javascript')).toBe('javascript');
				expect(getMonacoLanguageId('typescript')).toBe('typescript');
				expect(getMonacoLanguageId('rust')).toBe('rust');
			});
		});
	});

	describe('Agent utilities', () => {
		describe('getAgentDisplayInfo', () => {
			it('returns agent display info', () => {
				const sentinel = getAgentDisplayInfo('sentinel');
				expect(sentinel.name).toBe('Sentinel');
				expect(sentinel.color).toBe('#ff3366');
				expect(sentinel.icon).toBe('shield-alert');
				expect(sentinel.description).toBe('Segurança e vulnerabilidades');
			});
		});
	});

	describe('Challenge utilities', () => {
		describe('getSeverityColor', () => {
			it('returns severity colors', () => {
				expect(getSeverityColor('low')).toBe('#00ff87');
				expect(getSeverityColor('medium')).toBe('#ffaa00');
				expect(getSeverityColor('high')).toBe('#ff8800');
				expect(getSeverityColor('critical')).toBe('#ff3366');
			});
		});

		describe('getCategoryIcon', () => {
			it('returns category icons', () => {
				expect(getCategoryIcon('security')).toBe('shield');
				expect(getCategoryIcon('performance')).toBe('zap');
				expect(getCategoryIcon('refactor')).toBe('code-2');
				expect(getCategoryIcon('testing')).toBe('check-circle-2');
			});
		});
	});

	describe('Rating utilities', () => {
		describe('getRatingTier', () => {
			it('returns correct tier for rating', () => {
				expect(getRatingTier(1000).name).toBe('Unranked');
				expect(getRatingTier(1300).name).toBe('Bronze');
				expect(getRatingTier(1500).name).toBe('Silver');
				expect(getRatingTier(1700).name).toBe('Gold');
				expect(getRatingTier(1900).name).toBe('Platinum');
				expect(getRatingTier(2100).name).toBe('Diamond');
				expect(getRatingTier(2300).name).toBe('Master');
				expect(getRatingTier(2500).name).toBe('Grandmaster');
			});
		});

		describe('calculateEloChange', () => {
			it('calculates Elo rating change', () => {
				// Equal ratings, win = +16 (K=32 * 0.5)
				expect(calculateEloChange(1500, 1500, 1)).toBe(16);
				// Equal ratings, loss = -16
				expect(calculateEloChange(1500, 1500, 0)).toBe(-16);
				// Equal ratings, draw = 0
				expect(calculateEloChange(1500, 1500, 0.5)).toBe(0);

				// Higher rated player wins (expected)
				expect(calculateEloChange(1800, 1500, 1)).toBeLessThan(16);
				// Lower rated player wins (upset)
				expect(calculateEloChange(1500, 1800, 1)).toBeGreaterThan(16);
			});
		});
	});

	describe('Async utilities', () => {
		it('sleep resolves after delay', async () => {
			const start = Date.now();
			await sleep(50);
			const elapsed = Date.now() - start;
			expect(elapsed).toBeGreaterThanOrEqual(45);
		});
	});
});