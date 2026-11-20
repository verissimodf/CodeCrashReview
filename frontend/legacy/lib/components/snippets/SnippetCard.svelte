<script lang="ts">
	import { cn } from '$lib/utils';
	import { getLanguageDisplayName, formatCompactNumber } from '$lib/utils';
	import { Copy, Download, Star, ExternalLink, ChevronDown, ChevronUp } from 'lucide-svelte';
	import { Badge, Dropdown } from '$lib/components/ui';

	export let snippet: {
		id: string;
		title: string;
		description: string;
		language: string;
		code: string;
		tags: string[];
		complexity: string;
		author: string;
		author_id?: string;
		source_url?: string;
		rating: number;
		usage_count: number;
		is_official: boolean;
		created_at: string;
		updated_at: string;
	};

	export let onInsert: (code: string, language: string) => void;
	export let onView: (id: string) => void;
	export let onRate: (id: string, rating: number) => void;

	let showFullCode = false;
	let showActions = false;
	let userRating = 0;
	let hoverRating = 0;

	const complexityColors: Record<string, string> = {
		basic: 'neon-green',
		intermediate: 'neon-amber',
		advanced: 'neon-red'
	};

	const complexityColor = complexityColors[snippet.complexity] || 'neon-cyan';
</script>

<div class="card-neon group relative overflow-hidden" style="--neon-color: var(--tw-color-{complexityColor}); --neon-color-glow: var(--tw-color-{complexityColor}Glow);">
	<div class="absolute top-0 right-0 w-24 h-24 bg-{complexityColor}/5 rounded-full blur-2xl opacity-0 group-hover:opacity-100 transition-opacity" />

	<div class="relative">
		<!-- Header -->
		<div class="flex items-start justify-between mb-3">
			<div class="flex-1 min-w-0">
				<div class="flex items-center gap-2 mb-2">
					<Badge variant="neon" color={complexityColor} glowColor={complexityColor}Glow size="sm">
						{getLanguageDisplayName(snippet.language)}
					</Badge>
					<Badge variant="outline" size="sm" style="--badge-color: var(--tw-color-{complexityColor}); --badge-color-glow: var(--tw-color-{complexityColor}Glow);">
						{snippet.complexity.charAt(0).toUpperCase() + snippet.complexity.slice(1)}
					</Badge>
					{#if snippet.is_official}
						<Badge variant="neon" color="neon-amber" glowColor="neon-amberGlow" size="sm">
							Oficial
						</Badge>
					{/if}
				</div>

				<h3 class="font-display text-lg font-bold text-neutral-100 truncate mb-1">{snippet.title}</h3>
				<p class="text-neutral-400 text-sm line-clamp-2">{snippet.description}</p>
			</div>

			<div class="flex items-center gap-1 ml-4">
				<button class="btn-ghost btn-sm p-2" on:click={() => onView?.(snippet.id)} title="Ver detalhes">
					<ExternalLink class="w-4 h-4" />
				</button>
			</div>
		</div>

		<!-- Tags -->
		<div class="flex flex-wrap gap-1.5 mb-3">
			{#each snippet.tags.slice(0, 4) as tag}
				<Badge variant="outline" size="sm">{tag}</Badge>
			{/each}
			{#if snippet.tags.length > 4}
				<Badge variant="default" size="sm">+{snippet.tags.length - 4}</Badge>
			{/if}
		</div>

		<!-- Code Preview -->
		<div class="relative mb-3">
			<div class="font-mono text-xs bg-cyber-bgSecondary rounded-lg p-3 overflow-x-auto max-h-40">
				<pre class="whitespace-pre-wrap text-neutral-300">{showFullCode ? snippet.code : snippet.code.slice(0, 500) + (snippet.code.length > 500 ? '...' : '')}</pre>
			</div>

			{#if snippet.code.length > 500}
				<button
					class="absolute bottom-0 left-0 right-0 h-10 bg-gradient-to-t from-cyber-bgSecondary to-transparent flex items-center justify-center"
					on:click={() => showFullCode = !showFullCode}
				>
					<span class="text-xs text-neutral-400 px-3 py-1 rounded bg-cyber-bgSecondary border border-cyber-border">
						{showFullCode ? 'Mostrar menos' : 'Mostrar mais'}
						<ChevronUp class="w-3 h-3 ml-1" />
					</span>
				</button>
			{/if}
		</div>

		<!-- Stats -->
		<div class="flex items-center justify-between pt-3 border-t border-cyber-border">
			<div class="flex items-center gap-4 text-sm text-neutral-400">
				<span class="flex items-center gap-1">
					<Star class="w-4 h-4 text-neon-amber" />
					{snippet.rating.toFixed(1)}
				</span>
				<span class="flex items-center gap-1">
					<Download class="w-4 h-4" />
					{formatCompactNumber(snippet.usage_count)}
				</span>
			</div>

			<div class="flex items-center gap-2">
				<Dropdown
					items={[
						{ label: 'Inserir no Editor', value: 'insert', icon: Download },
						{ label: 'Copiar Código', value: 'copy', icon: Copy },
						{ label: 'Download', value: 'download', icon: Download },
						{ label: 'Ver Detalhes', value: 'view', icon: ExternalLink }
					]}
					placeholder="Ações"
					disabled={false}
				/>
			</div>
		</div>
	</div>
</div>