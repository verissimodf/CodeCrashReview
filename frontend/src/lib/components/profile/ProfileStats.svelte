<script lang="ts">
	import { cn } from '$lib/utils';
	import { formatRating, formatCompactNumber, getRatingTier, getAgentDisplayInfo } from '$lib/utils';
	import { Trophy, Medal, Users, TrendingUp, Target, Shield, Zap, Code2, CheckCircle2, BookOpen, Globe } from 'lucide-svelte';

	export let stats: any = null;

	const agentIcons: Record<string, any> = {
		sentinel: Shield,
		optimizer: Zap,
		refactor: Code2,
		guardian: CheckCircle2,
		docubot: BookOpen,
		polyglot: Globe
	};

	function getAgentColor(agent: string): string {
		const colors: Record<string, string> = {
			sentinel: '#ff3366',
			optimizer: '#00ff87',
			refactor: '#0099ff',
			guardian: '#ffaa00',
			docubot: '#ff00ff',
			polyglot: '#00ffff'
		};
		return colors[agent] || '#00ffff';
	}
</script>

{#if stats}
<div class="space-y-6">
	<!-- Overview Card -->
	<Card class="relative overflow-hidden">
		<div class="absolute top-0 right-0 w-48 h-48 bg-neon-cyan/5 rounded-full blur-2xl" />
		<div class="relative">
			<div class="flex items-center justify-between mb-6">
				<div>
					<h2 class="font-display text-xl font-bold text-neutral-100">Visão Geral</h2>
					<p class="text-neutral-400 mt-1">Estatísticas gerais do perfil</p>
				</div>
				<div class="w-16 h-16 rounded-2xl flex items-center justify-center font-display text-3xl font-bold text-cyber-bg"
					style="background: linear-gradient(135deg, var(--tw-color-neon-cyan), var(--tw-color-neon-magenta));">
					{getRatingTier(stats.overall.rating).name.charAt(0)}
				</div>
			</div>

			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<div class="p-4 bg-cyber-bgSecondary rounded-xl">
					<p class="text-sm text-neutral-400 mb-1">Rating Atual</p>
					<p class="font-display text-3xl font-bold text-neon-cyan tabular-nums">{formatRating(stats.overall.rating)}</p>
					<p class="text-xs text-neutral-500 mt-1">Pico: {formatRating(stats.overall.peak_rating)}</p>
				</div>

				<div class="p-4 bg-cyber-bgSecondary rounded-xl">
					<p class="text-sm text-neutral-400 mb-1">Vitórias</p>
					<p class="font-display text-3xl font-bold text-neon-green tabular-nums">{formatCompactNumber(stats.overall.wins)}</p>
					<p class="text-xs text-neutral-500 mt-1">{stats.overall.win_rate.toFixed(1)}% taxa</p>
				</div>

				<div class="p-4 bg-cyber-bgSecondary rounded-xl">
					<p class="text-sm text-neutral-400 mb-1">Partidas</p>
					<p class="font-display text-3xl font-bold text-neon-amber tabular-nums">{formatCompactNumber(stats.overall.total_matches)}</p>
					<p class="text-xs text-neutral-500 mt-1">Derrotas: {formatCompactNumber(stats.overall.losses)}</p>
				</div>

				<div class="p-4 bg-cyber-bgSecondary rounded-xl">
					<p class="text-sm text-neutral-400 mb-1">Taxa Vitória</p>
					<p class="font-display text-3xl font-bold text-neon-magenta">{stats.overall.win_rate.toFixed(1)}%</p>
					<p class="text-xs text-neutral-500 mt-1">RD: {stats.overall.rating_deviation.toFixed(1)}</p>
				</div>
			</div>
		</div>
	</Card>

	<!-- Agent Stats -->
	<Card class="space-y-4">
		<h3 class="font-display text-lg font-bold text-neutral-100">Desempenho por Agente</h3>

		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each Object.entries(stats.by_agent) as [agent, data]}
				<div class="p-4 bg-cyber-bgSecondary rounded-xl border border-cyber-border/50 hover:border-neon-cyan/30 transition-colors">
					<div class="flex items-center gap-3 mb-3">
						<div class="w-12 h-12 rounded-xl flex items-center justify-center font-display font-bold text-cyber-bg"
							style="background: linear-gradient(135deg, {getAgentColor(agent)}, {getAgentColor(agent)}80);">
							<svelte:component this={getAgentDisplayInfo(agent).icon} class="w-6 h-6" />
						</div>
						<div>
							<p class="font-display text-lg font-bold text-neutral-100">{getAgentDisplayInfo(agent).name}</p>
							<p class="text-xs text-neutral-500">{getAgentDisplayInfo(agent).description}</p>
						</div>
					</div>

					<div class="grid grid-cols-3 gap-3 text-center">
						<div class="p-2 bg-cyber-bg rounded-lg">
							<p class="font-display text-xl font-bold text-neon-green">{data.win_rate.toFixed(1)}%</p>
							<p class="text-[10px] text-neutral-500">Vitórias</p>
						</div>
						<div class="p-2 bg-cyber-bg rounded-lg">
							<p class="font-display text-xl font-bold text-neon-cyan">{data.average_score.toFixed(1)}</p>
							<p class="text-[10px] text-neutral-500">Score Médio</p>
						</div>
						<div class="p-2 bg-cyber-bg rounded-lg">
							<p class="font-display text-xl font-bold text-neon-amber">{data.total_matches}</p>
							<p class="text-[10px] text-neutral-500">Partidas</p>
						</div>
					</div>

					<div class="mt-3 pt-3 border-t border-cyber-border text-xs text-neutral-500 space-y-1">
						<p>Forte contra: <span class="text-neutral-400">{data.strongest_against ? getAgentDisplayInfo(data.strongest_against).name : 'N/A'}</span></p>
						<p>Fraco contra: <span class="text-neutral-400">{data.weakest_against ? getAgentDisplayInfo(data.weakest_against).name : 'N/A'}</span></p>
					</div>
				</div>
			{/each}
		</div>
	</Card>

	<!-- Language & Category Stats -->
	<div class="grid gap-6 md:grid-cols-2">
		<Card class="space-y-4">
			<h3 class="font-display text-lg font-bold text-neutral-100">Por Linguagem</h3>
			<div class="space-y-3">
				{#each Object.entries(stats.by_language) as [lang, data]}
					<div class="flex items-center justify-between p-3 bg-cyber-bgSecondary rounded-lg">
						<div class="flex items-center gap-3">
							<span class="w-8 h-8 rounded-lg flex items-center justify-center font-display text-xs text-cyber-bg" style="background: linear-gradient(135deg, var(--tw-color-neon-cyan), var(--tw-color-neon-magenta));">
								{lang.charAt(0).toUpperCase()}
							</span>
							<div>
								<p class="font-medium text-neutral-100">{lang}</p>
								<p class="text-xs text-neutral-500">{data.matches} partidas</p>
							</div>
						</div>
						<div class="text-right">
							<p class="font-bold text-neon-green">{data.win_rate.toFixed(1)}%</p>
							<p class="text-xs text-neutral-500">{data.wins}W / {data.losses}L</p>
						</p>
						</div>
					</div>
				{/each}
			</div>
		</Card>

		<Card class="space-y-4">
			<h3 class="font-display text-lg font-bold text-neutral-100">Por Categoria</h3>
			<div class="space-y-3">
				{#each Object.entries(stats.by_category) as [cat, data]}
					<div class="flex items-center justify-between p-3 bg-cyber-bgSecondary rounded-lg">
						<div class="flex items-center gap-3">
							<svelte:component this={getCategoryIcon(cat)} class="w-5 h-5" style="color: var(--tw-color-{getSeverityColor('high')})" />
							<div>
								<p class="font-medium text-neutral-100 capitalize">{cat}</p>
								<p class="text-xs text-neutral-500">{data.matches} partidas</p>
							</div>
						</div>
						<div class="text-right">
							<p class="font-bold text-neon-green">{data.win_rate.toFixed(1)}%</p>
							<p class="text-xs text-neutral-500">{data.wins}W / {data.losses}L</p>
						</div>
					</div>
				{/each}
			</div>
		</Card>
	</div>

	<!-- Recent Matches & Streaks -->
	<div class="grid gap-6 md:grid-cols-2">
		<Card class="space-y-4">
			<div class="flex items-center justify-between">
				<h3 class="font-display text-lg font-bold text-neutral-100">Partidas Recentes</h3>
			</div>

			<div class="space-y-3">
				{#each stats.recent_matches.slice(0, 10) as match}
					<div class="flex items-center justify-between p-3 bg-cyber-bgSecondary rounded-lg">
						<div class="flex items-center gap-3">
							<div class="w-8 h-8 rounded-lg flex items-center justify-center font-display font-bold text-cyber-bg"
								style="background: linear-gradient(135deg, {getAgentColor(match.my_agent)}, {getAgentColor(match.my_agent)}80);">
								{match.my_agent.charAt(0).toUpperCase()}
							</div>
							<div>
								<p class="font-medium text-neutral-100 truncate max-w-[200px]">{match.challenge_title}</p>
								<p class="text-xs text-neutral-400">{match.opponent} • {match.my_agent} vs {match.opponent_agent}</p>
							</div>
						</div>

						<div class="flex items-center gap-4 text-sm">
							<span class={cn(
								'px-2 py-1 rounded-full text-xs font-bold',
								match.result === 'win' && 'bg-neon-green/20 text-neon-green',
								match.result === 'loss' && 'bg-neon-red/20 text-neon-red',
								match.result === 'draw' && 'bg-neon-amber/20 text-neon-amber'
							)}>
								{match.result === 'win' ? 'Vitória' : match.result === 'loss' ? 'Derrota' : 'Empate'}
							</span>
							<span class="font-mono tabular-nums {match.rating_change > 0 ? 'text-neon-green' : 'text-neon-red'}">
								{match.rating_change > 0 ? '+' : ''}{match.rating_change}
							</span>
						</div>
					</div>
				{/each}

				{#if stats.recent_matches.length === 0}
					<div class="text-center py-8 text-neutral-400">
						<p>Nenhuma partida jogada ainda</p>
					</div>
				{/if}
			</div>
		</Card>

		<Card class="space-y-4">
			<h3 class="font-display text-lg font-bold text-neutral-100">Streaks & Conquistas</h3>

			<div class="grid grid-cols-3 gap-4">
				<div class="p-4 bg-cyber-bgSecondary rounded-xl text-center">
					<p class="font-display text-3xl font-bold text-neon-green">{stats.streaks.current > 0 ? '+' : ''}{stats.streaks.current}</p>
					<p class="text-xs text-neutral-500 mt-1">Streak Atual</p>
				</div>
				<div class="p-4 bg-cyber-bgSecondary rounded-xl text-center">
					<p class="font-display text-3xl font-bold text-neon-amber">{stats.streaks.longest_win}</p>
					<p class="text-xs text-neutral-500 mt-1">Maior Streak Vitórias</p>
				</div>
				<div class="p-4 bg-cyber-bgSecondary rounded-xl text-center">
					<p class="font-display text-3xl font-bold text-neon-red">{stats.streaks.longest_loss}</p>
					<p class="text-xs text-neutral-500 mt-1">Maior Streak Derrotas</p>
				</div>
			</div>

			<div class="pt-4 border-t border-cyber-border">
				<h4 class="font-medium text-neutral-100 mb-3">Conquistas</h4>
				<div class="flex flex-wrap gap-3">
					{#each stats.achievements as achievement}
						<div class="flex items-center gap-2 p-2 bg-cyber-bgSecondary rounded-lg {achievement.unlocked_at ? '' : 'opacity-50'}">
							<div class="w-8 h-8 rounded-lg flex items-center justify-center text-cyber-bg"
								style="background: linear-gradient(135deg, {achievement.color}, {achievement.color}80);">
								<svelte:component this={achievement.icon} class="w-4 h-4" />
							</div>
							<div>
								<p class="font-medium text-sm text-neutral-100">{achievement.name}</p>
								<p class="text-xs text-neutral-400">{achievement.description}</p>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</Card>
	</div>
{/if}