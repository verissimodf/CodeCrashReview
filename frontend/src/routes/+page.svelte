<script lang="ts">
    import { auth } from '$lib/stores';
</script>
<svelte:head><title>CodeCrashReview</title></svelte:head>
<h1>Arena de revisão de código</h1>
<section class="panel">
    <h2>Dispute uma partida</h2>
    <p>Escolha um desafio e um agente, envie sua solução e acompanhe a avaliação das duas entregas.</p>
    {#if $auth.isAuthenticated}<a class="btn-primary" href="/arena">Abrir arena</a>
    {:else}<div class="row"><a class="btn-primary" href="/auth/register">Criar conta</a><a href="/auth/login">Já tenho conta</a></div>{/if}
</section>
