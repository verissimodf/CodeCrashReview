<script lang="ts">
	import { onMount } from 'svelte';
	import { cn } from '$lib/utils';
	import Chart from 'chart.js/auto';
	import { formatRating } from '$lib/utils';

	export let userStats: any = null;
	export let globalStats: any = null;

	let ratingChart: Chart<'line'> | null = null;
	let agentWinRateChart: Chart<'bar'> | null = null;
	let languageChart: Chart<'doughnut'> | null = null;
	let winLossChart: Chart<'bar'> | null = null;

	let ratingCanvas: HTMLCanvasElement;
	let agentWinRateCanvas: HTMLCanvasElement;
	let languageCanvas: HTMLCanvasElement;
	let winLossCanvas: HTMLCanvasElement;

	const cyberpunkColors = {
		primary: '#00ffff',
		secondary: '#ff00ff',
		success: '#00ff87',
		warning: '#ffaa00',
		danger: '#ff3366',
		grid: 'rgba(0, 255, 255, 0.1)',
		text: '#e6edf3',
		subtext: '#8b949e'
	};

	function createGradient(ctx: CanvasRenderingContext2D, color1: string, color2: string) {
		const gradient = ctx.createLinearGradient(0, 0, 0, 300);
		gradient.addColorStop(0, color1 + '80');
		gradient.addColorStop(1, color2 + '00');
		return gradient;
	}

	async function initCharts() {
		if (!userStats) return;

		// Rating History Chart
		if (ratingCanvas && userStats.overall.rating_history) {
			const history = userStats.overall.rating_history.slice(-50).reverse();
			const labels = history.map((h: any) => new Date(h.timestamp).toLocaleDateString('pt-BR', { month: 'short', day: 'numeric' }));
			const data = history.map((h: any) => h.rating);

			const ctx = ratingCanvas.getContext('2d')!;
			const gradient = createGradient(ctx, cyberpunkColors.primary, cyberpunkColors.secondary);

			ratingChart = new Chart(ctx, {
				type: 'line',
				data: {
					labels,
					datasets: [{
						label: 'Rating',
						data,
						borderColor: cyberpunkColors.primary,
						backgroundColor: gradient,
						borderWidth: 2,
						fill: true,
						tension: 0.4,
						pointRadius: 0,
						pointHoverRadius: 6,
						pointHoverBackgroundColor: cyberpunkColors.primary,
						pointHoverBorderColor: '#0d1117',
						pointHoverBorderWidth: 2
					}]
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					interaction: {
						intersect: false,
						mode: 'index'
					},
					plugins: {
						legend: { display: false },
						tooltip: {
							backgroundColor: '#161b22',
							borderColor: cyberpunkColors.primary,
							borderWidth: 1,
							padding: 12,
							titleFont: { family: 'Orbitron', size: 12 },
							bodyFont: { family: 'JetBrains Mono', size: 11 },
							callbacks: {
								label: (context) => `Rating: ${formatRating(context.parsed.y)}`
							}
						}
					},
					scales: {
						x: {
							grid: { color: cyberpunkColors.grid },
							ticks: { color: cyberpunkColors.subtext, font: { family: 'JetBrains Mono', size: 10 } }
						},
						y: {
							grid: { color: cyberpunkColors.grid },
							ticks: { color: cyberpunkColors.subtext, font: { family: 'JetBrains Mono', size: 10 }, callback: (value) => formatRating(value) }
						}
					}
				}
			});
		}

		// Agent Win Rate Chart
		if (agentWinRateCanvas && userStats.by_agent) {
			const agents = Object.entries(userStats.by_agent).map(([key, val]: [string, any]) => ({
				name: key,
				winRate: val.win_rate,
				color: getAgentColor(key)
			}));

			const ctx = agentWinRateCanvas.getContext('2d')!;

			agentWinRateChart = new Chart(ctx, {
				type: 'bar',
				data: {
					labels: agents.map(a => a.name.charAt(0).toUpperCase() + a.name.slice(1)),
					datasets: [{
						label: 'Taxa de Vitória (%)',
						data: agents.map(a => a.winRate),
						backgroundColor: agents.map(a => a.color + '80'),
						borderColor: agents.map(a => a.color),
						borderWidth: 2,
						borderRadius: 4,
						borderSkipped: false
					}]
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					indexAxis: 'y',
					plugins: {
						legend: { display: false },
						tooltip: {
							backgroundColor: '#161b22',
							borderColor: cyberpunkColors.primary,
							borderWidth: 1,
							padding: 12,
							titleFont: { family: 'Orbitron', size: 12 },
							bodyFont: { family: 'JetBrains Mono', size: 11 },
							callbacks: {
								label: (context) => `${context.parsed.x.toFixed(1)}%`
							}
						}
					},
					scales: {
						x: {
							grid: { color: cyberpunkColors.grid },
							ticks: { color: cyberpunkColors.subtext, font: { family: 'JetBrains Mono', size: 10 }, callback: (value) => `${value}%` },
							min: 0,
							max: 100
						},
						y: {
							grid: { display: false },
							ticks: { color: cyberpunkColors.text, font: { family: 'JetBrains Mono', size: 11 } }
						}
					}
				}
			});
		}

		// Language Distribution
		if (languageCanvas && userStats.by_language) {
			const languages = Object.entries(userStats.by_language).map(([key, val]: [string, any]) => ({
				name: key,
				matches: val.matches,
				color: getLanguageColor(key)
			}));

			const ctx = languageCanvas.getContext('2d')!;

			languageChart = new Chart(ctx, {
				type: 'doughnut',
				data: {
					labels: languages.map(l => l.name),
					datasets: [{
						data: languages.map(l => l.matches),
						backgroundColor: languages.map(l => l.color + '80'),
						borderColor: languages.map(l => l.color),
						borderWidth: 2,
						borderRadius: 4
					}]
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					cutout: '70%',
					plugins: {
						legend: {
							position: 'right',
							labels: {
								color: cyberpunkColors.text,
								font: { family: 'JetBrains Mono', size: 11 },
								padding: 16,
								usePointStyle: true,
								pointStyle: 'circle'
							}
						},
						tooltip: {
							backgroundColor: '#161b22',
							borderColor: cyberpunkColors.primary,
							borderWidth: 1,
							padding: 12,
							callbacks: {
								label: (context) => `${context.label}: ${context.parsed} partidas`
							}
						}
					}
				}
			});
		}

		// Win/Loss Chart
		if (winLossCanvas && userStats.overall) {
			const recent = userStats.recent_matches.slice(-20).reverse();
			const labels = recent.map((m: any) => m.challenge_title.slice(0, 20));
			const wins = recent.map((m: any) => m.result === 'win' ? 1 : 0);
			const losses = recent.map((m: any) => m.result === 'loss' ? 1 : 0);

			const ctx = winLossCanvas.getContext('2d')!;

			winLossChart = new Chart(ctx, {
				type: 'bar',
				data: {
					labels,
					datasets: [
						{
							label: 'Vitórias',
							data: wins,
							backgroundColor: '#00ff8780',
							borderColor: '#00ff87',
							borderWidth: 2,
							borderRadius: 4
						},
						{
							label: 'Derrotas',
							data: losses,
							backgroundColor: '#ff336680',
							borderColor: '#ff3366',
							borderWidth: 2,
							borderRadius: 4
						}
					]
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					plugins: {
						legend: {
							labels: { color: cyberpunkColors.text, font: { family: 'JetBrains Mono', size: 11 } }
						},
						tooltip: {
							backgroundColor: '#161b22',
							borderColor: '#00ffff',
							borderWidth: 1,
							padding: 12
						}
					},
					scales: {
						x: {
							grid: { display: false },
							ticks: { color: '#8b949e', font: { family: 'JetBrains Mono', size: 9 }, maxRotation: 45, minRotation: 45 }
						},
						y: {
							grid: { color: 'rgba(0, 255, 255, 0.1)' },
							ticks: { color: '#8b949e', font: { family: 'JetBrains Mono', size: 10 }, stepSize: 1 },
							stacked: true
						}
					}
				}
			});
		}
	}

	function getAgentColor(agent: string): string {
		const colors: Record<string, string> = {
			sentinel: '#ff3366',
			optimizer: '#00ff87',
			refactor: '#0099ff',
			guardian: '#ffaa00',
			docubot: '#ff00ff',
			polyglot: '#00ffff'
		};
		return colors[agent] || '#00ffff';
	}

	function getLanguageColor(lang: string): string {
		const colors: Record<string, string> = {
			python: '#3776ab',
			javascript: '#f7df1e',
			typescript: '#3178c6',
			java: '#ed8b00',
			cpp: '#00599c',
			rust: '#dea584',
			go: '#00add8',
			shell: '#4eaa25',
			html: '#e34f26',
			gherkin: '#42a5f5'
		};
		return colors[lang] || '#00ffff';
	}

	function destroyCharts() {
		ratingChart?.destroy();
		agentWinRateChart?.destroy();
		languageChart?.destroy();
		winLossChart?.destroy();
	}
</script>

<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
	<!-- Rating History -->
	<div class="card">
		<div class="flex items-center justify-between mb-4">
			<h3 class="font-display text-lg font-bold text-neutral-100">Histórico de Rating</h3>
			{#if userStats?.overall?.rating_change}
				<span class="px-2 py-0.5 rounded text-sm font-medium {userStats.overall.rating_change > 0 ? 'text-neon-green' : 'text-neon-red'}">
					{userStats.overall.rating_change > 0 ? '+' : ''}{userStats.overall.rating_change}
				</span>
			{/if}
		</div>
		<div class="h-64">
			<canvas bind:this={ratingCanvas} />
		</div>
	</div>

	<!-- Agent Win Rates -->
	<div class="card">
		<h3 class="font-display text-lg font-bold text-neutral-100 mb-4">Taxa de Vitória por Agente</h3>
		<div class="h-64">
			<canvas bind:this={agentWinRateCanvas} />
		</div>
	</div>

	<!-- Language Distribution -->
	<div class="card">
		<h3 class="font-display text-lg font-bold text-neutral-100 mb-4">Distribuição por Linguagem</h3>
		<div class="h-64 flex items-center justify-center">
			<canvas bind:this={languageCanvas} style="max-width: 300px;" />
		</div>
	</div>

	<!-- Win/Loss Recent -->
	<div class="card">
		<h3 class="font-display text-lg font-bold text-neutral-100 mb-4">Últimas 20 Partidas</h3>
		<div class="h-64">
			<canvas bind:this={winLossCanvas} />
		</div>
	</div>
</div>