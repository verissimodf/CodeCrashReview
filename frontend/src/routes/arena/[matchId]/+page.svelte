<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { auth } from '$lib/auth';
    import { reviewError, reviewRequest, reviewStatus, type ReviewDetail, type ReviewProposal } from '$lib/reviews';

    let review: ReviewDetail|null = null;
    let error = '';
    let busy = false;
    let disposed = false;
    let rationale = '';
    const slots: ('A'|'B')[] = ['A', 'B'];
    $: id = $page.params.matchId || '';

    async function load() {
        if (!$auth.isAuthenticated || disposed) return;
        try { review = await reviewRequest<ReviewDetail>('/' + encodeURIComponent(id)); error = ''; }
        catch (e) { error = reviewError(e); }
    }
    async function verdict(winner: 'A'|'B'|'TIE') {
        busy = true; error = '';
        try { await reviewRequest('/' + encodeURIComponent(id) + '/verdict', {winner, rationale, scores: {}}); await load(); }
        catch (e) { error = reviewError(e); }
        finally { busy = false; }
    }
    function proposal(slot: 'A'|'B'): ReviewProposal|undefined { return review?.proposals?.find(item => item.slot === slot); }
    function downloadWinner() {
        if (!review?.verdict || review.verdict.winner === 'TIE') return;
        const winner = proposal(review.verdict.winner);
        if (!winner?.patch_valid || !winner.verification?.proof?.passed || !winner.patch || !review.snapshot_sha) return;
        const url = URL.createObjectURL(new Blob([winner.patch.trimEnd() + '\n'], {type: 'text/x-diff'}));
        const link = document.createElement('a');
        link.href = url;
        link.download = `review-${review.id}-${review.snapshot_sha.slice(0,12)}.patch`;
        link.click();
        setTimeout(() => URL.revokeObjectURL(url), 1000);
    }
    onMount(() => {
        load();
        const timer = setInterval(() => { if (review && !['completed','failed','awaiting_human'].includes(review.status) && !document.hidden) load(); }, 2500);
        return () => { disposed = true; clearInterval(timer); };
    });
</script>

<svelte:head><title>Revisão · CodeCrashReview</title></svelte:head>
<a href="/arena">← Revisões</a>
<h1>Revisão do repositório</h1>
{#if error}<p role="alert">{error}</p>{/if}
{#if !$auth.isAuthenticated}<p><a href="/auth/login">Entre para abrir esta revisão.</a></p>
{:else if !review}<p>Carregando…</p>
{:else}
    <section class="panel">
        <div class="row"><h2>{reviewStatus[review.status] || review.status}</h2><button on:click={load}>Atualizar</button></div>
        <p><strong>Origem:</strong> {review.source_location} · branch {review.source_branch}</p>
        <p><strong>Objetivo:</strong> {review.objective}</p>
        <p><strong>Especialidade:</strong> {review.review_focus}</p>
        {#if review.snapshot_sha}<p><strong>Snapshot:</strong> <code>{review.snapshot_sha.slice(0,12)}</code> · {review.snapshot_files} arquivos · {review.snapshot_bytes} bytes analisados</p>{/if}
        {#if review.error_message}<p role="alert">{review.error_message}</p>{/if}
        {#if !['completed','failed','awaiting_human'].includes(review.status)}<p role="status">O processamento continua no servidor; esta página pode ser fechada.</p>{/if}
    </section>

    <div class="review-grid">
        {#each slots as slot}
            {@const item = proposal(slot)}
            <section class="panel">
                <h2>Proposta {slot} · {item?.persona || (slot === 'A' ? review.agent_a : review.agent_b)}</h2>
                {#if !item || item.status === 'running'}<p>Agente analisando…</p>
                {:else if item.status === 'failed'}<p role="alert">{item.error_message || 'O agente falhou.'}</p>
                {:else}
                    <p>{item.summary}</p>
                    <p><strong>Patch:</strong> {item.patch_valid ? 'aplica limpo ao snapshot' : 'inválido — ' + (item.validation_error || 'sem validação')}</p>
                    <p><strong>Confiança:</strong> {Math.round(item.confidence * 100)}% · <strong>Latência:</strong> {item.latency_ms} ms</p>
                    {#if item.verification}
                        <div class="evidence-box">
                            <h3>Evidência verificada pelo sistema</h3>
                            <p><strong>Estado:</strong> {item.verification.proof ? (item.verification.proof.passed ? 'testes funcionais aprovados no perfil ' + item.verification.proof.profile : 'compilação ou testes reprovados') : 'validação funcional pendente'}</p>
                            {#if item.verification.proof}<details><summary>Prova executada e identificação dos artefatos</summary><p>Código SHA-256: <code>{item.verification.proof.source_sha256}</code></p><p>Testes SHA-256: <code>{item.verification.proof.tests_sha256}</code></p><p>Imagem: <code>{item.verification.proof.image}</code></p><p>Saída: {item.verification.proof.exit_code}</p><pre>{item.verification.proof.output}</pre></details>{/if}
                            <ul>{#each item.verification.system_evidence || [] as evidence}<li>{evidence}</li>{/each}</ul>
                            {#if item.verification.unsupported_claims?.length}
                                <details class="claim-warning"><summary>Alegações do agente sem prova independente ({item.verification.unsupported_claims.length})</summary><ul>{#each item.verification.unsupported_claims as claim}<li>{claim}</li>{/each}</ul></details>
                            {/if}
                        </div>
                    {/if}
                    {#if item.delivery}
                        <h3>Arquitetura e entrega</h3>
                        <p><strong>Topologia alvo:</strong> {item.delivery.topology}</p>
                        <p><strong>Design:</strong> {item.delivery.design}</p>
                        <p><strong>Objetivo final alcançável:</strong> {item.delivery.final_objective}</p>
                        <h4>Metodologia utilizada</h4><ol>{#each item.delivery.methodology || [] as step}<li>{step}</li>{/each}</ol>
                        <h4>Plano de implantação</h4><ol>{#each item.delivery.deployment_plan || [] as step}<li>{step}</li>{/each}</ol>
                        <h4>Critérios de aprovação</h4><ul>{#each item.delivery.approval_criteria || [] as criterion}<li>{criterion}</li>{/each}</ul>
                        <h4>Eficácia e evidências</h4><ul>{#each item.delivery.effectiveness_evidence || [] as evidence}<li>{evidence}</li>{/each}</ul>
                        <h4>Passos executados pelo agente</h4><ol>{#each item.delivery.executed_steps || [] as step}<li>{step}</li>{/each}</ol>
                    {/if}
                    <h3>Achados</h3>
                    {#each item.findings || [] as finding}
                        <article class="finding"><strong>[{finding.severity}] {finding.title}</strong><p><code>{finding.file}{finding.line ? ':' + finding.line : ''}</code> · {finding.category}</p><p>{finding.evidence}</p><p>{finding.recommendation}</p></article>
                    {/each}
                    <details><summary>Patch proposto</summary><pre>{item.patch}</pre></details>
                {/if}
            </section>
        {/each}
    </div>

    {#if review.verdict}
        {#if review.verdict.winner !== 'TIE' && proposal(review.verdict.winner)?.patch_valid && proposal(review.verdict.winner)?.verification?.proof?.passed}
            <section class="panel">
                <h2>Entrega da correção</h2>
                <p>Use o patch sobre a revisão Git <code>{review.snapshot_sha}</code>. Valide o resultado no destino antes de publicar.</p>
                <button on:click={downloadWinner}>Baixar patch vencedor</button>
            </section>
        {/if}
        <section class="panel"><h2>Veredito: {review.verdict.winner === 'TIE' ? 'empate' : 'proposta ' + review.verdict.winner}</h2><p>{review.verdict.rationale}</p><details><summary>Pontuações</summary><pre>{JSON.stringify(review.verdict.scores, null, 2)}</pre></details></section>
    {:else if review.status === 'awaiting_human'}
        <section class="panel"><h2>Veredito humano</h2><label>Justificativa<textarea bind:value={rationale} rows="5" minlength="10" maxlength="4000"></textarea></label><div class="row"><button disabled={busy || rationale.trim().length < 10} on:click={() => verdict('A')}>Escolher A</button><button disabled={busy || rationale.trim().length < 10} on:click={() => verdict('B')}>Escolher B</button><button disabled={busy || rationale.trim().length < 10} on:click={() => verdict('TIE')}>Declarar empate</button></div></section>
    {/if}
{/if}

<style>
    .review-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 32rem), 1fr)); gap: 1rem; margin: 1rem 0; }
    pre { white-space: pre-wrap; overflow-wrap: anywhere; max-height: 32rem; overflow: auto; }
    .finding { border-top: 1px solid var(--border, #334); padding: .8rem 0; }
    .evidence-box { border: 1px solid #376958; background: #0b2823; padding: .9rem 1rem; margin: 1rem 0; }
    .evidence-box h3 { margin-top: 0; }
    .claim-warning { color: #ffd39a; }
</style>
