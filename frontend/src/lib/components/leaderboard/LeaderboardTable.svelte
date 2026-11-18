<script lang="ts">
	import { cn } from '$lib/utils';
	import { formatRating, getRatingTier, getAgentDisplayInfo, getLanguageDisplayName } from '$lib/utils';
	import { Trophy, Medal, TrendingUp, ChevronUp, ChevronDown, User, Code, Shield, Zap, CheckCircle2, BookOpen, Globe } from 'lucide-svelte';

	export let entries: any[] = [];
	export let currentUserId: string | null = null;
	export let type: 'global' | 'weekly' | 'language' | 'agent' = 'global';
	export let language: string;
	export let agent: string;

	const tierIcons: Record<string, any> = {
		Grandmaster: Trophy,
		Master: Medal,
		Diamond: TrendingUp,
		Platinum: Trophy,
		Gold: Medal,
		Silver: Trophy,
		Bronze: Shield,
		Unranked: User
	};

	const agentIcons: Record<string, any> = {
		sentinel: Shield,
		optimizer: Zap,
		refactor: Code,
		guardian: CheckCircle2,
		docubot: BookOpen,
		polyglot: Globe
	};

	function getTrend(current: number, previous: number): 'up' | 'down' | 'stable' {
		if (current > previous) return 'up';
		if (current < previous) return 'down';
		return 'stable';
	}

	function getTrendIcon(trend: string) {
		switch (trend) {
			case 'up': return ChevronUp;
			case 'down': return ChevronDown;
			default: return null;
		}
	}
</script>

<div class="overflow-x-auto">
	<table class="w-full border-collapse">
		<thead>
			<tr class="bg-cyber-bgTerritary border-b border-cyber-border">
				<th class="px-4 py-3 text-left text-xs font-semibold text-neutral-500 uppercase tracking-wider">#</th>
				<th class="px-4 py-3 text-left text-xs font-semibold text-neutral-500 uppercase tracking-wider">Jogador</th>
				<th class="px-4 py-3 text-right text-xs font-semibold text-neutral-500 uppercase tracking-wider">Rating</th>
				<th class="px-4 py-3 text-right text-xs font-semibold text-neutral-500 uppercase tracking-wider">RD</th>
				<th class="px-4 py-3 text-right text-xs font-semibold text-neutral-500 uppercase tracking-wider">Partidas</th>
				<th class="px-4 py-3 text-right text-xs font-semibold text-neutral-500 uppercase tracking-wider">Vitórias</th>
				<th class="px-4 py-3 text-right text-xs font-semibold text-neutral-500 uppercase tracking-wider">Taxa</th>
				<th class="px-4 py-3 text-right text-xs font-semibold text-neutral-500 uppercase tracking-wider">Agente</th>
				<th class="px-4 py-3 text-right text-xs font-semibold text-neutral-500 uppercase tracking-wider">Linguagem</th>
				<th class="px-4 py-3 text-right text-xs font-semibold text-neutral-500 uppercase tracking-wider">Tendência</th>
			</tr>
		</thead>
		<tbody class="divide-y divide-cyber-border">
			{#each entries as entry, index}
				<tr class={cn(
					'hover:bg-cyber-bgTerritary/50 transition-colors',
					entry.user_id === currentUserId && 'bg-neon-cyan/5 ring-1 ring-inset ring-neon-cyan/20'
				)}>
					<td class="px-4 py-3 font-display font-bold text-lg {index < 3 ? 'text-neon-amber' : 'text-neutral-400'}">
						#{entry.rank}
					</td>

					<td class="px-4 py-3">
						<div class="flex items-center gap-3">
							<div class="w-8 h-8 rounded-full bg-neon-cyan/10 border border-neon-cyan/30 flex items-center justify-center">
								{#if entry.avatar_url}
									<img src={entry.avatar_url} alt="" class="w-full h-full rounded-full" />
								{:else}
									<span class="font-display text-neon-cyan font-bold">{entry.username.charAt(0).toUpperCase()}</span>
								{/if}
							</div>
							<div>
								<p class="font-medium text-neutral-100 {entry.user_id === currentUserId ? 'text-neon-cyan' : ''}">{entry.username}</p>
								{#if entry.user_id === currentUserId}
									<span class="text-xs text-neon-cyan">(Você)</span>
								{/if}
							</div>
						</div>
					</td>

					<td class="px-4 py-3 text-right font-display font-bold text-neutral-100 tabular-nums">
						{formatRating(entry.rating)}
					</td>

					<td class="px-4 py-3 text-right text-neutral-400 tabular-nums">
						{entry.rating_deviation.toFixed(1)}
					</td>

					<td class="px-4 py-3 text-right text-neutral-400 tabular-nums">
						{entry.total_matches}
					</td>

					<td class="px-4 py-3 text-right text-neutral-400 tabular-nums">
						{entry.wins}
					</td>

					<td class="px-4 py-3 text-right text-neutral-400 tabular-nums">
						{entry.win_rate.toFixed(1)}%
					</td>

					<td class="px-4 py-3 text-right">
						{#if entry.preferred_agent}
							<div class="inline-flex items-center justify-end gap-2">
								<span class="w-6 h-6 rounded-lg flex items-center justify-center font-display font-bold text-cyber-bg" style="background: linear-gradient(135deg, {getAgentDisplayInfo(entry.preferred_agent).color}, {getAgentDisplayInfo(entry.preferred_agent).color}80);">
									<svelte:component this={agentIcons[entry.preferred_agent]} class="w-4 h-4" />
								</div>
								<span class="text-sm text-neutral-400 hidden sm:inline">{getAgentDisplayInfo(entry.preferred_agent).name}</span>
							</div>
						{:else}
							<span class="text-neutral-500 text-sm">-</span>
						{/if}
					</td>

					<td class="px-4 py-3 text-right">
						{#if entry.preferred_language}
							<span class="text-sm text-neutral-400 px-2 py-0.5 rounded bg-cyber-bgTerritary">
								{getLanguageDisplayName(entry.preferred_language)}
							</span>
						{:else}
							<span class="text-neutral-500 text-sm">-</span>
						{/if}
					</td>

					<td class="px-4 py-3 text-right">
						{#if entry.trend !== 'stable'}
							<div class="inline-flex items-center justify-end gap-1">
								<svelte:component this={getTrendIcon(entry.trend)} class="w-4 h-4" style="color: {entry.trend === 'up' ? 'var(--tw-color-neon-green)' : 'var(--tw-color-neon-red)'}" />
								<span class="text-sm font-medium" style="color: {entry.trend === 'up' ? 'var(--tw-color-neon-green)' : 'var(--tw-color-neon-red)'}">
									{entry.trend === 'up' ? '↑' : '↓'}
								</span>
							</div>
						{:else}
							<span class="text-neutral-500 text-sm">→</span>
						{/if}
					</td>
				</tr>
			{/each}

			{#if entries.length === 0}
				<tr>
					<td colspan="10" class="px-4 py-12 text-center text-neutral-400">
						<p class="text-lg font-medium mb-2">Nenhum jogador encontrado</p>
						<p class="text-sm">Jogue sua primeira partida para aparecer no ranking!</p>
					</td>
				</tr>
			{/if}
		</tbody>
	</table>
</div>