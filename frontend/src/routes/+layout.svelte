<script lang="ts">
    import '../site.css';
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';
    import { auth } from '$lib/stores';
    const navigation = [
        ['/', 'Início'], ['/arena','Arena'], ['/challenges','Desafios'], ['/leaderboard','Ranking'],
        ['/snippets','Snippets'], ['/profile','Perfil']
    ] as const;
    function logout() { auth.clearAuth(); goto('/auth/login'); }
</script>
<header class="site">
    <a class="brand" href="/">CodeCrash<span>Review</span></a>
    <nav aria-label="Principal">{#each navigation as [path, label]}
        <a href={path} aria-current={($page.url.pathname === path || (path !== '/' && $page.url.pathname.startsWith(path + '/'))) ? 'page' : undefined}>{label}</a>
    {/each}</nav>
    {#if $auth.isAuthenticated}<button on:click={logout}>Sair · {$auth.user?.username}</button>
    {:else}<a class="login" href="/auth/login">Entrar</a>{/if}
</header>
<main>{#key $page.url.pathname}<slot />{/key}</main>
