// Diff utility functions for CodeEditor

export interface DiffLine {
	type: 'context' | 'add' | 'remove';
	content: string;
	line_number?: number;
}

export interface DiffHunk {
	old_start: number;
	old_lines: number;
	new_start: number;
	new_lines: number;
	lines: DiffLine[];
}

/**
 * Simple line-by-line diff algorithm
 * For production, consider using a proper diff library like 'diff' or 'diff-match-patch'
 */
export function createDiff(oldContent: string, newContent: string): DiffLine[] {
	const oldLines = oldContent.split('\n');
	const newLines = newContent.split('\n');
	const result: DiffLine[] = [];

	const maxLen = Math.max(oldLines.length, newLines.length);

	for (let i = 0; i < maxLen; i++) {
		const oldLine = oldLines[i];
		const newLine = newLines[i];

		if (oldLine === newLine) {
			if (oldLine !== undefined) {
				result.push({ type: 'context', content: oldLine, line_number: i + 1 });
			}
		} else if (oldLine !== undefined && newLine !== undefined) {
			result.push({ type: 'remove', content: oldLine, line_number: i + 1 });
			result.push({ type: 'add', content: newLine, line_number: i + 1 });
		} else if (oldLine !== undefined) {
			result.push({ type: 'remove', content: oldLine, line_number: i + 1 });
		} else if (newLine !== undefined) {
			result.push({ type: 'add', content: newLine, line_number: i + 1 });
		}
	}

	return result;
}

/**
 * Parse unified diff format into DiffLine array
 */
export function parseUnifiedDiff(diff: string): DiffLine[] {
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

/**
 * Generate unified diff from two strings
 */
export function generateUnifiedDiff(
	oldContent: string,
	newContent: string,
	oldFile: string = 'original',
	newFile: string = 'modified'
): string {
	const oldLines = oldContent.split('\n');
	const newLines = newContent.split('\n');

	const diffLines: string[] = [];
	diffLines.push(`--- ${oldFile}`);
	diffLines.push(`+++ ${newFile}`);

	// Simple line-by-line diff
	const maxLen = Math.max(oldLines.length, newLines.length);

	for (let i = 0; i < maxLen; i++) {
		const oldLine = oldLines[i];
		const newLine = newLines[i];

		if (oldLine === newLine) {
			if (oldLine !== undefined) {
				diffLines.push(` ${oldLine}`);
			}
		} else if (oldLine !== undefined && newLine !== undefined) {
			diffLines.push(`-${oldLine}`);
			diffLines.push(`+${newLine}`);
		} else if (oldLine !== undefined) {
			diffLines.push(`-${oldLine}`);
		} else if (newLine !== undefined) {
			diffLines.push(`+${newLine}`);
		}
	}

	return diffLines.join('\n');
}

/**
 * Apply a diff to content
 */
export function applyDiff(content: string, diffLines: DiffLine[]): string {
	const lines = content.split('\n');
	const result: string[] = [];

	for (const diff of diffLines) {
		if (diff.type === 'context' || diff.type === 'add') {
			result.push(diff.content);
		}
		// 'remove' lines are skipped
	}

	return result.join('\n');
}

/**
 * Generate hunks for Monaco diff editor
 */
export function generateHunks(oldContent: string, newContent: string) {
	const oldLines = oldContent.split('\n');
	const newLines = newContent.split('\n');

	const hunks = [];
	let oldStart = 0;
	let newStart = 0;

	while (oldStart < oldLines.length || newStart < newLines.length) {
		// Find matching lines
		let oldEnd = oldStart;
		let newEnd = newStart;

		// Find matching context (up to 3 lines)
		while (oldEnd < oldLines.length && newEnd < newLines.length && oldLines[oldEnd] === newLines[newEnd]) {
			oldEnd++;
			newEnd++;
		}

		if (oldEnd > oldStart || newEnd > newStart) {
			hunks.push({
				oldStart: oldStart + 1,
				oldLines: oldEnd - oldStart,
				newStart: newStart + 1,
				newLines: newEnd - newStart,
				lines: oldLines.slice(oldStart, oldEnd).map((content, i) => ({
					type: 'context' as const,
					content,
					line_number: oldStart + i + 1
				}))
			});
		}

		oldStart = oldEnd;
		newStart = newEnd;

		// Find changes
		while (oldStart < oldLines.length || newStart < newLines.length) {
			if (oldStart < oldLines.length && newStart < newLines.length) {
				if (oldLines[oldStart] !== newLines[newStart]) {
					break;
				}
			} else {
				break;
			}
			oldStart++;
			newStart++;
		}

		if (oldStart >= oldLines.length && newStart >= newLines.length) break;
	}

	return hunks;
}

/**
 * Calculate similarity between two strings (0-1)
 */
export function similarity(a: string, b: string): number {
	if (a === b) return 1;
	if (a.length === 0 || b.length === 0) return 0;

	const longer = a.length > b.length ? a : b;
	const shorter = a.length > b.length ? b : a;

	const longerLength = longer.length;
	if (longerLength === 0) return 1;

	const editDistance = levenshteinDistance(longer, shorter);
	return (longerLength - editDistance) / longerLength;
}

function levenshteinDistance(a: string, b: string): number {
	const matrix: number[][] = [];

	for (let i = 0; i <= b.length; i++) {
		matrix[i] = [i];
	}

	for (let j = 0; j <= a.length; j++) {
		matrix[0][j] = j;
	}

	for (let i = 1; i <= b.length; i++) {
		for (let j = 1; j <= a.length; j++) {
			if (b.charAt(i - 1) === a.charAt(j - 1)) {
				matrix[i][j] = matrix[i - 1][j - 1];
			} else {
				matrix[i][j] = Math.min(
					matrix[i - 1][j - 1] + 1, // substitution
					matrix[i][j - 1] + 1,     // insertion
					matrix[i - 1][j] + 1      // deletion
				);
			}
		}
	}

	return matrix[b.length][a.length];
}