<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth, ws, match, ui, apiConfig } from '$lib/stores';
	import { cn, formatSeconds, getAgentDisplayInfo, getLanguageDisplayName } from '$lib/utils';
	import {
		Play,
		Pause,
		Stop,
		RefreshCw,
		ChevronLeft,
		ChevronRight,
		Eye,
		EyeOff,
		Code,
		Terminal,
		Clock,
		Zap,
		Shield,
		CheckCircle2,
		AlertCircle,
		Loader2,
		Brain,
		MessageSquare,
		Copy,
		Download,
		Fullscreen,
		X,
		Settings,
		Users,
		Trophy
	} from 'lucide-svelte';

	let matchId = '';
	let isSpectating = false;
	let showDiff = false;
	let activeFile = '';
	let sidebarTab: 'agents' | 'timeline' | 'chat' = 'agents';
	let monacoEditor: any = null;
	let originalModel: any = null;
	let proposedModel: any = null;

	// Get match ID from URL or query param
	$effect(() => {
		const params = new URLSearchParams($page.url.search);
		const urlMatchId = params.get('match');
		if (urlMatchId) {
			matchId = urlMatchId;
			joinMatch(urlMatchId);
		} else if ($match.currentMatch) {
			matchId = $match.currentMatch.id;
		}
	});

	async function joinMatch(id: string, spectate = false) {
		isSpectating = spectate;
		get(ws).connect(id);
		get(match).setMatch({} as any, spectate); // Will be updated via WS
	}

	function leaveMatch() {
		get(ws).disconnect();
		get(match).clearMatch();
		matchId = '';
		goto('/arena');
	}

	// Handle WebSocket messages for match updates
	$effect(() => {
		const unsubscribe = ws.subscribe(state => {
			if (!state.lastMessage) return;

			const msg = state.lastMessage;
			switch (msg.type) {
				case 'match_state':
					get(match).setMatch(msg.payload.match, msg.payload.your_participant_id ? false : true);
					break;
				case 'match_update':
					get(match).updateMatch(msg.payload.match);
					if (msg.payload.participants) {
						msg.payload.participants.forEach((p: any) => get(match).updateParticipant(p.user_id, p));
					}
					break;
				case 'agent_status':
					get(match).updateAgentState(msg.payload.agent_persona, {
						status: msg.payload.status,
						progress: msg.payload.progress,
						current_file: msg.payload.current_file,
						current_line: msg.payload.current_line,
						thoughts: msg.payload.thoughts
					});
					break;
				case 'code_change':
					get(match).addCodeChange(msg.payload.change);
					break;
				case 'submission':
					get(match).addSubmission(msg.payload.submission);
					break;
				case 'turn_change':
					get(match).setTurn(msg.payload.turn, msg.payload.time_remaining_seconds);
					break;
				case 'timer_update':
					get(match).updateTimer(msg.payload.time_remaining_seconds);
					break;
				case 'match_end':
					handleMatchEnd(msg.payload);
					break;
				case 'error':
					ui.showToast({ type: 'error', title: 'Erro', message: msg.payload.message });
					break;
			}
		});
		return unsubscribe;
	});

	function handleMatchEnd(payload: any) {
		ui.showToast({
			type: 'success',
			title: 'Partida Finalizada',
			message: payload.winner_id ? `Vencedor: ${payload.winner_persona}` : 'Empate!',
			duration: 10000
		});
		// Show replay modal or navigate to results
	}

	function getMyAgent() {
		const user = $auth.user;
		if (!user || !$match.currentMatch) return null;
		return $match.participants.get(user.id) || null;
	}

	function getOpponentAgent() {
		const user = $auth.user;
		if (!user || !$match.currentMatch) return null;
		for (const [id, p] of $match.participants) {
			if (id !== user.id) return p;
		}
		return null;
	}

	function formatTimeRemaining(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	function getStatusColor(status: string): string {
		const colors: Record<string, string> = {
			thinking: 'neon-amber',
			editing: 'neon-blue',
			testing: 'neon-green',
			submitting: 'neon-magenta',
			waiting: 'neutral',
			error: 'neon-red',
			completed: 'neon-green',
			idle: 'neutral'
		};
		return colors[status] || 'neutral';
	}

	function getStatusIcon(status: string) {
		const icons: Record<string, any> = {
			thinking: Brain,
			editing: Code,
			testing: CheckCircle2,
			submitting: MessageSquare,
			waiting: Clock,
			error: AlertCircle,
			completed: CheckCircle2,
			idle: Circle
		};
		return icons[status] || Circle;
	}

	function copyCode() {
		navigator.clipboard.writeText($match.codeChanges.slice(-1)[0]?.new_content || '');
		ui.showToast({ type: 'success', title: 'Copiado!', message: 'Código copiado para a área de transferência' });
	}
</script>

{#if $match.currentMatch}
<div class="h-[calc(100vh-4rem)] flex flex-col overflow-hidden">
	<!-- Arena Header -->
	<header class="glass-strong border-b border-cyber-border p-4 flex-shrink-0">
		<div class="flex items-center justify-between">
			<!-- Match Info -->
			<div class="flex items-center gap-6">
				<button class="btn-ghost btn-sm" on:click={leaveMatch}>
					<ChevronLeft class="w-4 h-4" />
				</button>
				<div>
					<h1 class="font-display text-xl font-bold text-neutral-100">{$match.currentMatch.challenge?.title || 'Partida em Andamento'}</h1>
					<div class="flex items-center gap-4 text-sm text-neutral-400 mt-1">
						<span class="flex items-center gap-1">
							<Code class="w-3 h-3" />
							{$match.currentMatch.challenge?.target_paths?.join(', ') || 'Múltiplos arquivos'}
						</span>
						<span class="flex items-center gap-1">
							<Clock class="w-3 h-3" />
							{$match.currentMatch.time_limit_seconds / 60} min
						</span>
						<span class="flex items-center gap-1">
							<Users class="w-3 h-3" />
							Turno {$match.currentTurn} / {$match.currentMatch.max_turns}
						</span>
					</div>
				</div>
			</div>

			<!-- Timer -->
			<div class="flex items-center gap-4">
				<div class="timer text-neon-cyan tabular-nums" style="--warning-threshold: 30; --danger-threshold: 10;">
					{formatTimeRemaining($match.timeRemaining)}
				</div>
				{#if $match.timeRemaining <= 30 && $match.timeRemaining > 10}
					<div class="w-2 h-2 rounded-full bg-neon-amber animate-pulse" />
				{:else if $match.timeRemaining <= 10}
					<div class="w-2 h-2 rounded-full bg-neon-red animate-pulse" />
				{/if}
			</div>

			<!-- Spectator Count -->
			<div class="flex items-center gap-2 text-neutral-400">
				<Eye class="w-4 h-4" />
				<span>{$match.spectatorCount}</span>
			</div>
		</div>
	</header>

	<!-- Main Arena -->
	<div class="flex-1 flex overflow-hidden">
		<!-- Left Agent Panel -->
		<aside class="w-72 lg:w-80 flex-shrink-0 glass border-r border-cyber-border flex flex-col">
			{#each Array.from($match.agentStates.entries()) as [persona, state], i}
				<AgentPanel
					persona={persona}
					state={state}
					isActive={$match.currentTurn - 1 === i}
					isMyAgent={$auth.user && $match.participants.get($auth.user.id)?.agent_persona === persona}
				/>
			{/each}
		</aside>

		<!-- Center: Code Editor -->
		<main class="flex-1 flex flex-col min-w-0">
			<!-- Editor Tabs -->
			<div class="glass-strong border-b border-cyber-border flex-shrink-0">
				<div class="flex items-center p-3 gap-2 overflow-x-auto">
					{#each $match.codeChanges as change, idx}
						<button
							class={cn(
								'px-3 py-1.5 rounded-lg text-sm font-medium transition-all whitespace-nowrap flex items-center gap-2',
								activeFile === change.file_path
									? 'bg-neon-cyan/10 text-neon-cyan border border-neon-cyan/30'
									: 'text-neutral-400 hover:text-neutral-100 hover:bg-cyber-bgTerritary'
							)}
							on:click={() => activeFile = change.file_path}
						>
							<Code class="w-4 h-4" />
							<span>{change.file_path.split('/').pop()}</span>
							<span class="px-1.5 py-0.5 rounded text-[10px] font-medium" style="background: {getAgentColor(change.agent_persona)}20; color: {getAgentColor(change.agent_persona)};">
								{getAgentDisplayInfo(change.agent_persona).name}
							</span>
						</button>
					{/each}
					{#if $match.codeChanges.length === 0}
						<span class="text-neutral-500 text-sm px-3">Nenhuma alteração ainda</span>
					{/if}
				</div>
			</div>

			<!-- Monaco Editor Container -->
			<div class="flex-1 relative min-h-0" bind:this={editorContainer}>
				<div class="absolute inset-0" id="monaco-editor" />
			</div>

			<!-- Diff View Toggle -->
			<div class="glass-strong border-t border-cyber-border p-3 flex-shrink-0">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<label class="flex items-center gap-2 cursor-pointer">
							<input type="checkbox" bind:checked={showDiff} class="w-4 h-4 accent-neon-cyan" />
							<span class="text-sm text-neutral-300">Mostrar Diff</span>
						</label>
						{#if showDiff && activeFile}
							<div class="flex items-center gap-2 text-sm text-neutral-400">
								<span class="flex items-center gap-1 text-neon-green">
									<CheckCircle2 class="w-3 h-3" />
									Adições
								</span>
								<span class="flex items-center gap-1 text-neon-red">
									<X class="w-3 h-3" />
									Remoções
								</span>
							</div>
						{/if}
					</div>
					<div class="flex items-center gap-2">
						<button class="btn-ghost btn-sm" on:click={copyCode} title="Copiar código">
							<Copy class="w-4 h-4" />
						</button>
						<button class="btn-ghost btn-sm" title="Download">
							<Download class="w-4 h-4" />
						</button>
						<button class="btn-ghost btn-sm" title="Fullscreen">
							<Fullscreen class="w-4 h-4" />
						</button>
					</div>
				</div>
			</div>
		</main>

		<!-- Right Sidebar -->
		<aside class="w-80 lg:w-96 flex-shrink-0 glass border-l border-cyber-border flex flex-col">
			<!-- Tabs -->
			<div class="tabs p-3 border-b border-cyber-border flex-shrink-0">
				<button
					class={cn('tab', sidebarTab === 'agents' && 'tab-active')}
					on:click={() => sidebarTab = 'agents'}
				>
					<Users class="w-4 h-4 mr-1" />
					Agentes
				</button>
				<button
					class={cn('tab', sidebarTab === 'timeline' && 'tab-active')}
					on:click={() => sidebarTab = 'timeline'}
				>
					<Clock class="w-4 h-4 mr-1" />
					Timeline
				</button>
				<button
					class={cn('tab', sidebarTab === 'chat' && 'tab-active')}
					on:click={() => sidebarTab = 'chat'}
				>
					<MessageSquare class="w-4 h-4 mr-1" />
					Chat
				</button>
			</div>

			<!-- Tab Content -->
			<div class="flex-1 overflow-y-auto p-3">
				{#if sidebarTab === 'agents'}
					<AgentDetailsList />
				{:else if sidebarTab === 'timeline'}
					<TimelineView />
				{:else}
					<ChatView />
				{/if}
			</div>
		</aside>
	</div>
</div>
{:else}
<!-- Match Lobby / Queue -->
<div class="h-[calc(100vh-4rem)] flex items-center justify-center">
	<div class="text-center max-w-md px-6">
		<div class="w-24 h-24 mx-auto mb-6 rounded-2xl bg-neon-cyan/10 border border-neon-cyan/30 flex items-center justify-center">
			<Users class="w-12 h-12 text-neon-cyan" />
		</div>
		<h1 class="font-display text-3xl font-bold text-gradient text-gradient-cyan mb-4">
			Arena CodeCrashReview
		</h1>
		<p class="text-neutral-400 mb-8 text-lg">
			Entre na fila para uma partida ranqueada ou crie um desafio privado para competir com amigos.
		</p>
		<div class="space-y-4">
			<button class="btn-primary btn-xl w-full group" on:click={() => goto('/arena?queue=ranked')}>
				<Play class="w-6 h-6 mr-2" />
				Entrar na Fila Ranqueada
			</button>
			<button class="btn-secondary btn-lg w-full group" on:click={() => goto('/challenges/create')}>
				<Plus class="w-6 h-6 mr-2" />
				Criar Desafio Privado
			</button>
			<button class="btn-ghost btn-lg w-full group" on:click={() => goto('/arena?spectate=true')}>
				<Eye class="w-6 h-6 mr-2" />
				Assistir Partidas Ao Vivo
			</button>
		</div>
		<div class="mt-10 grid grid-cols-3 gap-4 text-center">
			<div class="p-4 bg-cyber-bgSecondary rounded-xl">
				<div class="font-display text-2xl font-bold text-neon-cyan">{$leaderboard.global.length || '--'}</div>
				<div class="text-sm text-neutral-400">Jogadores Ranqueados</div>
			</div>
			<div class="p-4 bg-cyber-bgSecondary rounded-xl">
				<div class="font-display text-2xl font-bold text-neon-green">{$challenges.challenges.length || '--'}</div>
				<div class="text-sm text-neutral-400">Desafios Disponíveis</div>
			</div>
			<div class="p-4 bg-cyber-bgSecondary rounded-xl">
				<div class="font-display text-2xl font-bold text-neon-amber">6</div>
				<div class="text-sm text-neutral-400">Agentes Únicos</div>
			</div>
		</div>
	</div>
</div>
{/if}

<!-- Agent Panel Component -->
<script lang="ts" context="module">
	export interface AgentPanelProps {
		persona: string;
		state: any;
		isActive: boolean;
		isMyAgent: boolean;
	}
</script>

{#if $match.currentMatch}
<script lang="ts" context="module">
	import { onMount, onDestroy } from 'svelte';
	import * as monaco from 'monaco-editor';
	import { match, auth, ui } from '$lib/stores';
	import { getAgentDisplayInfo, getMonacoLanguageId } from '$lib/utils';

	let editorContainer: HTMLDivElement;
	let monacoEditor: monaco.editor.IStandaloneCodeEditor;
	let originalModel: monaco.editor.ITextModel;
	let proposedModel: monaco.editor.ITextModel;
	let diffEditor: monaco.editor.IStandaloneDiffEditor;

	onMount(async () => {
		// Load Monaco Editor
		if (typeof window !== 'undefined' && editorContainer) {
			// Configure Monaco
			monaco.editor.defineTheme('codecrash', {
				base: 'vs-dark',
				inherit: true,
				rules: [],
				colors: {
					'editor.background': '#0d1117',
					'editor.foreground': '#e6edf3',
					'editor.lineHighlightBackground': '#161b22',
					'editorCursor.foreground': '#00ffff',
					'editor.selectionBackground': '#00ffff30',
					'editor.inactiveSelectionBackground': '#00ffff15',
					'editorWhitespace.foreground': '#30363d'
				}
			});

			// Create editor
			monacoEditor = monaco.editor.create(editorContainer, {
				theme: 'codecrash',
				automaticLayout: true,
				minimap: { enabled: false },
				lineNumbers: 'on',
				fontSize: 13,
				fontFamily: 'JetBrains Mono',
				lineHeight: 20,
				scrollBeyondLastLine: false,
				renderWhitespace: 'selection',
				tabSize: 2,
				wordWrap: 'on'
			});

			// Load initial files
			loadFiles();
		}
	});

	function loadFiles() {
		if (!$match.currentMatch?.challenge) return;

		// For demo, create sample content
		const sampleCode = `// Challenge: {$match.currentMatch.challenge.title}
// Target: {$match.currentMatch.challenge.target_paths[0]}
// Agent: Sentinel (Security)

function authenticateUser(username, password) {
    // VULNERABILITY: SQL Injection via string concatenation
    const query = "SELECT * FROM users WHERE username = '" + username + "' AND password = '" + password + "'";
    return database.execute(query);
}

function getUserProfile(userId) {
    // VULNERABILITY: No input validation
    return database.query("SELECT * FROM profiles WHERE id = " + userId);
}`;

		originalModel = monaco.editor.createModel(sampleCode, 'typescript', monaco.Uri.parse('file:///original.ts'));
		proposedModel = monaco.editor.createModel(sampleCode.replace("username + \"'", "username + \"'"), 'typescript', monaco.Uri.parse('file:///proposed.ts'));

		monacoEditor.setModel(originalModel);
	}

	onDestroy(() => {
		monacoEditor?.dispose();
		originalModel?.dispose();
		proposedModel?.dispose();
	});
</script>

<!-- Agent Panel Component -->
<script lang="ts" context="module">
	export interface AgentPanelProps {
		persona: string;
		state: any;
		isActive: boolean;
		isMyAgent: boolean;
	}
</script>

{#snippet AgentPanel(persona, state, isActive, isMyAgent)}
	<div class={cn(
		'p-4 border-b border-cyber-border transition-all duration-fast relative overflow-hidden',
		isActive && 'bg-neon-cyan/5 ring-1 ring-inset ring-neon-cyan/30',
		isMyAgent && 'border-l-4 border-neon-cyan'
	)}>
		<div class="flex items-center justify-between mb-3">
			<div class="flex items-center gap-3">
				<div class={cn(
					'w-12 h-12 rounded-xl flex items-center justify-center font-display font-bold text-cyber-bg',
					'animate-float'
				)} style="background: linear-gradient(135deg, {getAgentColor(persona)}, {getAgentColor(persona)}80);">
					{getAgentDisplayInfo(persona).icon}
				</div>
				<div>
					<p class="font-display text-lg font-bold text-neutral-100">{getAgentDisplayInfo(persona).name}</p>
					<p class="text-xs text-neutral-500">{getAgentDisplayInfo(persona).description}</p>
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
			<div class="progress">
				<div class="progress-bar" style="width: {state.progress}%; --progress-color: {getAgentColor(persona)}; --progress-color-glow: {getAgentColor(persona)}80;" />
			</div>
		</div>

		<!-- Stats -->
		<div class="grid grid-cols-3 gap-2 text-center">
			<div class="p-2 bg-cyber-bgSecondary rounded-lg">
				<p class="font-display text-xl font-bold text-{getAgentColor(persona)}">{state.tokens_used || 0}</p>
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
			<div class="mt-3 p-2 bg-cyber-bgSecondary rounded-lg border border-cyber-border">
				<p class="text-xs text-neutral-500 mb-1">Trabalhando em:</p>
				<p class="font-mono text-sm text-neutral-100 truncate">{state.current_file}</p>
				{#if state.current_line}
					<p class="text-xs text-neon-cyan">Linha {state.current_line}</p>
				{/if}
			</div>
		{/if}

		<!-- Thoughts -->
		{#if state.thoughts.length > 0}
			<div class="mt-3">
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
{/snippet}

<!-- Agent Details List -->
{#snippet AgentDetailsList()}
	<div class="space-y-4">
		{#each Array.from($match.agentStates.entries()) as [persona, state]}
			<div class="p-4 bg-cyber-bgSecondary rounded-xl border border-cyber-border">
				<div class="flex items-center gap-3 mb-3">
					<div class="w-10 h-10 rounded-lg flex items-center justify-center font-display font-bold text-cyber-bg" style="background: linear-gradient(135deg, {getAgentColor(persona)}, {getAgentColor(persona)}80);">
						{getAgentDisplayInfo(persona).name.charAt(0)}
					</div>
					<div>
						<p class="font-medium">{getAgentDisplayInfo(persona).name}</p>
						<p class="text-xs text-neutral-500">{getAgentDisplayInfo(persona).description}</p>
					</div>
				</div>
				<div class="flex items-center gap-4 text-sm">
					<span class={cn('flex items-center gap-1', `text-${getStatusColor(state.status)}`)}>
						<span class="w-2 h-2 rounded-full bg-current" />
						{state.status}
					</span>
					<span class="text-neutral-400">{state.progress}%</span>
					<span class="text-neutral-400">{state.proposed_changes?.length || 0} alterações</span>
				</div>
				<div class="progress mt-3">
					<div class="progress-bar" style="width: {state.progress}%; --progress-color: {getAgentColor(persona)}; --progress-color-glow: {getAgentColor(persona)}80;" />
				</div>
			</div>
		{/each}
	</div>
{/snippet}

<!-- Timeline View -->
{#snippet TimelineView()}
	<div class="space-y-4">
		{#each $match.codeChanges as change}
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
		{#if $match.codeChanges.length === 0}
			<div class="text-center py-8 text-neutral-400">
				<Clock class="w-10 h-10 mx-auto mb-2 text-neutral-600" />
				<p>Nenhuma atividade ainda</p>
			</div>
		{/if}
	</div>
{/snippet}

<!-- Chat View -->
{#snippet ChatView()}
	<div class="space-y-3">
		<div class="text-center py-8 text-neutral-400">
			<MessageSquare class="w-10 h-10 mx-auto mb-2 text-neutral-600" />
			<p>Chat do espectador</p>
			<p class="text-sm mt-1">Disponível para espectadores</p>
		</div>
	</div>
{/snippet}