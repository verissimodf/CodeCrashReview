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
    const metrics = [['correctness','Correção'],['security','Segurança'],['maintainability','Manutenibilidade'],['reliability','Confiabilidade'],['efficiency','Eficiência'],['evidence','Evidências']] as const;
    let scores: Record<'A'|'B',Record<string,number>> = {A:{},B:{}};
    let publishSlot: 'A'|'B' = 'A';
    let publishBranch = '';
    let publishTag = '';
    let publishRelease = false;
    let githubToken = '';
    let deploymentTargets: {id:string;label:string;profile:string}[] = [];
    let deploymentTarget = '';
    let deploymentReceipt: {url:string;source_sha256:string;functional_probe_passed:boolean}|null = null;
    async function deploy() {
        busy=true; error='';
        try { deploymentReceipt = await reviewRequest('/'+encodeURIComponent(id)+'/deploy',{target:deploymentTarget,slot:publishSlot}); }
        catch(e) { error=reviewError(e); } finally { busy=false; }
    }
    let publication: {commit_sha?:string; branch_url?:string; release_url?:string}|null = null;
    let receipts: {id:string;status:string;error_message:string|null;receipt:{commit_sha?:string;branch_url?:string}}[] = [];
    $: scoreComplete = slots.every(slot => metrics.every(([key]) => typeof scores[slot][key] === 'number' && Number.isFinite(scores[slot][key]) && (scores[slot][key] ?? -1) >= 0 && (scores[slot][key] ?? 101) <= 100));
    function total(slot:'A'|'B') { return Math.round(metrics.reduce((sum,[key]) => sum+(scores[slot]?.[key] || 0),0)/6*100)/100; }
    let scoredWinner: 'A'|'B'|'TIE' = 'TIE';
    $: scoredWinner = total('A') === total('B') ? 'TIE' : total('A') > total('B') ? 'A' : 'B';
    async function publish() {
        busy = true; error = ''; publication = null;
        const token = githubToken; githubToken = '';
        try {
            publication = await reviewRequest('/'+encodeURIComponent(id)+'/publish', {token,slot:publishSlot,branch:publishBranch,tag:publishTag,release:publishRelease});
            receipts = await reviewRequest('/'+encodeURIComponent(id)+'/publications');
        } catch(e) { error = reviewError(e); }
        finally { busy = false; }
    }
    const slots: ('A'|'B')[] = ['A', 'B'];
    $: id = $page.params.matchId || '';

    async function load() {
        if (!$auth.isAuthenticated || disposed) return;
        try { review = await reviewRequest<ReviewDetail>('/' + encodeURIComponent(id)); error = ''; if(review.verdict?.winner !== 'TIE' && review.verdict) publishSlot = review.verdict.winner; if (!publishBranch) publishBranch = 'review/' + id.slice(0,8); receipts = await reviewRequest('/'+encodeURIComponent(id)+'/publications'); }
        catch (e) { error = reviewError(e); }
    }
    async function verdict(winner: 'A'|'B'|'TIE') {
        busy = true; error = '';
        try { await reviewRequest('/' + encodeURIComponent(id) + '/verdict', {winner, rationale, scores}); await load(); }
        catch (e) { error = reviewError(e); }
        finally { busy = false; }
    }
    function proposal(slot: 'A'|'B'): ReviewProposal|undefined { return review?.proposals?.find(item => item.slot === slot); }
    function downloadWinner() {
        if (!review?.verdict) return;
        const winner = proposal(review.verdict.winner === 'TIE' ? publishSlot : review.verdict.winner);
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
        reviewRequest<{authorized:boolean;targets:typeof deploymentTargets}>('/deployment-targets').then(result => { deploymentTargets=result.authorized ? result.targets : []; deploymentTarget=deploymentTargets[0]?.id || ''; }).catch(() => {});
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
        <details><summary>Objetivo da revisão</summary><p>{review.objective}</p></details>
        <p><strong>Especialidade:</strong> {review.review_focus}</p>
        {#if review.snapshot_sha}<p><strong>Snapshot:</strong> <code>{review.snapshot_sha.slice(0,12)}</code> · {review.snapshot_files} arquivos · {review.snapshot_bytes} bytes analisados</p>{/if}
        {#if review.error_message}<p role="alert">{review.error_message}</p>{/if}
        {#if !['completed','failed','awaiting_human'].includes(review.status)}<p role="status">O processamento continua no servidor; esta página pode ser fechada.</p>{/if}
    </section>

    {#if review.verdict}<section class="panel"><h2>Veredito: {review.verdict.winner === 'TIE' ? 'empate' : 'proposta ' + review.verdict.winner}</h2><p>{review.verdict.rationale}</p><details><summary>Pontuações</summary><pre>{JSON.stringify(review.verdict.scores, null, 2)}</pre></details></section>{/if}
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
                            {#if item.verification.proof?.baseline}<p><strong>Antes:</strong> {item.verification.proof.baseline.passed ? 'testes aprovados' : 'testes reprovados'} · <strong>Depois:</strong> {item.verification.proof.passed ? 'testes aprovados' : 'testes reprovados'} · {item.verification.proof.test_count} testes · {item.verification.proof.tests_preserved ? 'testes originais preservados' : 'preservação não comprovada'}</p><details><summary>Resultado antes da correção</summary><pre>{item.verification.proof.baseline.output}</pre></details>{/if}
                            {#if item.verification.proof}<details><summary>Prova executada e identificação dos artefatos</summary><p>Código SHA-256: <code>{item.verification.proof.source_sha256}</code></p><p>Testes SHA-256: <code>{item.verification.proof.tests_sha256}</code></p><p>Imagem: <code>{item.verification.proof.image}</code></p><p>Saída: {item.verification.proof.exit_code}</p><pre>{item.verification.proof.output}</pre></details>{/if}
                            <ul>{#each item.verification.system_evidence || [] as evidence}<li>{evidence}</li>{/each}</ul>
                            {#if item.verification.unsupported_claims?.length}
                                <details class="claim-warning"><summary>Alegações do agente sem prova independente ({item.verification.unsupported_claims.length})</summary><ul>{#each item.verification.unsupported_claims as claim}<li>{claim}</li>{/each}</ul></details>
                            {/if}
                        </div>
                    {/if}
                    {#if item.delivery}
                        <details><summary>Arquitetura, método e plano de entrega</summary>
                        <p><strong>Topologia alvo:</strong> {item.delivery.topology}</p>
                        <p><strong>Design:</strong> {item.delivery.design}</p>
                        <p><strong>Objetivo final alcançável:</strong> {item.delivery.final_objective}</p>
                        <h4>Metodologia utilizada</h4><ol>{#each item.delivery.methodology || [] as step}<li>{step}</li>{/each}</ol>
                        <h4>Plano de implantação</h4><ol>{#each item.delivery.deployment_plan || [] as step}<li>{step}</li>{/each}</ol>
                        <h4>Critérios de aprovação</h4><ul>{#each item.delivery.approval_criteria || [] as criterion}<li>{criterion}</li>{/each}</ul>
                        <h4>Eficácia e evidências</h4><ul>{#each item.delivery.effectiveness_evidence || [] as evidence}<li>{evidence}</li>{/each}</ul>
                        <h4>Passos executados pelo agente</h4><ol>{#each item.delivery.executed_steps || [] as step}<li>{step}</li>{/each}</ol></details>
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
        {#if proposal(publishSlot)?.patch_valid && proposal(publishSlot)?.verification?.proof?.passed}
            <section class="panel">
                <h2>Entrega da correção</h2>
                <p>Use o patch sobre a revisão Git <code>{review.snapshot_sha}</code>. Valide o resultado no destino antes de publicar.</p>
                {#if review.verdict.winner === 'TIE'}<label>Proposta empatada para entrega<select bind:value={publishSlot}><option value="A">Proposta A</option><option value="B">Proposta B</option></select></label>{/if}
                <button on:click={downloadWinner}>Baixar patch vencedor</button>
                <details><summary>Publicar no repositório de origem</summary>
                    <p>Cria uma branch nova com a correção validada. Referências existentes com outro conteúdo não serão sobrescritas. O token é enviado apenas ao GitHub pelo servidor e não é salvo.</p>
                    <form on:submit|preventDefault={publish}>
                        <label>Nova branch<input bind:value={publishBranch} required pattern="(review|hotfix|release)/.+" maxlength="120" /></label>
                        <label>Tag opcional<input bind:value={publishTag} maxlength="120" /></label>
                        <label><input type="checkbox" bind:checked={publishRelease} /> Criar release para esta tag</label>
                        <label>Token GitHub com acesso ao repositório<input type="password" bind:value={githubToken} autocomplete="off" required minlength="20" maxlength="1024" /></label>
                        <button class="action" disabled={busy || (publishRelease && !publishTag)}>{busy ? 'Publicando…' : 'Publicar correção validada'}</button>
                    </form>
                </details>
                {#if deploymentTargets.length}<details><summary>Implantar em destino autorizado</summary><p>A implantação verifica a resposta funcional configurada. Em caso de falha, restaura o serviço anterior. O endereço de validação é acessível pelo próprio host.</p><label>Destino<select bind:value={deploymentTarget}>{#each deploymentTargets as target}<option value={target.id}>{target.label}</option>{/each}</select></label><button class="action" on:click={deploy} disabled={busy || !deploymentTarget}>{busy ? 'Verificando implantação…' : 'Implantar proposta validada'}</button></details>{/if}
                {#if deploymentReceipt}<p role="status">Serviço implantado e verificado: <code>{deploymentReceipt.url}</code>. Artefato: <code>{deploymentReceipt.source_sha256}</code>.</p>{/if}
                {#if publication}<p role="status">Publicação confirmada. Commit <code>{publication.commit_sha}</code>.</p><a href={publication.branch_url} target="_blank" rel="noopener noreferrer">Abrir branch no GitHub</a>{/if}
                {#if receipts.length}<details><summary>Histórico de publicações ({receipts.length})</summary>{#each receipts as receipt}<article class="finding"><strong>{receipt.status}</strong><p>{receipt.receipt.commit_sha || 'Commit ainda não confirmado'}</p>{#if receipt.error_message}<p role="alert">{receipt.error_message}</p>{/if}{#if receipt.receipt.branch_url}<a href={receipt.receipt.branch_url} target="_blank" rel="noopener noreferrer">Conferir no GitHub</a>{/if}</article>{/each}</details>{/if}
            </section>
        {/if}

    {:else if review.status === 'awaiting_human'}
        <section class="panel"><h2>Veredito humano</h2>
            <p>Atribua notas de 0 a 100. A média dos seis critérios determina o resultado; notas iguais geram empate.</p>
            <table class="scores"><thead><tr><th>Critério</th><th>Proposta A</th><th>Proposta B</th></tr></thead><tbody>
            {#each metrics as [key,label]}<tr><th>{label}</th>{#each slots as slot}<td><input type="number" min="0" max="100" step="1" aria-label={label+' da proposta '+slot} bind:value={scores[slot][key]} required /></td>{/each}</tr>{/each}
            <tr><th>Média</th><td>{total('A')}</td><td>{total('B')}</td></tr></tbody></table>
            <label>Justificativa<textarea bind:value={rationale} rows="4" minlength="10" maxlength="4000"></textarea></label>
            <button class="action" disabled={busy || !scoreComplete || rationale.trim().length < 10} on:click={() => verdict(scoredWinner)}>Registrar {scoredWinner === 'TIE' ? 'empate' : 'vitória da proposta '+scoredWinner}</button>
        </section>
    {/if}
{/if}

<style>
    .review-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 24rem), 1fr)); gap: 1rem; margin: 1rem 0; }
    pre { white-space: pre-wrap; overflow-wrap: anywhere; max-height: 32rem; overflow: auto; }
    .finding { border-top: 1px solid var(--border, #334); padding: .8rem 0; }
    .evidence-box { border: 1px solid var(--line); background: var(--evidence); padding: .9rem 1rem; margin: 1rem 0; }
    .evidence-box h3 { margin-top: 0; }
    .claim-warning { color: var(--warning); }
</style>
