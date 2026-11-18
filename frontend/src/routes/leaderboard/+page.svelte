<script lang="ts">
	import { onMount } from 'svelte';
	import { leaderboard, stats } from '$lib/stores';
	import { LeaderboardTable, LeaderboardCharts } from '$lib/components/leaderboard';
	import { Trophy, TrendingUp, Globe, Users, RefreshCw } from 'lucide-svelte';

	let activeTab = 'global';
	let loading = true;

	const tabs = [
		{ value: 'global', label: 'Global', icon: Trophy },
		{ value: 'weekly', label: 'Semanal', icon: TrendingUp },
		{ value: 'language', label: 'Por Linguagem', icon: Code },
		{ value: 'agent', label: 'Por Agente', icon: Users }
	];

	async function loadLeaderboard() {
		loading = true;
		await leaderboard.fetchLeaderboard(activeTab as any, {
			language: activeTab === 'language' ? 'python' : undefined,
			agent: activeTab === 'agent' ? 'sentinel' : undefined
		});
		loading = false;
	}

	onMount(() => {
		loadLeaderboard();
		stats.fetchStats();
	});

	function handleTabChange(tab: string) {
		activeTab = tab;
		loadLeaderboard();
	}
</script>

<div class="space-y-6 animate-fade-in">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="font-display text-2xl font-bold text-neutral-100">Ranking</h1>
			<p class="text-neutral-400 mt-1">Veja os melhores jogadores e suas estatísticas</p>
		</div>

		<div class="flex items-center gap-3">
			<button class="btn-secondary" on:click={() => leaderboard.refreshAll()}>
				<RefreshCw class="w-5 h-5 mr-2" />
				Atualizar
			</button>
		</div>
	</div>

	<!-- Tabs -->
	<div class="tabs mb-6">
		{#each tabs as tab}
			<button
				class={tab.value === activeTab ? 'tab-active' : 'tab'}
				on:click={() => handleTabChange(tab.value)}
			>
				<svelte:component this={tab.icon} class="w-4 h-4 mr-1" />
				{tab.label}
			</button>
		{/each}
	</div>

	<!-- Leaderboard Table -->
	<div class="card">
		{#if loading}
			<div class="space-y-4">
				{#each Array(5) as _}
					<div class="h-12 bg-cyber-bgSecondary rounded-lg animate-pulse" />
				{/each}
			</div>
		{:else}
			<LeaderboardTable
				entries={activeTab === 'global' ? $leaderboard.global : activeTab === 'weekly' ? $leaderboard.weekly : activeTab === 'language' ? $leaderboard.byLanguage.python || [] : $leaderboard.byAgent.sentinel || []}
				currentUserId={$stats.stats?.overall?.user_id}
				type={activeTab}
				language="python"
				agent="sentinel"
			/>
		{/if}
	</div>

	<!-- Charts (only for global) -->
	{#if activeTab === 'global' && $stats.stats}
		<LeaderboardCharts
			userStats={$stats.stats}
			globalStats={{}}
		/>
	{/if}
</div>

<script>
	import { Code } from 'lucide-svelte';

	function handleTabChange(tab: string) {
		activeTab = tab;
		loadLeaderboard();
	}
</script>