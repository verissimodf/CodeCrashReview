<script lang="ts">
	import { onMount } from 'svelte';
	import { challenges, apiConfig } from '$lib/stores';
	import { ChallengeForm } from '$lib/components/challenges';
	import { goto } from '$app/navigation';
	import { Loader2, ArrowLeft, Save } from 'lucide-svelte';

	let challenge = {
		title: '',
		description: '',
		repository_url: '',
		repository_branch: 'main',
		target_paths: [''],
		exclude_paths: [''],
		severity: 'medium',
		category: 'security',
		time_limit_seconds: 300,
		max_turns: 10,
		acceptance_criteria: [],
		scoring_weights: {
			correctness: 0.4,
			security: 0.3,
			performance: 0.15,
			style: 0.1,
			test_coverage: 0.05
		},
		known_issues: [],
		tags: [''],
		template_id: '',
		is_public: true
	};

	let templates: any[] = [];
	let loading = false;
	let submitting = false;

	async function loadTemplates() {
		try {
			const config = get(apiConfig);
			const response = await fetch(`${config.api_url}/challenges/templates`, {
				headers: { Authorization: `Bearer ${localStorage.getItem('ccr_tokens') ? JSON.parse(localStorage.getItem('ccr_tokens')!).access_token : ''}` }
			});
			if (response.ok) {
				const data = await response.json();
				templates = data.data;
			}
		} catch (error) {
			console.error('Failed to load templates:', error);
		}
	}

	onMount(() => {
		loadTemplates();
	});

	async function handleSubmit(formChallenge: any) {
		submitting = true;
		try {
			const config = get(apiConfig);
			const response = await fetch(`${config.api_url}/challenges`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${localStorage.getItem('ccr_tokens') ? JSON.parse(localStorage.getItem('ccr_tokens')!).access_token : ''}`
				},
				body: JSON.stringify(formChallenge)
			});

			if (response.ok) {
				const data = await response.json();
				await challenges.fetchChallenges();
				await challenges.fetchTemplates();
				goto(`/challenges/${data.data.id}`);
			} else {
				const error = await response.json();
				alert(`Erro: ${error.error?.message || 'Falha ao criar desafio'}`);
			}
		} catch (error) {
			alert('Erro ao criar desafio');
		} finally {
			submitting = false;
		}
	}

	function handleCancel() {
		goTo('/challenges');
	}
</script>

<div class="max-w-4xl mx-auto">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<a href="/challenges" class="btn-ghost mb-4">
				<ArrowLeft class="w-5 h-5 mr-1" />
				Voltar
			</a>
			<h1 class="font-display text-2xl font-bold text-neutral-100">Criar Novo Desafio</h1>
			<p class="text-neutral-400 mt-1">Configure um novo desafio de revisão de código</p>
		</div>
	</div>

	<ChallengeForm
		bind:challenge
		{templates}
		onSubmit={handleSubmit}
		onCancel={() => goto('/challenges')}
		loading={submitting}
	/>
</div>

<script>
	import { ArrowLeft } from 'lucide-svelte';
</script>