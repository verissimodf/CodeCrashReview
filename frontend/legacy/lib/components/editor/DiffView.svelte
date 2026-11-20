<script lang="ts">
	import { createDiff, type DiffLine } from './diff';

	export let oldContent = '';
	export let newContent = '';
	export let language = 'typescript';
	export let showLineNumbers = true;

	let diffLines: any[] = [];

	$: diffLines = createDiff(oldContent, newContent);

	function getLineClass(line: DiffLine): string {
		switch (line.type) {
			case 'add': return 'diff-added';
			case 'remove': return 'diff-removed';
			case 'context': return 'diff-context';
			default: return '';
		}
	}

	function getLineMarker(line: DiffLine): string {
		switch (line.type) {
			case 'add': return '+';
			case 'remove': return '-';
			case 'context': return ' ';
			default: return '';
		}
	}
</script>

<div class="font-mono text-sm bg-cyber-bgSecondary rounded-lg overflow-hidden">
	<div class="overflow-x-auto">
		<table class="w-full border-collapse">
			<thead>
				<tr class="bg-cyber-bgTerritary border-b border-cyber-border">
					{#if showLineNumbers}
						<th class="w-12 px-3 py-2 text-right text-neutral-500 font-mono text-xs">Linha</th>
					{/if}
					<th class="px-3 py-2 text-left text-neutral-500 font-mono text-xs">Código</th>
				</tr>
			</thead>
			<tbody>
				{#each diffLines as line, index}
					<tr class={['border-t border-cyber-border/50', line.type === 'add' && 'bg-neon-green/10', line.type === 'remove' && 'bg-neon-red/10'].filter(Boolean).join(' ')}>
						{#if showLineNumbers}
							<td class="w-12 px-3 py-1 text-right text-neutral-500 font-mono text-xs select-none">
								{#if line.type !== 'remove' && line.line_number}
									{line.line_number}
								{/if}
							</td>
						{/if}
						<td class="px-3 py-1 whitespace-pre-wrap border-l-2 {line.type === 'add' ? 'border-neon-green' : line.type === 'remove' ? 'border-neon-red' : 'border-transparent'}">
							<div class="flex items-start gap-2">
								<span class="flex-shrink-0 text-neutral-500 select-none">{getLineMarker(line)}</span>
								<span class="flex-1 {line.type === 'add' ? 'text-neon-green' : line.type === 'remove' ? 'text-neon-red line-through' : 'text-neutral-100'}">
									{line.content || ' '}
								</span>
							</div>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	<div class="p-3 bg-cyber-bgTerritary border-t border-cyber-border flex items-center justify-between text-xs text-neutral-400">
		<div class="flex items-center gap-4">
			<span class="flex items-center gap-1 text-neon-green">
				<span class="w-2 h-2 rounded-full bg-neon-green" />
				Adições: {diffLines.filter(l => l.type === 'add').length}
			</span>
			<span class="flex items-center gap-1 text-neon-red">
				<span class="w-2 h-2 rounded-full bg-neon-red" />
				Remoções: {diffLines.filter(l => l.type === 'remove').length}
			</span>
			<span class="flex items-center gap-1 text-neutral-500">
				<span class="w-2 h-2 rounded-full bg-neutral-500" />
				Inalterado: {diffLines.filter(l => l.type === 'context').length}
			</span>
		</div>
		<span class="text-neutral-500">
			Total: {diffLines.length} linhas
		</span>
	</div>
</div>