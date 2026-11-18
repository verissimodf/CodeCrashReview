<script lang="ts">
	import { cn } from '$lib/utils';
	import { getSeverityColor, getCategoryIcon, formatRelativeTime } from '$lib/utils';
	import { Shield, Zap, Code2, CheckCircle2, BookOpen, GitBranch, Bug, Sparkles, Clock, ArrowRight, ExternalLink } from 'lucide-svelte';
	import { Badge } from '$lib/components/ui';

	export let challenge: {
		id: string;
		title: string;
		description: string;
		repository_url: string;
		repository_branch: string;
		target_paths: string[];
		exclude_paths: string[];
		severity: string;
		category: string;
		time_limit_seconds: number;
		max_turns: number;
		acceptance_criteria: any[];
		scoring_weights: Record<string, number>;
		known_issues: any[];
		tags: string[];
		template_id?: string;
		is_public: boolean;
		created_by: string;
		play_count: number;
		average_rating: number;
		created_at: string;
		updated_at: string;
	};

	export let onPlay: (id: string) => void;
	export let onView: (id: string) => void;
	export let onEdit: (id: string) => void;
	export let onDelete: (id: string) => void;

	const categoryIcons: Record<string, any> = {
		security: Shield,
		performance: Zap,
		refactor: Code2,
		testing: CheckCircle2,
		documentation: BookOpen,
		migration: GitBranch,
		bugfix: Bug,
		feature: Sparkles
	};

	const CategoryIcon = categoryIcons[challenge.category] || Code2;
	const severityColor = getSeverityColor(challenge.severity);
</script>

<div class="card-neon group relative overflow-hidden" style="--neon-color: var(--tw-color-{severityColor}); --neon-color-glow: var(--tw-color-{severityColor}Glow);">
	<div class="absolute top-0 right-0 w-24 h-24 bg-{severityColor}/5 rounded-full blur-2xl opacity-0 group-hover:opacity-100 transition-opacity" />

	<div class="relative">
		<div class="flex items-start justify-between mb-4">
			<div class="flex-1 min-w-0">
				<div class="flex items-center gap-2 mb-2">
					<Badge variant="neon" color={severityColor} glowColor={severityColor}Glow size="sm">
						{challenge.severity.toUpperCase()}
					</Badge>
					<Badge variant="outline" size="sm">
						<svelte:component this={CategoryIcon} class="w-3 h-3 mr-1" />
						{challenge.category}
					</Badge>
					{#if challenge.template_id}
						<Badge variant="default" size="sm">
							Template
						</Badge>
					{/if}
				</div>

				<h3 class="font-display text-xl font-bold text-neutral-100 truncate mb-1">
					{challenge.title}
				</h3>
				<p class="text-neutral-400 text-sm line-clamp-2">{challenge.description}</p>
			</div>

			<div class="flex items-center gap-2 ml-4">
				<Badge variant="outline" size="sm" class="hidden sm:inline-flex">
					<Clock class="w-3 h-3 mr-1" />
					{Math.floor(challenge.time_limit_seconds / 60)}min
				</Badge>
				<Badge variant="outline" size="sm" class="hidden sm:inline-flex">
					<Code class="w-3 h-3 mr-1" />
					{challenge.max_turns} turnos
				</Badge>
			</div>
		</div>

		<!-- Repository Info -->
		<div class="flex items-center gap-2 text-sm text-neutral-500 mb-4">
			<ExternalLink class="w-4 h-4" />
			<a href={challenge.repository_url} target="_blank" rel="noopener noreferrer" class="truncate max-w-[200px] hover:text-neon-cyan transition-colors">
				{challenge.repository_url.replace('https://', '').replace('http://', '')}
			</a>
			<span>•</span>
			<span>Branch: {challenge.repository_branch}</span>
			<span>•</span>
			<span>{challenge.target_paths.length} paths</span>
		</div>

		<!-- Tags -->
		<div class="flex flex-wrap gap-2 mb-4">
			{#each challenge.tags.slice(0, 5) as tag}
				<Badge variant="outline" size="sm">{tag}</Badge>
			{/each}
			{#if challenge.tags.length > 5}
				<Badge variant="default" size="sm">+{challenge.tags.length - 5}</Badge>
			{/if}
		</div>

		<!-- Stats -->
		<div class="flex items-center justify-between pt-4 border-t border-cyber-border">
			<div class="flex items-center gap-4 text-sm text-neutral-400">
				<span class="flex items-center gap-1">
					<Clock class="w-4 h-4" />
					{formatRelativeTime(challenge.created_at)}
				</span>
				<span class="flex items-center gap-1">
					<Code class="w-4 h-4" />
					{challenge.play_count} partidas
				</span>
				<span class="flex items-center gap-1">
					<Star class="w-4 h-4" />
					{challenge.average_rating.toFixed(1)}
				</span>
			</div>

			<div class="flex items-center gap-2">
				<button class="btn-secondary btn-sm" on:click={() => onPlay?.(challenge.id)}>
					Jogar
				</button>
				<button class="btn-ghost btn-sm" on:click={() => onView?.(challenge.id)}>
					Ver detalhes
				</button>
			</div>
		</div>
	</div>
</div>

<script>
	import { Star } from 'lucide-svelte';
</script>