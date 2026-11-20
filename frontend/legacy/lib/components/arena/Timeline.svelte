<script lang="ts">
	import { cn } from '$lib/utils';
	import { getAgentDisplayInfo, getAgentColor } from '$lib/utils';
	import { Code, Clock, CheckCircle2, X, AlertCircle, Clock as ClockIcon } from 'lucide-svelte';

	export let changes: any[] = [];
	export let currentTurn = 0;

	export function formatTime(date: string): string {
		const d = new Date(date);
		return d.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
	}
</script>

<div class="space-y-4">
	{#if changes.length === 0}
		<div class="text-center py-8 text-neutral-400">
			<ClockIcon class="w-10 h-10 mx-auto mb-2 text-neutral-600" />
			<p>Nenhuma atividade ainda</p>
			<p class="text-sm mt-1">As alterações dos agentes aparecerão aqui</p>
		</div>
	{:else}
		<div class="space-y-4">
			{#each changes as change}
				<div class="timeline-item p-3 bg-cyber-bgSecondary rounded-lg" style="--timeline-color: {getAgentColor(change.agent_persona)};">
					<div class="flex items-start gap-3">
						<div class="w-8 h-8 rounded-lg flex items-center justify-center font-display font-bold text-cyber-bg flex-shrink-0" style="background: linear-gradient(135deg, {getAgentColor(change.agent_persona)}, {getAgentColor(change.agent_persona)}80);">
							{getAgentDisplayInfo(change.agent_persona).name.charAt(0)}
						</div>
						<div class="flex-1 min-w-0">
							<div class="flex items-center justify-between">
								<p class="font-medium">{getAgentDisplayInfo(change.agent_persona).name}</p>
								<span class="text-xs text-neutral-500">Turno {change.turn}</span>
							</div>
							<p class="text-sm text-neutral-400 font-mono truncate">{change.file_path}</p>
							<p class="text-xs text-neutral-500 mt-1">{change.description}</p>
							<div class="flex items-center gap-2 mt-2">
								<span class="px-2 py-0.5 rounded text-[10px] font-medium" style="background: {getAgentColor(change.agent_persona)}20; color: {getAgentColor(change.agent_persona)};">
									{change.accepted === true ? 'Aceita' : change.accepted === false ? 'Rejeitada' : 'Pendente'}
								</span>
							</div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>