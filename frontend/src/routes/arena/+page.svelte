<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { auth } from '$lib/auth';
    import { personas, reviewError, reviewRequest, reviewStatus, type ReviewSummary } from '$lib/reviews';

    let reviews: ReviewSummary[] = [];
    let sourceType: 'github'|'local' = 'github';
    let source = 'https://github.com/verissimodf/CodeCrashReview';
    let branch = 'main';
    let objective = 'Revisar o repositório e propor a correção mais importante considerando segurança, correção, confiabilidade, disponibilidade, manutenibilidade e eficiência.';
    let focusPaths = '';
    let reviewFocus = 'comprehensive';
    let validationProfile = 'go-test-v1';
    let validationDirectory = '.';
    let agentA = 'sentinel';
    let agentB = 'refactor';
    let evaluatorMode: 'ai'|'human' = 'ai';
    let busy = false;
    let error = '';
    const focusOptions = [
        ['comprehensive','Revisão ampla'], ['security','Segurança'], ['reliability','Confiabilidade'],
        ['refactoring','Refatoração'], ['performance','Desempenho'], ['architecture','Arquitetura'], ['documentation','Documentação']
    ];
    const specialists: Record<string,string[]> = {
        comprehensive: ['sentinel','optimizer','refactor','guardian','polyglot'],
        security: ['sentinel','guardian','polyglot'], reliability: ['guardian','sentinel','refactor'],
        refactoring: ['refactor','guardian','polyglot'], performance: ['optimizer','refactor','guardian'],
        architecture: ['refactor','docubot','guardian','polyglot'], documentation: ['docubot','refactor','polyglot']
    };
    $: compatiblePersonas = specialists[reviewFocus] || personas;
    $: if (!compatiblePersonas.includes(agentA)) agentA = compatiblePersonas[0] ?? 'sentinel';
    $: if (!compatiblePersonas.includes(agentB) || agentB === agentA) agentB = compatiblePersonas.find(value => value !== agentA) ?? compatiblePersonas[0] ?? 'guardian';

    async function load() {
        if (!$auth.isAuthenticated) return;
        try { reviews = await reviewRequest<ReviewSummary[]>(); }
        catch (e) { error = reviewError(e); }
    }

    async function create() {
        busy = true;
        error = '';
        try {
            const result = await reviewRequest<{id:string}>('', {
                source_type: sourceType, source, branch, objective, review_focus: reviewFocus, agent_a: agentA, agent_b: agentB,
                evaluator_mode: evaluatorMode,
                validation_profile: validationProfile, validation_directory: validationDirectory,
                focus_paths: focusPaths.split(/[\n,]/).map(value => value.trim()).filter(Boolean)
            });
            await goto('/arena/' + result.id);
        } catch (e) { error = reviewError(e); }
        finally { busy = false; }
    }
    onMount(load);
</script>

<svelte:head><title>Revisão autônoma · CodeCrashReview</title></svelte:head>
<h1>Revisões de código</h1>
<p>Dois agentes analisam o mesmo snapshot de forma independente. Um terceiro agente ou você escolhe a melhor proposta.</p>

{#if !$auth.isAuthenticated}
    <section class="panel"><p><a href="/auth/login">Entre na sua conta</a> para iniciar uma revisão.</p></section>
{:else}
    {#if error}<p role="alert">{error}</p>{/if}
    <section class="panel workspace-list">
        <div class="row"><h2>Suas revisões</h2><button on:click={load} disabled={busy}>Atualizar</button></div>
        {#each reviews as review}
            <a class="entry" href={'/arena/' + review.id}><strong>{review.source_location}</strong><span>{review.review_focus} · {review.agent_a} × {review.agent_b} · {reviewStatus[review.status] || review.status}</span></a>
        {:else}<div class="empty-state"><h3>Sua primeira revisão começa aqui.</h3><p>Indique um repositório, defina o objetivo e escolha dois especialistas. As propostas e suas provas ficarão reunidas nesta área.</p><a href="#new-review">Configurar uma revisão</a></div>{/each}
    </section>
    <section class="panel" id="new-review">
        <h2>Nova revisão</h2>
        <form on:submit|preventDefault={create}>
            <div class="row">
                <label>Origem<select bind:value={sourceType}><option value="github">GitHub público</option><option value="local">Repositório local autorizado</option></select></label>
                <label>Branch<input bind:value={branch} required maxlength="200" /></label>
            </div>
            <label>{sourceType === 'github' ? 'URL HTTPS do GitHub' : 'Caminho dentro de /repositories'}
                <input bind:value={source} required maxlength="2000" placeholder={sourceType === 'github' ? 'https://github.com/owner/repo' : 'meu-projeto'} />
            </label>
            {#if sourceType === 'local'}<p>Somente snapshots Git sob a raiz local montada como leitura são aceitos.</p>{/if}
            <label>Objetivo<textarea bind:value={objective} rows="5" maxlength="4000" required></textarea></label>
            <label>Especialidade da tarefa<select bind:value={reviewFocus}>{#each focusOptions as option}<option value={option[0]}>{option[1]}</option>{/each}</select></label>
            <p>Os perfis são filtrados pela especialidade da tarefa. Atualmente usam o mesmo modelo, com instruções independentes; essa seleção não equivale a uma certificação de competência.</p>
            <label>Testes de aceitação<select bind:value={validationProfile}><option value="go-test-v1">Go — testes do projeto</option><option value="python-unittest-v1">Python — unittest</option><option value="node-test-v1">JavaScript — testes do Node</option><option value="arena-grade-v1">Avaliador de pontuação da arena</option></select></label>
            <label>Diretório dos testes<input bind:value={validationDirectory} required placeholder=". ou backend" /></label>
            <p>Os testes originais serão preservados e executados antes e depois da correção. Uma proposta sem testes aprovados não pode ser publicada.</p>
            <label>Foco opcional — caminhos relativos separados por vírgula ou linha<textarea bind:value={focusPaths} rows="3" placeholder="backend/internal/auth, frontend/src"></textarea></label>
            <div class="row">
                <label>Agente A<select bind:value={agentA}>{#each compatiblePersonas as p}<option value={p}>{p}</option>{/each}</select></label>
                <label>Agente B<select bind:value={agentB}>{#each compatiblePersonas as p}<option value={p}>{p}</option>{/each}</select></label>
                <label>Avaliador<select bind:value={evaluatorMode}><option value="ai">Agente autônomo</option><option value="human">Operador humano</option></select></label>
            </div>
            <button class="action" disabled={busy || agentA === agentB}>{busy ? 'Enfileirando…' : 'Iniciar revisão'}</button>
            {#if agentA === agentB}<p role="alert">Escolha agentes diferentes.</p>{/if}
        </form>
    </section>

{/if}
