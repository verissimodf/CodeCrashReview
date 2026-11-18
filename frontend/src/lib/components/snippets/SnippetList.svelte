<script lang="ts">
	import { onMount } from 'svelte';
	import { cn } from '$lib/utils';
	import { SnippetCard } from './SnippetCard.svelte';
	import { Filter, Search, ChevronDown, Loader2, Download, Upload } from 'lucide-svelte';
	import { Dropdown } from '$lib/components/ui';

	export let snippets: any[] = [];
	export let loading = false;
	export let onInsert: (code: string, language: string) => void;
	export let onView: (id: string) => void;
	export let onImport: () => void;

	let searchQuery = '';
	let selectedLanguage = '';
	let selectedTag = '';
	let selectedComplexity = '';
	let sortBy = 'rating';
	let sortOrder = 'desc';

	const languages = [
		'cpp', 'python', 'java', 'javascript', 'typescript', 'rust', 'shell', 'html', 'gherkin'
	];

	const complexities = ['basic', 'intermediate', 'advanced'];

	const sortOptions = [
		{ value: 'rating', label: 'Maior avaliação' },
		{ value: 'usage_count', label: 'Mais usados' },
		{ value: 'created_at', label: 'Mais recentes' },
		{ value: 'title', label: 'Nome (A-Z)' }
	];

	let filteredSnippets = $derived(
		snippets
			.filter(s => {
				if (searchQuery && !s.title.toLowerCase().includes(searchQuery.toLowerCase()) &&
					!s.description.toLowerCase().includes(searchQuery.toLowerCase()) &&
					!s.code.toLowerCase().includes(searchQuery.toLowerCase())) {
					return false;
				}
				if (selectedLanguage && s.language !== selectedLanguage) return false;
				if (selectedComplexity && s.complexity !== selectedComplexity) return false;
				return true;
			})
			.sort((a, b) => {
				const aVal = a[sortBy];
				const bVal = b[sortBy];
				if (aVal < bVal) return sortOrder === 'asc' ? -1 : 1;
				if (aVal > bVal) return sortOrder === 'asc' ? 1 : -1;
				return 0;
			})
	);
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h2 class="font-display text-2xl font-bold text-neutral-100">Biblioteca de Snippets</h2>
			<p class="text-neutral-400 mt-1">{snippets.length} snippets disponíveis</p>
		</div>

		<div class="flex flex-wrap items-center gap-3">
			<button class="btn-primary" on:click={() => { /* import */ }}>
				<Upload class="w-5 h-5 mr-2" />
				Importar Snippets
			</button>
		</div>
	</div>

	<!-- Filters -->
	<div class="card">
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 p-4">
			<div class="lg:col-span-2">
				<label class="label">Buscar</label>
				<Input
					type="search"
					placeholder="Buscar por título, descrição, código..."
					bind:value={searchQuery}
				/>
			</div>

			<div>
				<label class="label">Linguagem</label>
				<Dropdown
					items={[
						{ label: 'Todas', value: '' },
						...[
							{ label: 'Python', value: 'python' },
							{ label: 'JavaScript', value: 'javascript' },
							{ label: 'TypeScript', value: 'typescript' },
							{ label: 'Java', value: 'java' },
							{ label: 'C++', value: 'cpp' },
							{ label: 'Rust', value: 'rust' },
							{ label: 'Go', value: 'go' },
							{ label: 'Shell', value: 'shell' },
							{ label: 'HTML', value: 'html' },
							{ label: 'Gherkin', value: 'gherkin' }
						].map(l => ({ label: l.label, value: l.value }))
					]}
					bind:value={selectedLanguage}
					placeholder="Todas"
				/>
			</div>

			<div>
				<label class="label">Complexidade</label>
				<Dropdown
					items={[
						{ label: 'Todas', value: '' },
						{ label: 'Básico', value: 'basic' },
						{ label: 'Intermediário', value: 'intermediate' },
						{ label: 'Avançado', value: 'advanced' }
					]}
					bind:value={selectedComplexity}
					placeholder="Todas"
				/>
			</div>

			<div>
				<label class="label">Ordenar</label>
				<Dropdown
					items={[
						{ label: 'Maior avaliação', value: 'rating' },
						{ label: 'Mais usados', value: 'usage_count' },
						{ label: 'Mais recentes', value: 'created_at' },
						{ label: 'Nome (A-Z)', value: 'title' }
					]}
					bind:value={sortBy}
					placeholder="Ordenar"
				/>
			</div>
		</div>
	</div>

	<!-- Snippets Grid -->
	{#if loading}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each Array(6) as _}
				<div class="card animate-pulse">
					<div class="h-6 bg-cyber-bgSecondary rounded w-3/4 mb-3" />
					<div class="h-4 bg-cyber-bgSecondary rounded w-full mb-2" />
					<div class="h-4 bg-cyber-bgSecondary rounded w-2/3 mb-2" />
					<div class="h-32 bg-cyber-bgSecondary rounded mb-4" />
					<div class="flex gap-2">
						<div class="h-5 bg-cyber-bgSecondary rounded w-20" />
						<div class="h-5 bg-cyber-bgSecondary rounded w-24" />
					</div>
					<div class="flex gap-2 mt-4">
						<div class="h-8 bg-cyber-bgSecondary rounded w-24" />
						<div class="h-8 bg-cyber-bgSecondary rounded w-28" />
					</div>
				</div>
			{/each}
		</div>
	{:else if snippets.length === 0}
		<div class="text-center py-16">
			<p class="text-lg font-medium text-neutral-100 mb-2">Nenhum snippet encontrado</p>
			<p class="text-neutral-400 mb-6">Tente ajustar os filtros ou importar snippets de repositórios upstream</p>
			<button class="btn-primary">Importar Snippets</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each snippets as snippet}
				<SnippetCard
					snippet={snippet}
					onInsert={onInsert}
					onView={onView}
				/>
			{/each}
		</div>
	{/if}