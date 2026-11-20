<script lang="ts">
 export let stats: any = null;
</script>
{#if stats}
<section>
<h2>Estatísticas do perfil</h2>
<dl>
<dt>Rating</dt><dd>{stats.overall?.rating ?? 1200}</dd>
<dt>Partidas</dt><dd>{stats.overall?.total_matches ?? 0}</dd>
<dt>Vitórias</dt><dd>{stats.overall?.wins ?? 0}</dd>
<dt>Derrotas</dt><dd>{stats.overall?.losses ?? 0}</dd>
</dl>
<h3>Partidas recentes</h3>
{#each stats.recent_matches ?? [] as match}
<p>{match.challenge_title} · {match.result} · {match.rating_change}</p>
{:else}<p>Nenhuma partida jogada ainda.</p>{/each}
</section>
{/if}
