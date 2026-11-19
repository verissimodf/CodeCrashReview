<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { cn } from '$lib/utils';
	import { getMonacoLanguageId, getLanguageFromPath } from '$lib/utils';
	import { Copy, Download, Maximize, Minimize, ChevronLeft, ChevronRight } from 'lucide-svelte';
	import { createDiff } from './diff';

	export let value = '';
	export let originalValue = '';
	export let language = 'typescript';
	export let readOnly = false;
	export let theme = 'vs-dark';
	export let fontSize = 13;
	export let lineHeight = 20;
	export let showDiffMinimap = false;
	export let wordWrap: 'on' | 'off' | 'bounded' = 'on';
	export let tabSize = 2;
	export let showDiff = false;
	export let diffOriginal = '';
	export let fileName = '';
	export let onChange: (value: string) => void;

	let editorContainer: HTMLDivElement;
	let editor: any;
	let monaco: any;
	let originalModel: any;
	let modifiedModel: any;
	let diffEditor: any;
	let isDiffMode = false;

	async function initMonaco() {
		if (typeof window === 'undefined') return;

		// Dynamic import of Monaco
		const monacoModule = await import('monaco-editor');
		monaco = monacoModule;

		// DefineDefine custom theme
		monaco.editor.defineTheme('codecrash', {
			base: 'vs-dark',
			inherit: true,
			rules: [],
			colors: {
				'editor.background': '#0d1117',
				'editor.foreground': '#e6edf3',
				'editor.lineHighlightBackground': '#161b22',
				'editorCursor.foreground': '#00ffff',
				'editor.selectionBackground': '#00ffff30',
				'editor.inactiveSelectionBackground': '#00ffff15',
				'editorWhitespace.foreground': '#30363d',
				'editor.lineHighlightBorder': '#30363d',
				'editorIndentGuide.background': '#30363d',
				'editorIndentGuide.activeBackground': '#484f58',
				'editorRuler.foreground': '#30363d',
			}
		});

		// Set language
		const langId = getMonacoLanguageId(language as any);

		if (showDiff && diffOriginal) {
			// Create diff editor
			diffEditor = monaco.editor.createDiffEditor(editorContainer, {
				theme: 'codecrash',
				readOnly: readOnly,
				fontSize,
				lineHeight,
				minimap: { enabled: showMinimap },
				wordWrap,
				tabSize,
				renderSideBySide: true,
				ignoreTrimWhitespace: true,
			});

			originalModel = monaco.editor.createModel(diffOriginal, langId);
			modifiedModel = monaco.editor.createModel(value, langId);

			diffEditor.setModel({
				original: originalModel,
				modified: modifiedModel
			});

			// Listen for changes
			modifiedModel.onDidChangeContent(() => {
				if (onChange) {
					onChange(modifiedModel.getValue());
				}
			});

			isDiffMode = true;
		} else {
			// Create regular editor
			editor = monaco.editor.create(editorContainer, {
				value,
				language: langId,
				theme: 'codecrash',
				readOnly: readOnly,
				automaticLayout: true,
				minimap: { enabled: showMinimap },
				lineNumbers: 'on',
				fontSize,
				lineHeight,
				scrollBeyondLastLine: false,
				renderWhitespace: 'selection',
				tabSize,
				wordWrap,
			});

			editor.onDidChangeModelContent(() => {
				if (onChange) {
					onChange(editor.getValue());
				}
			});
		}
	}

	function updateValue(newValue: string) {
		if (isDiffMode && modifiedModel) {
			modifiedModel.setValue(newValue);
		} else if (editor) {
			editor.setValue(newValue);
		}
	}

	function updateDiffOriginal(newOriginal: string) {
		if (isDiffMode && originalModel) {
			originalModel.setValue(newOriginal);
		}
	}

	function toggleDiffMode() {
		if (isDiffMode) {
			// Switch to regular editor
			diffEditor?.dispose();
			originalModel?.dispose();
			modifiedModel?.dispose();

			editor = monaco.editor.create(editorContainer, {
				value: value,
				language: getMonacoLanguageId(language as any),
				theme: 'codecrash',
				readOnly: readOnly,
				automaticLayout: true,
				minimap: { enabled: showMinimap },
				lineNumbers: 'on',
				fontSize,
				lineHeight,
				scrollBeyondLastLine: false,
				renderWhitespace: 'selection',
				tabSize,
				wordWrap,
			});

			editor.onDidChangeModelContent(() => {
				if (onChange) {
					onChange(editor.getValue());
				}
			});

			isDiffMode = false;
		} else {
			// Switch to diff editor
			editor?.dispose();

			diffEditor = monaco.editor.createDiffEditor(editorContainer, {
				theme: 'codecrash',
				readOnly: readOnly,
				fontSize,
				lineHeight,
				minimap: { enabled: showMinimap },
				wordWrap,
				tabSize,
				renderSideBySide: true,
				ignoreWhitespace: true,
			});

			originalModel = monaco.editor.createModel(diffOriginal || value, getMonacoLanguageId(language as any));
			modifiedModel = monaco.editor.createModel(value, getMonacoLanguageId(language as any));

			diffEditor.setModel({
				original: originalModel,
				modified: modifiedModel
			});

			modifiedModel.onDidChangeContent(() => {
				if (onChange) {
					onChange(modifiedModel.getValue());
				}
			});

			isDiffMode = true;
		}
	}

	function copyContent() {
		const content = isDiffMode ? modifiedModel?.getValue() : editor?.getValue();
		if (content) {
			navigator.clipboard.writeText(content);
		}
	}

	function downloadContent() {
		const content = isDiffMode ? modifiedModel?.getValue() : editor?.getValue();
		if (content) {
			const blob = new Blob([content], { type: 'text/plain' });
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = fileName || `code.${language}`;
			a.click();
			URL.revokeObjectURL(url);
		}
	}

	onMount(() => {
		initMonaco();
	});

	onDestroy(() => {
		editor?.dispose();
		diffEditor?.dispose();
		originalModel?.dispose();
		modifiedModel?.dispose();
	});

	// Update value when props change
	$: if (editor && !isDiffMode) {
		if (editor.getValue() !== value) {
			editor.setValue(value);
		}
	}

	$: if (isDiffMode && modifiedModel) {
		if (modifiedModel.getValue() !== value) {
			modifiedModel.setValue(value);
		}
	}

	$: if (isDiffMode && originalModel && diffOriginal) {
		if (originalModel.getValue() !== diffOriginal) {
			originalModel.setValue(diffOriginal);
		}
	}

	$: if (monaco && editor) {
		editor.updateOptions({ readOnly, fontSize, lineHeight, tabSize });
	}
</script>

<div class="relative">
	<!-- Editor Header -->
	<div class="flex items-center justify-between p-3 glass-strong border-b border-cyber-border">
		<div class="flex items-center gap-2">
			<span class="text-xs text-neutral-400 font-mono">{fileName || 'Untitled'}</span>
			<span class="px-2 py-0.5 rounded text-[10px] font-medium bg-cyber-bgSecondary text-neutral-400">
				{language.toUpperCase()}
			</span>
		</div>

		<div class="flex items-center gap-1">
			{#if showDiff && diffOriginal}
				<button
					class="btn-ghost btn-sm p-2 {showDiff ? 'bg-neon-cyan/10 text-neon-cyan' : ''}"
					on:click={() => showDiff = !showDiff}
					title={showDiff ? 'Mostrar apenas código' : 'Mostrar diff'}
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
					</svg>
				</button>
			{/if}

			<button class="btn-ghost btn-sm p-2" on:click={copyContent} title="Copiar">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2h-2M12 19h2a2 2 0 002-2V7a2 2 0 00-2-2h-2v2a2 2 0 01-2 2H8a2 2 0 01-2-2V5a2 2 0 012-2h2" />
				</svg>
			</button>

			<button class="btn-ghost btn-sm p-2" on:click={downloadContent} title="Download">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
				</svg>
			</button>

			<button class="btn-ghost btn-sm p-2" on:click={() => { /* fullscreen */ }} title="Fullscreen">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
				</svg>
			</button>
		</div>
	</div>

	<!-- Editor Container -->
	<div
		bind:this={editorContainer}
		style="height: calc(100% - 50px);"
	/>


</div>