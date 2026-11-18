<script lang="ts">
	import { onMount } from 'svelte';
	import { cn } from '$lib/utils';
	import { ChallengeCard } from './ChallengeCard.svelte';
	import { Filter, Search, ChevronDown, Plus, Loader2 } from 'lucide-svelte';
	import { Dropdown } from '$lib/components/ui';

	export let challenges: any[] = [];
	export let loading = false;
	export let onPlay: (id: string) => void;
	export let onView: (id: string) => void;
	export let onEdit: (id: string) => void;
	export let onDelete: (id: string) => void;
	export let onCreate: () => void;

	let searchQuery = '';
	let selectedCategory = '';
	let selectedSeverity = '';
	let sortBy = 'created_at';
	let sortOrder = 'desc';
	let showFilters = false;

	const categories = [
		'security',
		'performance',
		'refactor',
		'testing',
		'documentation',
		'migration',
		'bugfix',
		'feature'
	];

	const severities = ['low', 'medium', 'high', 'critical'];

	const sortOptions = [
		{ value: 'created_at', label: 'Mais recentes' },
		{ value: 'updated_at', label: 'Atualizados recentemente' },
		{ value: 'play_count', label: 'Mais jogados' },
		{ value: 'average_rating', label: 'Melhor avaliados' },
		{ value: 'title', label: 'Nome (A-Z)' }
	];

	let filteredChallenges = $derived(
		challenges
			.filter(c => {
				if (searchQuery && !c.title.toLowerCase().includes(searchQuery.toLowerCase()) &&
					!c.description.toLowerCase().includes(searchQuery.toLowerCase())) {
					return false;
				}
				if (selectedCategory && c.category !== selectedCategory) return false;
				if (selectedSeverity && c.severity !== selectedSeverity) return false;
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

	function clearFilters() {
		searchQuery = '';
		selectedCategory = '';
		selectedSeverity = '';
	}
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h2 class="font-display text-2xl font-bold text-neutral-100">Desafios</h2>
			<p class="text-neutral-400 mt-1">{filteredChallenges.length} de {challenges.length} desafios</p>
		</div>

		<div class="flex flex-wrap items-center gap-3">
			<button class="btn-primary" on:click={onCreate}>
				<Plus class="w-5 h-5 mr-2" />
				Novo Desafio
			</button>

			<button class="btn-secondary {showFilters ? 'bg-neon-cyan/10 text-neon-cyan border-neon-cyan/30' : ''}" on:click={() => showFilters = !showFilters}>
				<Filter class="w-5 h-5 mr-2" />
				Filtros
				{/* Active filter indicator */}
				{#if selectedCategory || selectedSeverity || searchQuery}
					<span class="ml-1 px-1.5 py-0.5 rounded-full text-[10px] font-medium bg-neon-cyan/20 text-neon-cyan">
						{(selectedCategory ? 1 : 0) + (selectedSeverity ? 1 : 0) + (searchQuery ? 1 : 0)}
					</span>
				{/if}
			</button>
		</div>
	</div>

	<!-- Filters Panel -->
	{#if showFilters}
		<div class="card animate-slide-down">
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
				<div>
					<label class="label">Buscar</label>
					<Input
						type="search"
						placeholder="Buscar por título, descrição..."
						bind:value={searchQuery}
					/>
				</div>

				<div>
					<label class="label">Categoria</label>
					<Dropdown
						items={[
							{ label: 'Todas', value: '' },
							...categories.map(c => ({ label: c.charAt(0).toUpperCase() + c.slice(1), value: c }))
						]}
						bind:value={selectedCategory}
						placeholder="Todas"
					/>
				</div>

				<div>
					<label class="label">Severidade</label>
					<Dropdown
						items={[
							{ label: 'Todas', value: '' },
							...severities.map(s => ({ label: s.charAt(0).toUpperCase() + s.slice(1), value: s }))
						]}
						bind:value={selectedSeverity}
						placeholder="Todas"
					/>
				</div>

				<div>
					<label class="label">Ordenar por</label>
					<Dropdown
						items={sortOptions.map(o => ({ label: o.label, value: o.value }))}
						bind:value={sortBy}
						placeholder="Ordenar"
					/>
				</div>
			</div>

			<div class="flex items-center justify-end gap-2 pt-4 border-t border-cyber-border">
				<button class="btn-ghost btn-sm" on:click={clearFilters}>
					Limpar filtros
				</button>
			</div>
		</div>
	{/if}

	<!-- Challenges Grid -->
	{#if loading}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each Array(6) as _}
				<div class="card animate-pulse">
					<div class="h-6 bg-cyber-bgSecondary rounded w-3/4 mb-4" />
					<div class="h-4 bg-cyber-bgSecondary rounded w-full mb-2" />
					<div class="h-4 bg-cyber-bgSecondary rounded w-2/3 mb-4" />
					<div class="flex gap-2">
						<div class="h-5 bg-cyber-bgSecondary rounded w-20" />
						<div class="h-5 bg-cyber-bgSecondary rounded w-24" />
					</div>
					<div class="h-4 bg-cyber-bgSecondary rounded w-full mt-4" />
					<div class="flex gap-2 mt-4">
						<div class="h-8 bg-cyber-bgSecondary rounded w-24" />
						<div class="h-8 bg-cyber-bgSecondary rounded w-28" />
					</div>
				</div>
			{/each}
		</div>
	{:else if filteredChallenges.length === 0}
		<div class="text-center py-16">
			<Search class="w-16 h-16 mx-auto mb-4 text-neutral-600" />
			<h3 class="font-display text-xl font-bold text-neutral-100 mb-2">Nenhum desafio encontrado</h3>
			<p class="text-neutral-400 mb-6">Tente ajustar os filtros ou criar um novo desafio</p>
			<button class="btn-primary" on:click={onCreate}>Criar primeiro desafio</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each filteredChallenges as challenge}
				<ChallengeCard
					challenge={challenge}
					onPlay={onPlay}
					onView={onView}
					onEdit={onEdit}
					onDelete={onDelete}
				/>
			{/each}
		</div>
	{/if}
</div>