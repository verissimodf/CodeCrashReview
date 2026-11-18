<script lang="ts">
	import { cn } from '$lib/utils';

	export let variant: 'default' | 'neon' | 'outline' | 'success' | 'warning' | 'danger' | 'info' = 'default';
	export let size: 'sm' | 'md' | 'lg' = 'md';
	export let color = 'neon-cyan';
	export let glowColor = 'neon-cyanGlow';
	export let removable = false;
	export let onRemove: () => void;

	const baseClasses = 'inline-flex items-center gap-1.5 font-medium rounded-full transition-all duration-fast';

	const variantClasses = {
		default: 'bg-cyber-bgTerritary text-neutral-300 border border-cyber-border',
		neon: 'text-cyber-bg font-bold',
		outline: 'border border-current bg-transparent',
		success: 'bg-neon-green/20 text-neon-green border border-neon-green/30',
		warning: 'bg-neon-amber/20 text-neon-amber border border-neon-amber/30',
		danger: 'bg-neon-red/20 text-neon-red border border-neon-red/30',
		info: 'bg-neon-cyan/20 text-neon-cyan border border-neon-cyan/30',
	};

	const sizeClasses = {
		sm: 'px-2 py-0.5 text-xs gap-1',
		md: 'px-2.5 py-0.5 text-sm gap-1.5',
		lg: 'px-3 py-1 text-base gap-2',
	};

	const neonStyle = variant === 'neon' ? `
		background: linear-gradient(135deg, var(--tw-color-${color}), var(--tw-color-${glowColor}));
	` : '';
</script>

<span
	class={`
		${baseClasses}
		${variantClasses[variant]}
		${sizeClasses[size]}
		${removable ? 'pr-1' : ''}
	`}
	style={neonStyle}
>
	<slot />
	{#if removable}
		<button
			on:click={onRemove}
			class="ml-1 p-0.5 rounded-full hover:bg-black/20 transition-colors"
			aria-label="Remover"
		>
			<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
			</svg>
		</button>
	{/if}
</span>

<style>
	/* Neon gradient background for neon variant */
	span[style*="linear-gradient"] {
		background: linear-gradient(135deg, var(--tw-color-neon-cyan), var(--tw-color-neon-cyanGlow));
	}
</style>