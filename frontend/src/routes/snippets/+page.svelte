<script lang="ts">
	import { onMount } from 'svelte';
	import { snippets } from '$lib/stores';
	import { SnippetList } from '$lib/components/snippets';
	import { BookOpen, Download, Upload, RefreshCw } from 'lucide-svelte';

	async function loadSnippets() {
		await snippets.fetchSnippets();
	}

	onMount(() => {
		loadSnippets();
	});

	function handleInsert(code: string, language: string) {
		// This would integrate with the Monaco editor
		console.log('Insert snippet:', code, language);
	}
</script>

<div class="space-y-6 animate-fade-in">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="font-display text-2xl font-bold text-neutral-100">Biblioteca de Snippets</h1>
			<p class="text-neutral-400 mt-1">Algoritmos, padrões e trechos de código prontos para uso</p>
		</div>

		<div class="flex flex-wrap items-center gap-3">
			<button class="btn-primary" on:click={() => snippets.importSnippets('awesome-algorithms')}>
				<Upload class="w-5 h-5 mr-2" />
				Importar de Repositórios
			</button>
			<button class="btn-secondary" on:click={() => snippets.fetchSnippets()}>
				<RefreshCw class="w-5 h-5 mr-2" />
				Atualizar
			</button>
		</div>
	</div>

	<!-- Snippet List -->
	<SnippetList
		onInsert={(code, language) => { /* insert into editor */ }}
		onView={(id) => { /* navigate to detail */ }}
	/>
</div>