<script lang="ts">
    import { auth } from '$lib/auth';
    import { goto } from '$app/navigation';
    let email = '', password = '', error = '', busy = false;
    async function submit() {
        busy = true; error = '';
        try { await auth.login(email, password); await goto('/arena'); }
        catch (e) { error = e instanceof Error ? e.message : 'Não foi possível entrar.'; }
        finally { busy = false; }
    }
</script>
<svelte:head><title>Entrar · CodeCrashReview</title></svelte:head>
<section class="panel auth-panel">
    <p class="kicker">Sua arena de revisão</p><h1>Entrar</h1>
    <p>Acompanhe propostas, provas e decisões dos seus repositórios.</p>
    {#if error}<p role="alert">{error}</p>{/if}
    <form on:submit|preventDefault={submit}>
        <label>E-mail<input type="email" bind:value={email} autocomplete="email" required /></label>
        <label>Senha<input type="password" bind:value={password} autocomplete="current-password" required /></label>
        <button class="action" disabled={busy}>{busy ? 'Entrando…' : 'Entrar'}</button>
    </form>
    <p>Ainda não tem conta? <a href="/auth/register">Criar conta</a></p>
</section>
<style>.auth-panel{max-width:32rem;margin:2rem auto}.auth-panel label,.auth-panel input{width:100%}</style>
