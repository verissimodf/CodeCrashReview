<script lang="ts">
    import { auth } from '$lib/auth';
</script>
<svelte:head><title>CodeCrashReview</title></svelte:head>
<section class="hero">
    <div class="hero-copy">
        <h1>Duas leituras do código.<br />Um veredito com evidências.</h1>
        <p class="lead">Aponte um repositório GitHub ou local autorizado. Dois agentes trabalham sobre o mesmo snapshot, em paralelo, e entregam correções independentes para avaliação autônoma ou humana.</p>
        {#if $auth.isAuthenticated}<a class="btn-primary" href="/arena">Abrir nova revisão</a>
        {:else}<div class="hero-actions"><a class="btn-primary" href="/auth/register">Criar conta</a><a class="text-link" href="/auth/login">Entrar na arena</a></div>{/if}
        <dl class="proof-strip">
            <div><dt>2</dt><dd>análises independentes</dd></div>
            <div><dt>1</dt><dd>snapshot imutável</dd></div>
            <div><dt>SHA</dt><dd>origem identificada</dd></div>
        </dl>
    </div>
    <div class="pipeline" aria-label="Fluxo da revisão concorrente">
        <div class="snapshot-node"><span>Snapshot Git</span><code>sha: imutável</code></div>
        <div class="agent-node agent-a"><span>Agente A</span><strong>Segurança e risco</strong><small>análise independente</small></div>
        <div class="agent-node agent-b"><span>Agente B</span><strong>Correção e design</strong><small>análise independente</small></div>
        <div class="judge-node"><span>Avaliador</span><strong>Veredito auditável</strong><small>IA autônoma ou operador</small></div>
    </div>
</section>

<section class="principles" aria-label="Garantias da plataforma">
    <article><h2>Correções aplicáveis</h2><p>O backend converte substituições exatas em patches determinísticos e verifica a aplicação sobre o snapshot.</p></article>
    <article><h2>Concorrência real</h2><p>Os agentes começam juntos, não compartilham conclusões e têm tempos e resultados persistidos separadamente.</p></article>
    <article><h2>Provas identificadas</h2><p>O perfil de pontuação executa testes isolados. Outros componentes exigem validação adicional; alegações do modelo não substituem testes.</p></article>
</section>
