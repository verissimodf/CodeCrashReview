<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { auth } from '$lib/stores';
    import { request, personas, statusLabel, errorText, type ArenaMatch, type ArenaChallenge } from '$lib/arena';

    let match: ArenaMatch | null = null;
    let challenge: ArenaChallenge | undefined;
    let code = '';
    let persona = 'guardian';
    let error = '';
    let pollError = '';
    let busy = false;
    let initialized = false;
    let refreshing = false;
    let disposed = false;
    $: matchPath = 'arena/matches/' + encodeURIComponent($page.params.matchId || '');
    $: me = match?.participants?.find(p => p.username === $auth.user?.username);
    $: submitted = match?.submissions?.some(s => s.username === $auth.user?.username);
    $: winner = match?.participants?.find(p => p.user_id === match?.winner_id);
    async function load() {
        if (refreshing || disposed) return;
        refreshing = true;
        try {
            const current = await request<ArenaMatch>(matchPath);
            if (disposed) return;
            match = current;
            if (!initialized) {
                const list = await request<ArenaChallenge[]>('arena/challenges');
                if (disposed) return;
                challenge = list.find(c => c.id === current.challenge_id);
                code = challenge?.starter_code || '';
                initialized = true;
            }
            pollError = '';
        } catch (e) { pollError = errorText(e); }
        finally { refreshing = false; }
    }
    async function act(action: 'join'|'suggest'|'submit') {
        busy = true;
        error = '';
        try {
            if (action === 'suggest') {
                const result = await request<{code:string}>(matchPath + '/suggest', {});
                code = result.code;
            } else {
                await request(matchPath + '/' + action, action === 'join' ? {persona} : {code});
            }
            await load();
        } catch (e) { error = errorText(e); }
        finally { busy = false; }
    }
    onMount(() => {
        if (!$auth.isAuthenticated) return;
        load();
        const interval = setInterval(() => {
            if (!busy && !document.hidden && match?.status !== 'completed') load();
        }, 2500);
        return () => { disposed = true; clearInterval(interval); };
    });
</script>
<svelte:head><title>Partida · CodeCrashReview</title></svelte:head>
<a href="/arena">← Arena</a>
<h1>{challenge?.title || 'Partida'}</h1>
{#if !$auth.isAuthenticated}
    <p><a href="/auth/login">Entre na sua conta</a> para acompanhar a partida.</p>
{:else}
    {#if error || pollError}<p role="alert">{error || pollError}</p>{/if}
    {#if match}
        <section class="panel">
            <div class="row"><h2>{statusLabel[match.status] || match.status}</h2><button disabled={busy} on:click={load}>Atualizar</button></div>
            <p>{challenge?.description || ''}</p>
            {#each match.participants || [] as player}
                <div class="entry"><strong>{player.username}</strong><span>{player.persona} · {statusLabel[player.status] || player.status}{player.score == null ? '' : ' · ' + player.score + ' pts'}</span></div>
            {/each}
            {#if match.status === 'completed'}
                <h3>{match.winner_id ? 'Vencedor: ' + (winner?.username || match.winner_id) : 'Empate'}</h3>
            {:else if !me && match.participants?.length < 2}
                <label>Agente<select bind:value={persona}>{#each personas as p}<option value={p}>{p}</option>{/each}</select></label>
                <button class="action" on:click={() => act('join')} disabled={busy}>Entrar na partida</button>
            {/if}
        </section>
        {#if me && !submitted && match.status !== 'completed'}
            <section class="panel"><h2>Sua solução</h2>
                <label>Código<textarea bind:value={code} rows="12" spellcheck="false" disabled={busy}></textarea></label>
                <div class="row"><button on:click={() => act('suggest')} disabled={busy}>Gerar sugestão do agente</button>
                    <button class="action" on:click={() => act('submit')} disabled={busy || !code.trim()}>Enviar solução</button></div>
                {#if busy}<p role="status">Aguardando o servidor…</p>{/if}
            </section>
        {/if}
        {#if match.submissions?.length}
            <section class="panel"><h2>Histórico de entregas</h2>
                {#each match.submissions as submission}
                    <article><h3>{submission.username} · {submission.score} pts</h3><p>{submission.feedback}</p>
                        {#each (Array.isArray(submission.files) ? submission.files.map(f => [f.path, f.content]) : Object.entries(submission.files || {})) as [file, content]}
                            <details><summary>{file}</summary><pre>{typeof content === 'string' ? content : JSON.stringify(content, null, 2)}</pre></details>
                        {/each}
                    </article>
                {/each}
            </section>
        {/if}
    {:else if !pollError}<p>Carregando partida…</p>{/if}
{/if}
