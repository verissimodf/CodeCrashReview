<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLDivAttributes } from 'svelte/elements';

	export let variant: 'default' | 'glow' | 'neon' | 'glass' = 'default';
	export let padding: 'none' | 'sm' | 'md' | 'lg' = 'md';
	export let hover = true;
	export let glowColor = 'neon-cyan';
	export let neonColor = 'neon-cyan';
	export let neonColorGlow = 'neon-cyanGlow';
	export let borderColor = 'cyber-border';

	$$restProps: HTMLDivAttributes;

	const paddingClasses = {
		none: '',
		sm: 'p-3',
		md: 'p-4',
		lg: 'p-6',
	};

	const baseClasses = 'rounded-xl transition-all duration-normal';

	const variantClasses = {
		default: 'glass border border-cyber-border',
		glow: 'glass border border-cyber-border relative overflow-hidden',
		neon: 'glass border border-cyber-border',
		glass: 'glass',
	};

	const hoverClasses = {
		true: 'hover:card-hover',
		false: '',
	};

	const paddingClass = paddingClasses[padding];
	const baseClass = baseClasses;
	const variantClass = variantClasses[variant];
	const hoverClass = hoverClasses[hover];
	const neonStyle = variant === 'neon' ? `
		background: linear-gradient(135deg, var(--tw-color-${neonColor})/10, var(--tw-color-${neonColorGlow})/10);
		box-shadow: 0 0 0 1px var(--tw-color-${neonColor}), 0 0 20px var(--tw-color-${neonColorGlow});
	` : '';
</script>

<div
	class={cn(baseClass, variantClass, hoverClass, paddingClass)}
	style={neonStyle}
	{...$$restProps}
>
	{#if variant === 'glow'}
		<div
			class="absolute inset-0 rounded-xl opacity-0 transition-opacity duration-normal pointer-events-none"
			style="background: linear-gradient(135deg, var(--tw-color-{glowColor}), transparent);"
		/>
	{/if}
	<slot />
</div>