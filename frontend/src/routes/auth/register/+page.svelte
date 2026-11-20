<script lang="ts">
    import { auth } from '$lib/auth';
    import { goto } from '$app/navigation';
    let username = '', email = '', password = '', confirmation = '', error = '', busy = false;
    async function submit() {
        error = '';
        if (password !== confirmation) { error = 'As senhas não coincidem.'; return; }
        busy = true;
        try { await auth.register(username, email, password); await goto('/arena'); }
        catch (e) { error = e instanceof Error ? e.message : 'Não foi possível criar a conta.'; }
        finally { busy = false; }
    }
</script>
<svelte:head><title>Criar conta · CodeCrashReview</title></svelte:head>
<section class="panel auth-panel">
    <p class="kicker">Comece uma revisão independente</p><h1>Criar conta</h1>
    {#if error}<p role="alert">{error}</p>{/if}
    <form on:submit|preventDefault={submit}>
        <label>Nome de usuário<input bind:value={username} autocomplete="username" pattern={'[a-zA-Z0-9_\\-]{3,20}'} minlength="3" maxlength="20" required /></label>
        <label>E-mail<input type="email" bind:value={email} autocomplete="email" required /></label>
        <label>Senha<input type="password" bind:value={password} autocomplete="new-password" minlength="8" required /></label>
        <label>Confirmar senha<input type="password" bind:value={confirmation} autocomplete="new-password" minlength="8" required /></label>
        <button class="action" disabled={busy}>{busy ? 'Criando…' : 'Criar conta'}</button>
    </form>
    <p>Já tem conta? <a href="/auth/login">Entrar</a></p>
</section>
<style>.auth-panel{max-width:32rem;margin:2rem auto}.auth-panel label,.auth-panel input{width:100%}</style>
