<script lang="ts">
	import { onMount } from 'svelte';
	import { auth, stats, apiConfig } from '$lib/stores';
	import { ProfileStats } from '$lib/components/profile';
	import { User, Settings, Trophy, ArrowRight, Loader2 } from 'lucide-svelte';

	let statsData: any = null;
	let loading = true;

	async function loadStats() {
		loading = true;
		try {
			await stats.fetchStats();
			statsData = $stats.stats;
		} catch (error) {
			console.error('Failed to load stats:', error);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		if ($auth.isAuthenticated) {
			loadStats();
		}
	});

	$: if ($auth.isAuthenticated && !$stats.stats) {
		loadStats();
	}

	import { Users } from 'lucide-svelte';
	import { formatRating } from '$lib/utils';
	import { getRatingTier } from '$lib/utils';
	import { getRatingTier as getTier } from '$lib/utils';
</script>

<div class="space-y-6 animate-fade-in">
	{#if !$auth.isAuthenticated}
		<!-- Not authenticated -->
		<div class="card text-center py-16">
			<User class="w-16 h-16 mx-auto mb-4 text-neon-cyan/50" />
			<h2 class="font-display text-2xl font-bold text-neutral-100 mb-2">Faça login para ver seu perfil</h2>
			<p class="text-neutral-400 mb-6">Acesse suas estatísticas, histórico de partidas e conquistas</p>
			<div class="flex items-center justify-center gap-4">
				<a href="/auth/login" class="btn-primary">
					Entrar
				</a>
				<a href="/auth/register" class="btn-secondary">
					Registrar
				</a>
			</div>
		</div>
	{:else if loading}
		<div class="card">
			<div class="flex items-center justify-center py-12">
				<Loader2 class="w-8 h-8 text-neon-cyan animate-spin" />
				<span class="ml-3 text-neutral-400">Carregando estatísticas...</span>
			</div>
		</div>
	{:else}
		<!-- Profile Header -->
		<div class="card relative overflow-hidden">
			<div class="absolute top-0 right-0 w-64 h-64 bg-neon-cyan/5 rounded-full blur-2xl" />
			<div class="relative p-6 md:p-8">
				<div class="flex flex-col md:flex-row md:items-center md:justify-between gap-6">
					<div class="flex items-center gap-4">
						<div class="w-24 h-24 rounded-2xl flex items-center justify-center font-display text-3xl font-bold text-cyber-bg"
							style="background: linear-gradient(135deg, var(--tw-color-neon-cyan), var(--tw-color-neon-magenta));">
							{$auth.user?.username?.charAt(0).toUpperCase()}
						</div>
						<div>
							<h1 class="font-display text-3xl font-bold text-neutral-100">{$auth.user?.username}</h1>
							<div class="flex items-center gap-4 mt-2">
								<span class="flex items-center gap-1 text-neon-cyan">
									<Trophy class="w-5 h-5" />
									<strong>{formatRating($auth.user?.rating || 0)}</strong>
								</span>
								<span class="flex items-center gap-1 text-neutral-400">
									<Users class="w-4 h-4" />
									{#if $stats.stats}#{$stats.stats.overall.total_matches}{:else}0{/if} partidas
								</span>
							</div>
						</div>
					</div>

					<div class="flex items-center gap-3">
						<a href="/profile/edit" class="btn-secondary">
							<Settings class="w-5 h-5 mr-2" />
							Editar Perfil
						</a>
					</div>
				</div>
			</div>
		</div>

		<!-- Rating Tier -->
		<div class="card">
			<div class="flex items-center justify-between">
				<div>
					<h3 class="font-display text-lg font-bold text-neutral-100">Tier Atual</h3>
					<p class="text-neutral-400 mt-1">Seu ranking atual no sistema</p>
				</div>
				<div class="flex items-center gap-4">
					<div class="w-16 h-16 rounded-2xl flex items-center justify-center font-display text-2xl font-bold text-cyber-bg"
						style="background: linear-gradient(135deg, {getRatingTier($auth.user?.rating || 0).color}, {getRatingTier($auth.user?.rating || 0).color}80);">
						{getRatingTier($auth.user?.rating || 0).name.charAt(0)}
					</div>
					<div>
						<p class="font-display text-2xl font-bold" style="color: {getRatingTier($auth.user?.rating || 0).color};">
							{getRatingTier($auth.user?.rating || 0).name}
						</p>
						<p class="text-neutral-400 text-sm">{$auth.user?.rating || 0} rating</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Detailed Stats -->
		<ProfileStats stats={$stats.stats} />
	{/if}
</div>
