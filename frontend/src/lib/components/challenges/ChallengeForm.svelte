<script lang="ts">
	import { onMount } from 'svelte';
	import { cn } from '$lib/utils';
	import { Badge, Input, Dropdown, Modal, Button, Card, Progress, Tabs } from '$lib/components/ui';
	import { Plus, Minus, Trash2, Save, Loader2, AlertCircle, CheckCircle } from 'lucide-svelte';
	import { getSeverityColor, getCategoryIcon } from '$lib/utils';

	export let challenge: any = {
		title: '',
		description: '',
		repository_url: '',
		repository_branch: 'main',
		target_paths: [''],
		exclude_paths: [''],
		severity: 'medium',
		category: 'security',
		time_limit_seconds: 300,
		max_turns: 10,
		acceptance_criteria: [],
		scoring_weights: {
			correctness: 0.4,
			security: 0.3,
			performance: 0.15,
			style: 0.1,
			test_coverage: 0.05
		},
		known_issues: [],
		tags: [''],
		template_id: '',
		is_public: true
	};
	export let templates: any[] = [];
	export let onSubmit: (challenge: any) => void;
	export let onCancel: () => void;
	export let loading = false;

	let activeTab = 0;
	const tabs = [
		{ value: 'basic', label: 'Básico', icon: 'FileText' },
		{ value: 'criteria', label: 'Critérios', icon: 'CheckSquare' },
		{ value: 'scoring', label: 'Pontuação', icon: 'Target' },
		{ value: 'issues', label: 'Issues', icon: 'AlertTriangle' },
		{ value: 'advanced', label: 'Avançado', icon: 'Settings' }
	];

	let validationErrors: Record<string, string> = {};

	function validateForm(): boolean {
		validationErrors = {};

		if (!challenge.title.trim()) {
			validationErrors.title = 'Título é obrigatório';
		}
		if (!challenge.description.trim()) {
			validationErrors.description = 'Descrição é obrigatória';
		}
		if (!challenge.repository_url.trim()) {
			validationErrors.repository_url = 'URL do repositório é obrigatória';
		} else if (!/^https?:\/\/.+/.test(challenge.repository_url)) {
			validationErrors.repository_url = 'URL inválida';
		}
		if (challenge.target_paths.length === 0 || challenge.target_paths.every((p: string) => !p.trim())) {
			validationErrors.target_paths = 'Pelo menos um caminho alvo é obrigatório';
		}
		if (challenge.time_limit_seconds < 60 || challenge.time_limit_seconds > 3600) {
			validationErrors.time_limit_seconds = 'Tempo deve ser entre 60 e 3600 segundos';
		}
		if (challenge.max_turns < 1 || challenge.max_turns > 50) {
			validationErrors.max_turns = 'Turnos devem ser entre 1 e 50';
		}

		// Validate scoring weights sum to 1
		const weightSum = Object.values(challenge.scoring_weights).reduce((a: number, b: number) => a + b, 0);
		if (Math.abs(weightSum - 1) > 0.01) {
			validationErrors.scoring_weights = `Pesos devem somar 1.0 (atual: ${weightSum.toFixed(2)})`;
		}

		return Object.keys(validationErrors).length === 0;
	}

	function handleSubmit() {
		if (validateForm()) {
			// Clean empty arrays
			const cleanChallenge = {
				...challenge,
				target_paths: challenge.target_paths.filter((p: string) => p.trim()),
				exclude_paths: challenge.exclude_paths.filter((p: string) => p.trim()),
				tags: challenge.tags.filter((t: string) => t.trim()),
				acceptance_criteria: challenge.acceptance_criteria.filter((c: any) => c.description.trim()),
				known_issues: challenge.known_issues.filter((i: any) => i.file.trim() && i.issue.trim())
			};
			onSubmit(cleanChallenge);
		}
	}

	function addArrayItem(array: string[], value = '') {
		challenge = { ...challenge, [array]: [...challenge[array], value] };
	}

	function removeArrayItem(array: string[], index: number) {
		const newArray = challenge[array].filter((_: any, i: number) => i !== index);
		challenge = { ...challenge, [array]: newArray };
	}

	function updateArrayItem(array: string[], index: number, value: string) {
		const newArray = [...challenge[array]];
		newArray[index] = value;
		challenge = { ...challenge, [array]: newArray };
	}

	function addCriterion() {
		challenge.acceptance_criteria = [
			...challenge.acceptance_criteria,
			{ id: `ac${Date.now()}`, description: '', type: 'test', weight: 0.25, pass_threshold: 1.0, config: {} }
		];
	}

	function removeCriterion(index: number) {
		challenge.acceptance_criteria = challenge.acceptance_criteria.filter((_: any, i: number) => i !== index);
	}

	function addKnownIssue() {
		challenge.known_issues = [
			...challenge.known_issues,
			{ id: `ki${Date.now()}`, file: '', line: 0, issue: '', severity: 'medium', priority: 'medium', cwe_id: '', description: '' }
		];
	}

	function removeKnownIssue(index: number) {
		challenge.known_issues = challenge.known_issues.filter((_: any, i: number) => i !== index);
	}

	function addTag() {
		challenge.tags = [...challenge.tags, ''];
	}

	function removeTag(index: number) {
		challenge.tags = challenge.tags.filter((_: any, i: number) => i !== index);
	}

	const categories = [
		{ value: 'security', label: 'Segurança' },
		{ value: 'performance', label: 'Performance' },
		{ value: 'refactor', label: 'Refatoração' },
		{ value: 'testing', label: 'Testes' },
		{ value: 'documentation', label: 'Documentação' },
		{ value: 'migration', label: 'Migração' },
		{ value: 'bugfix', label: 'Correção de Bug' },
		{ value: 'feature', label: 'Nova Feature' }
	];

	const severities = [
		{ value: 'low', label: 'Baixa' },
		{ value: 'medium', label: 'Média' },
		{ value: 'high', label: 'Alta' },
		{ value: 'critical', label: 'Crítica' }
	];

	const criterionTypes = [
		{ value: 'test', label: 'Teste' },
		{ value: 'static_analysis', label: 'Análise Estática' },
		{ value: 'metric', label: 'Métrica' },
		{ value: 'manual', label: 'Revisão Manual' }
	];

	const severityOptions = [
		{ value: 'low', label: 'Baixa' },
		{ value: 'medium', label: 'Média' },
		{ value: 'high', label: 'Alta' },
		{ value: 'critical', label: 'Crítica' }
	];

	const priorityOptions = [
		{ value: 'low', label: 'Baixa' },
		{ value: 'medium', label: 'Média' },
		{ value: 'high', label: 'Alta' },
		{ value: 'critical', label: 'Crítica' }
	];

	function calculateWeightSum(): number {
		return Object.values(challenge.scoring_weights).reduce((a: number, b: number) => a + b, 0);
	}
</script>

<div class="max-w-4xl mx-auto">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="font-display text-2xl font-bold text-neutral-100">
				{challenge.id ? 'Editar Desafio' : 'Criar Novo Desafio'}
			</h1>
			<p class="text-neutral-400 mt-1">
				Preencha as informações do desafio. Campos marcados com * são obrigatórios.
			</p>
		</div>

		<div class="flex items-center gap-2">
			<Button variant="ghost" on:click={onCancel}>Cancelar</Button>
			<Button variant="primary" on:click={handleSubmit} loading={loading} disabled={!validateForm()}>
				<Save class="w-5 h-5 mr-2" />
				Salvar
			</Button>
		</div>
	</div>

	<!-- Validation Errors -->
	{#if Object.keys(validationErrors).length > 0}
		<div class="mb-6 p-4 bg-neon-red/10 border border-neon-red/30 rounded-lg">
			<div class="flex items-center gap-2 text-neon-red mb-2">
				<AlertCircle class="w-5 h-5" />
				<span class="font-medium">Erros de validação:</span>
			</div>
			<ul class="list-disc list-inside text-sm text-neon-red/80 space-y-1">
				{#each Object.entries(validationErrors) as [field, error]}
					<li>{field}: {error}</li>
				{/each}
			</ul>
		</div>
	{/if}

	<!-- Tabs -->
	<Tabs tabs={tabs} bind:value={activeTab} />

	<!-- Tab Content -->
	<div class="mt-6 space-y-6">
		<!-- Basic Info Tab -->
		{#if activeTab === 'basic'}
			<Card class="space-y-6">
				<h3 class="font-display text-lg font-bold text-neutral-100">Informações Básicas</h3>

				<div class="grid gap-4 md:grid-cols-2">
					<Input
						label="Título *"
						placeholder="Ex: Fix SQL Injection in User Service"
						bind:value={challenge.title}
						error={validationErrors.title}
						required
					/>

					<Dropdown
						label="Severidade *"
						items={severities.map(s => ({ label: s.label, value: s.value }))}
						bind:value={challenge.severity}
						placeholder="Selecione"
					/>
				</div>

				<Input
					label="Descrição *"
					placeholder="Descreva o problema detalhadamente..."
					bind:value={challenge.description}
					error={validationErrors.description}
					required
				/>

				<div class="grid gap-4 md:grid-cols-2">
					<Input
						label="URL do Repositório *"
						placeholder="https://github.com/org/repo"
						bind:value={challenge.repository_url}
						error={validationErrors.repository_url}
						required
					/>

					<Input
						label="Branch"
						placeholder="main"
						bind:value={challenge.repository_branch}
					/>
				</div>

				<Input
					label="Caminhos Alvo *"
					placeholder="services/user/**/*.go"
					bind:value={challenge.target_paths[0]}
					required
				/>
			{/if}

			<!-- Criteria Tab -->
			{#if activeTab === 'criteria'}
				<Card class="space-y-6">
					<div class="flex items-center justify-between">
						<h3 class="font-display text-lg font-bold text-neutral-100">Critérios de Aceitação</h3>
						<Button variant="secondary" size="sm" on:click={addCriterion}>
							<Plus class="w-4 h-4 mr-1" />
							Adicionar Critério
						</Button>
					</h3>

					{#each challenge.acceptance_criteria as criterion, index}
						<Card class="space-y-4">
							<div class="flex items-start justify-between">
								<div class="flex-1">
									<label class="label">Critério #{index + 1}</label>
									<Input
										placeholder="Descrição do critério..."
										bind:value={criterion.description}
									/>
								</div>
								<Button variant="ghost" size="sm" on:click={() => removeCriterion(index)}>
									<Trash2 class="w-4 h-4 text-neon-red" />
								</Button>
							</div>

							<div class="grid gap-4 md:grid-cols-3">
								<Dropdown
									label="Tipo"
									items={criterionTypes.map(t => ({ label: t.label, value: t.value }))}
									bind:value={criterion.type}
								/>

								<Input
									type="number"
									label="Peso (0-1)"
									min="0"
									max="1"
									step="0.05"
									bind:value={criterion.weight}
								/>

								<Input
									type="number"
									label="Threshold"
									min="0"
									max="1"
									step="0.05"
									bind:value={criterion.pass_threshold}
								/>
							</div>
						</Card>
					{/each}

					{#if challenge.acceptance_criteria.length === 0}
						<div class="text-center py-8 text-neutral-400">
							<p>Nenhum critério definido. Adicione pelo menos um.</p>
						</div>
					{/if}
				</Card>
			{/if}

			<!-- Scoring Tab -->
			{#if activeTab === 'scoring'}
				<Card class="space-y-6">
					<h3 class="font-display text-lg font-bold text-neutral-100">Pesos de Pontuação</h3>
					<p class="text-neutral-400">A soma dos pesos deve ser igual a 1.0</p>

					<div class="space-y-4">
						<div class="grid gap-4 md:grid-cols-2">
							<Input
								type="number"
								label="Corretude"
								min="0"
								max="1"
								step="0.05"
								bind:value={challenge.scoring_weights.correctness}
							/>
							<Input
								type="number"
								label="Segurança"
								min="0"
								max="1"
								step="0.05"
								bind:value={challenge.scoring_weights.security}
							/>
						</div>

						<div class="grid gap-4 md:grid-cols-3">
							<Input
								type="number"
								label="Performance"
								min="0"
								max="1"
								step="0.05"
								bind:value={challenge.scoring_weights.performance}
							/>
							<Input
								type="number"
								label="Estilo"
								min="0"
								max="1"
								step="0.05"
								bind:value={challenge.scoring_weights.style}
							/>
							<Input
								type="number"
								label="Cobertura de Testes"
								min="0"
								max="1"
								step="0.05"
								bind:value={challenge.scoring_weights.test_coverage}
							/>
						</div>

						<div class="p-4 bg-cyber-bgSecondary rounded-lg">
							<div class="flex items-center justify-between">
								<span class="font-medium">Soma total</span>
								<span class="font-display text-xl font-bold {Math.abs(calculateWeightSum() - 1) < 0.01 ? 'text-neon-green' : 'text-neon-red'}">
									{calculateWeightSum().toFixed(2)}
								</span>
							</div>
							<Progress value={calculateWeightSum()} max={1} size="lg" color="neon-green" glowColor="neon-greenGlow" />
						</div>
					</div>
				</Card>
			{/if}

			<!-- Issues Tab -->
			{#if activeTab === 'issues'}
				<Card class="space-y-6">
					<div class="flex items-center justify-between">
						<h3 class="font-display text-lg font-bold text-neutral-100">Issues Conhecidos</h3>
						<Button variant="secondary" size="sm" on:click={addKnownIssue}>
							<Plus class="w-4 h-4 mr-1" />
							Adicionar Issue
						</Button>
					</h3>

					{#each challenge.known_issues as issue, index}
						<Card class="space-y-4">
							<div class="flex items-start justify-between">
								<div class="flex-1">
									<label class="label">Issue #{index + 1}</label>
								</div>
								<Button variant="ghost" size="sm" on:click={() => removeKnownIssue(index)}>
									<Trash2 class="w-4 h-4 text-neon-red" />
								</Button>
							</div>

							<div class="grid gap-4 md:grid-cols-3">
								<Input label="Arquivo" placeholder="services/user/repository.go" bind:value={issue.file} />
								<Input type="number" label="Linha" min="0" bind:value={issue.line} />
								<Dropdown label="Severidade" items={severityOptions.map(s => ({ label: s.label, value: s.value }))} bind:value={issue.severity} />
							</div>

							<Input label="Descrição do Issue" placeholder="Descreva o problema..." bind:value={issue.issue} />

							<div class="grid gap-4 md:grid-cols-3">
								<Dropdown label="Prioridade" items={priorityOptions.map(p => ({ label: p.label, value: p.value }))} bind:value={issue.priority} />
								<Input label="CWE ID (opcional)" placeholder="CWE-89" bind:value={issue.cwe_id} />
								<Input label="Descrição Detalhada" placeholder="Detalhes adicionais..." bind:value={issue.description} />
							</div>
						</Card>
					{/each}
				</Card>
			{/if}

			<!-- Advanced Tab -->
			{#if activeTab === 'advanced'}
				<Card class="space-y-6">
					<h3 class="font-display text-lg font-bold text-neutral-100">Configurações Avançadas</h3>

					<div class="grid gap-4 md:grid-cols-2">
						<Input
							type="number"
							label="Tempo Limite (segundos)"
							min="60"
							max="3600"
							bind:value={challenge.time_limit_seconds}
						/>
						<Input
							type="number"
							label="Máximo de Turnos"
							min="1"
							max="50"
							bind:value={challenge.max_turns}
						/>
					</div>

					<div class="grid gap-4 md:grid-cols-2">
						<Dropdown
							label="Template (opcional)"
							items={[
								{ label: 'Nenhum', value: '' },
								...templates.map(t => ({ label: t.name, value: t.id }))
							]}
							bind:value={challenge.template_id}
						/>

						<Dropdown
							label="Categoria"
							items={categories.map(c => ({ label: c.label, value: c.value }))}
							bind:value={challenge.category}
						/>
					</div>

					<!-- Exclude Paths -->
					<div class="space-y-4">
						<div class="flex items-center justify-between">
							<label class="label">Caminhos Excluídos</label>
							<Button variant="secondary" size="sm" on:click={() => addArrayItem('exclude_paths')}>
								<Plus class="w-4 h-4 mr-1" />
								Adicionar
							</Button>
						</div>

						{#each challenge.exclude_paths as path, index}
							<div class="flex gap-2">
								<Input bind:value={challenge.exclude_paths[index]} placeholder="vendor/**" />
								<Button variant="ghost" size="sm" on:click={() => removeArrayItem('exclude_paths', index)}>
									<Trash2 class="w-4 h-4" />
								</Button>
							</div>
						{/each}
					</div>

					<!-- Tags -->
					<div class="space-y-4">
						<div class="flex items-center justify-between">
							<label class="label">Tags</label>
							<Button variant="secondary" size="sm" on:click={addTag}>
								<Plus class="w-4 h-4 mr-1" />
								Adicionar
							</Button>
						</div>

						<div class="flex flex-wrap gap-2">
							{#each challenge.tags as tag, index}
								<Badge variant="outline" removable onRemove={() => removeTag(index)}>
									<Input bind:value={challenge.tags[index]} class="w-32" />
								</Badge>
							{/each}
						</div>
					</div>

					<!-- Public/Private -->
					<div class="flex items-center gap-4">
						<input
							type="checkbox"
							id="is_public"
							bind:checked={challenge.is_public}
							class="w-4 h-4 accent-neon-cyan"
						/>
						<label for="is_public" class="text-neutral-300">
							Desafio público (visível para todos os usuários)
						</label>
					</div>
				</Card>
			{/if}
		</div>
	</div>
</div>