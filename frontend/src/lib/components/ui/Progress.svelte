<script lang="ts">
	import { cn } from '$lib/utils';

	export let value = 0;
	export let max = 100;
	export let color = 'neon-cyan';
	export let glowColor = 'neon-cyanGlow';
	export let size: 'sm' | 'md' | 'lg' = 'md';
	export let showLabel = false;
	export let label: string;
	export let animated = false;
	export let striped = false;

	const percentage = Math.min(100, Math.max(0, (value / max) * 100));

	const sizeClasses = {
		sm: 'h-1',
		md: 'h-2',
		lg: 'h-3',
	};

	const progressStyle = `
		width: ${percentage}%;
		--progress-color: var(--tw-color-${color});
		--progress-color-glow: var(--tw-color-${glowColor});
		${animated ? 'animation: progress-pulse 2s ease-in-out infinite;' : ''}
		${striped ? 'background-image: linear-gradient(45deg, rgba(255,255,255,.15) 25%, transparent 25%, transparent 50%, rgba(255,255,255,.15) 50%, rgba(255,255,255,.15) 75%, transparent 75%, transparent); background-size: 1rem 1rem; animation: progress-stripe 1s linear infinite;' : ''}
	`;
</script>

<div class="w-full">
	{#if showLabel || label}
		<div class="flex items-center justify-between mb-1">
			{#if label}
				<span class="text-sm font-medium text-neutral-300">{label}</span>
			{/if}
			{#if showLabel}
				<span class="text-sm font-mono text-neon-cyan tabular-nums">{percentage.toFixed(0)}%</span>
			{/if}
		</div>
	{/if}

	<div class={cn('progress', sizeClasses[size])} role="progressbar" aria-valuenow={percentage} aria-valuemin={0} aria-valuemax={100}>
		<div class="progress-bar" style={progressStyle} />
	</div>

	<style>
		@keyframes progress-pulse {
			0%, 100% { opacity: 1; }
			50% { opacity: 0.7; }
		}

		@keyframes progress-stripe {
			0% { background-position: 1rem 0; }
			100% { background-position: 0 0; }
		}
	</style>
</div>