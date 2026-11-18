<script lang="ts">
	import { onMount } from 'svelte';
	import { challenges } from '$lib/stores';
	import { ChallengeList } from '$lib/components/challenges';

	async function loadChallenges() {
		await challenges.fetchChallenges();
	}

	onMount(() => {
		loadChallenges();
	});
</script>

<div class="space-y-6 animate-fade-in">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="font-display text-2xl font-bold text-neutral-100">Desafios</h1>
			<p class="text-neutral-400 mt-1">Explore, crie e jogue desafios de revisão de código</p>
		</div>

		<div class="flex flex-wrap items-center gap-3">
			<a href="/challenges/create" class="btn-primary">
				<Plus class="w-5 h-5 mr-2" />
				Criar Desafio
			</a>
			<button class="btn-secondary" on:click={() => challenges.fetchChallenges()}>
				<RefreshCw class="w-5 h-5 mr-2" />
				Atualizar
			</button>
		</div>
	</div>

	<!-- Challenge List -->
	<ChallengeList
		onPlay={(id) => { /* navigate to arena */ }}
		onView={(id) => { /* navigate to challenge detail */ }}
		onEdit={(id) => { /* navigate to edit */ }}
		onDelete={(id) => { /* delete challenge */ }}
		onCreate={() => { /* navigate to create */ }}
	/>
</div>

<script>
	import { Plus, RefreshCw } from 'lucide-svelte';
</script>