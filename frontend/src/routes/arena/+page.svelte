<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { page } from '$app/stores';
    import { auth } from '$lib/stores';
    import { request, personas, statusLabel, errorText, type ArenaChallenge, type ArenaMatch } from '$lib/arena';

    let challenges: ArenaChallenge[] = [];
    let matches: ArenaMatch[] = [];
    let selected = '';
    let opponent = 'guardian';
    let persona = personas[0];
    let error = '';
    let busy = false;
    async function load() {
        busy = true;
        error = '';
        try {
            [challenges, matches] = await Promise.all([
                request<ArenaChallenge[]>('arena/challenges'), request<ArenaMatch[]>('arena/matches')
            ]);
            const desired = $page.url.searchParams.get('challenge') || selected;
            selected = challenges.find(c => c.id === desired)?.id || challenges[0]?.id || '';
        } catch (e) { error = errorText(e); }
        finally { busy = false; }
    }
    async function create(duel = false) {
        busy = true;
        error = '';
        try {
            const match = await request<{id:string}>(duel ? 'arena/duels' : 'arena/matches', {challenge_id: selected, persona, ...(duel ? {opponent_persona: opponent} : {})});
            await goto('/arena/' + match.id);
        } catch (e) { error = errorText(e); }
        finally { busy = false; }
    }
    onMount(() => { if ($auth.isAuthenticated) load(); });
</script>

<svelte:head><title>Arena · CodeCrashReview</title></svelte:head>
<h1>Arena</h1>
{#if !$auth.isAuthenticated}
    <p><a href="/auth/login">Entre na sua conta</a> para criar ou participar de uma partida.</p>
{:else}
    {#if error}<p role="alert">{error}</p>{/if}
    <section class="panel">
        <h2>Nova partida</h2>
        <form on:submit|preventDefault={() => create()}>
            <label>Desafio<select bind:value={selected}>{#each challenges as challenge}<option value={challenge.id}>{challenge.title}</option>{/each}</select></label>
            <p>{challenges.find(c => c.id === selected)?.description || ''}</p>
            <label>Agente<select bind:value={persona}>{#each personas as p}<option value={p}>{p}</option>{/each}</select></label>
            <button class="action" disabled={busy || !selected}>Criar partida</button>
            <label>Agente adversário<select bind:value={opponent}>{#each personas as p}<option value={p}>{p}</option>{/each}</select></label>
            <button type="button" disabled={busy || !selected || persona === opponent} on:click={() => create(true)}>Iniciar duelo entre agentes</button>
            {#if busy}<p role="status">Aguardando o servidor…</p>{/if}
        </form>
    </section>
    <section class="panel">
        <div class="row"><h2>Partidas</h2><button on:click={load} disabled={busy}>Atualizar</button></div>
        {#each matches as match}
            <a class="entry" href={'/arena/' + match.id}>
                <strong>{challenges.find(c => c.id === match.challenge_id)?.title || match.challenge_id}</strong>
                <span>{statusLabel[match.status] || match.status} · {match.id.slice(0,8)}</span>
            </a>
        {:else}
            {#if !error}<p>{busy ? 'Carregando…' : 'Nenhuma partida encontrada.'}</p>{/if}
        {/each}
    </section>
{/if}
