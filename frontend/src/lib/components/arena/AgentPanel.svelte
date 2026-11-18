<script lang="ts">
	import { cn } from '$lib/utils';
	import { Brain, Code, CheckCircle2, MessageSquare, Clock, AlertCircle, Circle, Zap, Shield, BookOpen, Globe } from 'lucide-svelte';
	import { getAgentDisplayInfo, getAgentColor } from '$lib/utils';
	import { Progress } from '$lib/components/ui';

	export let persona: string;
	export let state: {
		status: string;
		progress: number;
		tokens_used: number;
		latency_ms: number;
		last_action: string;
		current_file?: string;
		current_line?: number;
		thoughts: string[];
		proposed_changes: any[];
	} = {
		status: 'idle',
		progress: 0,
		tokens_used: 0,
		latency_ms: 0,
		last_action: '',
		thoughts: [],
		proposed_changes: []
	};
	export let isActive = false;
	export let isMyAgent = false;

	const agentInfo = getAgentDisplayInfo(persona);
	const agentColor = getAgentColor(persona);
	const statusColors: Record<string, string> = {
		thinking: 'neon-amber',
		editing: 'neon-blue',
		testing: 'neon-green',
		submitting: 'neon-magenta',
		waiting: 'neutral',
		error: 'neon-red',
		completed: 'neon-green',
		idle: 'neutral'
	};

	const statusIcons: Record<string, any> = {
		thinking: Brain,
		editing: Code,
		testing: CheckCircle2,
		submitting: MessageSquare,
		waiting: Clock,
		error: AlertCircle,
		completed: CheckCircle2,
		idle: Circle
	};

	const StatusIcon = statusIcons[state.status] || Circle;
	const statusColor = statusColors[state.status] || 'neutral';
</script>

<div class={cn(
	'p-4 border-b border-cyber-border transition-all duration-fast relative overflow-hidden',
	isActive && 'bg-neon-cyan/5 ring-1 ring-inset ring-neon-cyan/30',
	isMyAgent && 'border-l-4 border-neon-cyan'
)}>
	<div class="flex items-center justify-between mb-3">
		<div class="flex items-center gap-3">
			<div class={cn(
				'w-12 h-12 rounded-xl flex items-center justify-center font-display font-bold text-cyber-bg animate-float',
				'animate-float'
			)} style="background: linear-gradient(135deg, {agentColor}, {agentColor}80);">
				<svelte:component this={agentInfo.icon} class="w-6 h-6" />
			</div>
			<div>
				<p class="font-display text-lg font-bold text-neutral-100">{agentInfo.name}</p>
				<p class="text-xs text-neutral-500">{agentInfo.description}</p>
			</div>
		</div>
		{#if isMyAgent}
			<span class="px-2 py-0.5 rounded text-[10px] font-medium bg-neon-cyan/20 text-neon-cyan">VOCÊ</span>
		{/if}
	</div>

	<!-- Status -->
	<div class="mb-3">
		<div class="flex items-center justify-between mb-1">
			<span class="text-xs text-neutral-400 uppercase tracking-wide">Status</span>
			<span class={cn('agent-status', `agent-status-${state.status}`)}>
				<span class="w-2 h-2 rounded-full bg-current animate-pulse" />
				{state.status}
			</span>
		</div>
		<Progress value={state.progress} max={100} size="sm" color={agentColor} glowColor={statusColor}Glow />
	</div>

	<!-- Stats -->
	<div class="grid grid-cols-3 gap-2 text-center mb-3">
		<div class="p-2 bg-cyber-bgSecondary rounded-lg">
			<p class="font-display text-xl font-bold" style="color: {agentColor};">{state.tokens_used || 0}</p>
			<p class="text-[10px] text-neutral-500">Tokens</p>
		</div>
		<div class="p-2 bg-cyber-bgSecondary rounded-lg">
			<p class="font-display text-xl font-bold text-neon-cyan">{state.latency_ms || 0}ms</p>
			<p class="text-[10px] text-neutral-500">Latência</p>
		</div>
		<div class="p-2 bg-cyber-bgSecondary rounded-lg">
			<p class="font-display text-xl font-bold text-neon-green">{state.proposed_changes?.length || 0}</p>
			<p class="text-[10px] text-neutral-500">Alterações</p>
		</div>
	</div>

	<!-- Current Action -->
	{#if state.current_file}
		<div class="mb-3 p-2 bg-cyber-bgSecondary rounded-lg border border-cyber-border">
			<p class="text-xs text-neutral-500 mb-1">Trabalhando em:</p>
			<p class="font-mono text-sm text-neutral-100 truncate">{state.current_file}</p>
			{#if state.current_line}
				<p class="text-xs text-neon-cyan">Linha {state.current_line}</p>
			{/if}
		</div>
	{/if}

	<!-- Thoughts -->
	{#if state.thoughts.length > 0}
		<div class="mb-3">
			<p class="text-xs text-neutral-500 mb-2 flex items-center gap-1">
				<Brain class="w-3 h-3" />
				Processamento
			</p>
			<div class="space-y-1 max-h-32 overflow-y-auto">
				{#each state.thoughts.slice(-5) as thought}
					<div class="text-xs text-neutral-400 bg-cyber-bgSecondary p-2 rounded animate-fade-in">
						{thought}
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>