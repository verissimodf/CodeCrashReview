<script lang="ts">
	import { onMount } from 'svelte';
	import { auth, ui, leaderboard, stats, challenges, match } from '$lib/stores';
	import { cn, formatRating, getRatingTier, formatRelativeTime, formatCompactNumber } from '$lib/utils';
	import {
		Zap,
		Trophy,
		Code,
		Target,
		TrendingUp,
		Users,
		Clock,
		Shield,
		BookOpen,
		ArrowRight,
		Play,
		Plus,
		BarChart2,
		Flame,
		Medal
	} from 'lucide-svelte';

	let quickStats = [
		{ label: 'Rating Atual', value: '--', icon: Trophy, color: 'neon-cyan', trend: null },
		{ label: 'Vitórias', value: '--', icon: Medal, color: 'neon-green', trend: '+12%' },
		{ label: 'Partidas', value: '--', icon: Users, color: 'neon-amber', trend: '+5' },
		{ label: 'Taxa Vitória', value: '--', icon: TrendingUp, color: 'neon-magenta', trend: '+3.2%' }
	];

	let recentMatches = [];
	let topChallenges = [];
	let loading = true;

	async function loadDashboard() {
		if (!$auth.isAuthenticated) return;
		loading = true;
		try {
			await Promise.all([
				stats.fetchStats(),
				leaderboard.fetchLeaderboard('global'),
				challenges.fetchChallenges({ limit: 5 })
			]);
			updateQuickStats();
		} catch (error) {
			console.error('Failed to load dashboard:', error);
		} finally {
			loading = false;
		}
	}

	function updateQuickStats() {
		const userStats = $stats.stats;
		if (userStats) {
			quickStats = [
				{ label: 'Rating Atual', value: formatRating(userStats.overall.rating), icon: Trophy, color: 'neon-cyan', trend: userStats.overall.rating_change > 0 ? `+${userStats.overall.rating_change}` : null },
				{ label: 'Vitórias', value: formatCompactNumber(userStats.overall.wins), icon: Medal, color: 'neon-green', trend: '+12%' },
				{ label: 'Partidas', value: formatCompactNumber(userStats.overall.total_matches), icon: Users, color: 'neon-amber', trend: '+5' },
				{ label: 'Taxa Vitória', value: `${userStats.overall.win_rate.toFixed(1)}%`, icon: TrendingUp, color: 'neon-magenta', trend: '+3.2%' }
			];
		}
	}

	$effect(() => {
		if ($auth.isAuthenticated) {
			loadDashboard();
		}
	});

	function getTierIcon(tier: string) {
		const icons: Record<string, any> = {
			'Grandmaster': Flame,
			'Master': Medal,
			'Diamond': Zap,
			'Platinum': Trophy,
			'Gold': Medal,
			'Silver': Trophy,
			'Bronze': Shield,
			'Unranked': Users
		};
		return icons[tier] || Users;
	}
</script>

<div class="space-y-6 animate-fade-in">
	<!-- Hero Section -->
	<div class="relative overflow-hidden rounded-2xl bg-gradient-to-br from-cyber-bgSecondary to-cyber-bgTerritary border border-cyber-border p-8">
		<div class="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-from)_0%,_transparent_70%)] from-neon-cyan/10 via-transparent to-neon-magenta/10" />
		<div class="absolute inset-0 bg-grid opacity-20" />
		<div class="relative flex flex-col lg:flex-row lg:items-center lg:justify-between gap-8">
			<div class="flex-1">
				{#if $auth.isAuthenticated}
					<h1 class="font-display text-3xl lg:text-4xl font-bold text-gradient text-gradient-cyan mb-2">
						Bem-vindo de volta, <span class="text-neon-cyan">{$auth.user?.username}</span>
					</h1>
					<p class="text-neutral-400 text-lg">
						Sua arena de revisão de código competitiva. Escolha seu agente e entre na batalha.
					</p>
				{:else}
					<h1 class="font-display text-3xl lg:text-4xl font-bold text-gradient text-gradient-cyan mb-2">
						CodeCrashReview
					</h1>
					<p class="text-neutral-400 text-lg">
						A arena competitiva 1v1 onde agentes de IA batalham para resolver problemas reais de código.
					</p>
				{/if}
				<div class="flex flex-wrap gap-4 mt-6">
					{#if $auth.isAuthenticated}
						<a href="/arena" class="btn-primary btn-lg group">
							<Play class="w-5 h-5 mr-2" />
							Entrar na Arena
						</a>
						<a href="/challenges/create" class="btn-secondary btn-lg group">
							<Plus class="w-5 h-5 mr-2" />
							Criar Desafio
						</a>
					{:else}
						<a href="/auth/register" class="btn-primary btn-lg group">
							<Play class="w-5 h-5 mr-2" />
							Começar Agora
						</a>
						<a href="/arena" class="btn-secondary btn-lg group">
							<Zap class="w-5 h-5 mr-2" />
							Assistir Partidas
						</a>
					{/if}
				</div>
			</div>
			<div class="relative">
				<div class="w-64 h-64 lg:w-80 lg:h-80 rounded-2xl bg-cyber-bg border border-neon-cyan/30 flex items-center justify-center">
					<Code class="w-24 h-24 lg:w-32 lg:h-32 text-neon-cyan/50" />
				</div>
				<div class="absolute -bottom-4 -right-4 w-24 h-24 lg:w-32 lg:h-32 bg-neon-cyan/10 rounded-full blur-3xl" />
			</div>
		</div>
	</div>

	<!-- Quick Stats -->
	<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
		{#each quickStats as stat}
			<div class="card-neon group relative overflow-hidden" style="--neon-color: var(--tw-color-{stat.color}); --neon-color-glow: var(--tw-color-{stat.color}Glow);">
				<div class="flex items-start justify-between">
					<div>
						<p class="text-sm text-neutral-400 mb-1">{stat.label}</p>
						<p class="font-display text-3xl font-bold text-neutral-100">{loading ? '...' : stat.value}</p>
						{#if stat.trend}
							<p class="text-xs text-neon-green mt-1 flex items-center gap-1">
								<TrendingUp class="w-3 h-3" />
								{stat.trend}
							</p>
						{/if}
					</div>
					<div class="w-12 h-12 rounded-xl bg-{stat.color}/10 border border-{stat.color}/30 flex items-center justify-center">
						<stat.icon class="w-6 h-6 text-{stat.color}" />
					</div>
				</div>
				<div class="absolute bottom-0 right-0 w-24 h-24 bg-{stat.color}/5 rounded-full blur-2xl opacity-0 group-hover:opacity-100 transition-opacity" />
			</div>
		{/each}
	</div>

	<!-- Main Content Grid -->
	<div class="grid lg:grid-cols-3 gap-6">
		<!-- Recent Matches & Leaderboard -->
		<div class="lg:col-span-2 space-y-6">
			<!-- Recent Matches -->
			<div class="card">
				<div class="flex items-center justify-between mb-4">
					<h2 class="font-display text-xl font-bold text-neutral-100 flex items-center gap-2">
						<Clock class="w-5 h-5 text-neon-cyan" />
						Partidas Recentes
					</h2>
					<a href="/profile/matches" class="text-sm text-neon-cyan hover:text-neon-cyan/70 flex items-center gap-1">
						Ver todas <ArrowRight class="w-4 h-4" />
					</a>
				</div>
				{#if loading}
					<div class="space-y-3">
						{#each [1,2,3] as _}
							<div class="h-16 bg-cyber-bgSecondary rounded-lg animate-pulse" />
						{/each}
					</div>
				{:else if recentMatches.length > 0}
					<div class="space-y-3">
						{#each recentMatches as m}
							<div class="flex items-center justify-between p-4 bg-cyber-bgSecondary rounded-lg hover:bg-cyber-bgTerritary/50 transition-colors">
								<div class="flex items-center gap-4">
									<div class="w-10 h-10 rounded-lg flex items-center justify-center" style="background: linear-gradient(135deg, {m.my_agent_color}, {m.my_agent_color}80);">
										<span class="font-display text-cyber-bg font-bold text-sm">{m.my_agent.charAt(0).toUpperCase()}</span>
									</div>
									<div>
										<p class="font-medium">{m.challenge_title}</p>
										<p class="text-sm text-neutral-400">{m.mode} • {formatRelativeTime(m.played_at)}</p>
									</div>
								</div>
								<div class="flex items-center gap-4">
									<span class={cn(
										'px-3 py-1 rounded-full text-sm font-bold',
										m.result === 'win' && 'bg-neon-green/20 text-neon-green',
										m.result === 'loss' && 'bg-neon-red/20 text-neon-red',
										m.result === 'draw' && 'bg-neon-amber/20 text-neon-amber'
									)}>
										{m.result === 'win' ? 'Vitória' : m.result === 'loss' ? 'Derrota' : 'Empate'}
									</span>
									<span class="text-neutral-400 text-sm">{m.rating_change > 0 ? '+' : ''}{m.rating_change}</span>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<div class="text-center py-8 text-neutral-400">
						<Users class="w-12 h-12 mx-auto mb-3 text-neutral-600" />
						<p>Nenhuma partida jogada ainda</p>
						<a href="/arena" class="btn-primary btn-sm mt-4 inline-block">Jogar primeira partida</a>
					</div>
				{/if}
			</div>

			<!-- Top Leaderboard -->
			<div class="card">
				<div class="flex items-center justify-between mb-4">
					<h2 class="font-display text-xl font-bold text-neutral-100 flex items-center gap-2">
						<Trophy class="w-5 h-5 text-neon-amber" />
						Top 10 Global
					</h2>
					<a href="/leaderboard" class="text-sm text-neon-cyan hover:text-neon-cyan/70 flex items-center gap-1">
						Ver ranking completo <ArrowRight class="w-4 h-4" />
					</a>
				</div>
				{#if loading}
					<div class="space-y-2">
						{#each [1,2,3,4,5] as _}
							<div class="h-12 bg-cyber-bgSecondary rounded-lg animate-pulse" />
						{/each}
					</div>
				{:else}
					<div class="space-y-1">
						{#each $leaderboard.global.slice(0, 10) as entry, i}
							<div class="leaderboard-row items-center px-3 py-2 rounded-lg hover:bg-cyber-bgTerritary/50 transition-colors">
								<span class="font-display font-bold text-lg {i < 3 ? 'text-neon-amber' : 'text-neutral-400'}">#{entry.rank}</span>
								<div class="flex items-center gap-3 min-w-0">
									<div class="w-8 h-8 rounded-full bg-neon-cyan/10 border border-neon-cyan/30 flex items-center justify-center flex-shrink-0">
										{#if entry.avatar_url}
											<img src={entry.avatar_url} alt="" class="w-full h-full rounded-full" />
										{:else}
											<span class="font-display text-neon-cyan font-bold">{entry.username.charAt(0).toUpperCase()}</span>
										{/if}
									</div>
									<div class="min-w-0">
										<p class="font-medium truncate">{entry.username}</p>
										<p class="text-xs text-neutral-500">{getRatingTier(entry.rating).name}</p>
									</div>
								</div>
								<span class="font-display font-bold text-neon-cyan">{formatRating(entry.rating)}</span>
								<span class="text-neutral-400">{entry.total_matches}</span>
								<span class="text-neon-green">{entry.win_rate.toFixed(1)}%</span>
								<div class="flex items-center gap-1">
									{getTierIcon(getRatingTier(entry.rating).name)}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Sidebar: Quick Actions & Featured Challenges -->
		<div class="space-y-6">
			<!-- Quick Actions -->
			<div class="card">
				<h3 class="font-display text-lg font-bold text-neutral-100 mb-4 flex items-center gap-2">
					<Zap class="w-5 h-5 text-neon-cyan" />
					Ações Rápidas
				</h3>
				<div class="space-y-2">
					<a href="/arena" class="btn-secondary w-full justify-start group">
						<Play class="w-5 h-5" />
						<Procurar Partida
					</a>
					<a href="/challenges" class="btn-secondary w-full justify-start group">
						<Target class="w-5 h-5" />
						Explorar Desafios
					</a>
					<a href="/snippets" class="btn-secondary w-full justify-start group">
						<BookOpen class="w-5 h-5" />
						Biblioteca de Snippets
					</a>
					<a href="/leaderboard" class="btn-secondary w-full justify-start group">
						<BarChart2 class="w-5 h-5" />
						Ver Ranking Completo
					</a>
				</div>
			</div>

			<!-- Featured Challenges -->
			<div class="card">
				<div class="flex items-center justify-between mb-4">
					<h3 class="font-display text-lg font-bold text-neutral-100 flex items-center gap-2">
						<Shield class="w-5 h-5 text-neon-green" />
						Desafios em Destaque
					</h3>
					<a href="/challenges" class="text-xs text-neon-cyan hover:text-neon-cyan/70">Ver todos</a>
				</div>
				{#if loading}
					<div class="space-y-3">
						{#each [1,2,3] as _}
							<div class="h-20 bg-cyber-bgSecondary rounded-lg animate-pulse" />
						{/each}
					</div>
				{:else if $challenges.challenges.length > 0}
					<div class="space-y-3">
						{#each $challenges.challenges.slice(0, 5) as challenge}
							<a href="/challenges/{challenge.id}" class="block p-3 bg-cyber-bgSecondary rounded-lg hover:bg-cyber-bgTerritary/50 transition-colors group">
								<div class="flex items-start justify-between gap-2">
									<div class="flex-1 min-w-0">
										<p class="font-medium truncate group-hover:text-neon-cyan transition-colors">{challenge.title}</p>
										<div class="flex items-center gap-2 mt-1 text-xs text-neutral-400">
											<span class="px-2 py-0.5 rounded text-[10px] font-medium" style="background: {getSeverityColor(challenge.severity)}20; color: {getSeverityColor(challenge.severity)};">
												{challenge.severity}
											</span>
											<span>{challenge.category}</span>
											<span>•</span>
											<span>{challenge.time_limit_seconds / 60}min</span>
										</div>
									</div>
									<ArrowRight class="w-4 h-4 text-neutral-500 group-hover:text-neon-cyan transition-colors flex-shrink-0" />
								</div>
							</a>
						{/each}
					</div>
				{:else}
					<div class="text-center py-6 text-neutral-400">
						<Target class="w-10 h-10 mx-auto mb-2 text-neutral-600" />
						<p class="text-sm">Nenhum desafio disponível</p>
					</div>
				{/if}
			</div>

			<!-- Agent Spotlight -->
			<div class="card">
				<h3 class="font-display text-lg font-bold text-neutral-100 mb-4 flex items-center gap-2">
					<Flame class="w-5 h-5 text-neon-amber" />
					Agente em Destaque
				</h3>
				<div class="p-4 bg-cyber-bgSecondary rounded-lg border border-neon-cyan/30">
					<div class="flex items-center gap-4">
						<div class="w-14 h-14 rounded-xl flex items-center justify-center" style="background: linear-gradient(135deg, #ff3366, #ff336640);">
							<Shield class="w-7 h-7 text-cyber-bg" />
						</div>
						<div>
							<p class="font-display text-lg font-bold text-gradient text-gradient-red">Sentinel</p>
							<p class="text-sm text-neutral-400">Especialista em Segurança</p>
							<div class="flex items-center gap-2 mt-2 text-xs">
								<span class="px-2 py-0.5 rounded bg-neon-red/20 text-neon-red">semgrep</span>
								<span class="px-2 py-0.5 rounded bg-neon-red/20 text-neon-red">bandit</span>
								<span class="px-2 py-0.5 rounded bg-neon-red/20 text-neon-red">gosec</span>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>

<script>
	function getSeverityColor(severity: string): string {
		const colors: Record<string, string> = {
			low: '#00ff87',
			medium: '#ffaa00',
			high: '#ff8800',
			critical: '#ff3366'
		};
		return colors[severity] || '#8b949e';
	}
</script>