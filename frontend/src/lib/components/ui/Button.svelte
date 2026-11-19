<script lang="ts">
	import { cn } from '$lib/utils';


	export let variant: 'primary' | 'secondary' | 'danger' | 'ghost' | 'neon' = 'primary';
	export let size: 'sm' | 'md' | 'lg' | 'xl' = 'md';
	export let disabled = false;
	export let loading = false;
	export let fullWidth = false;
	export let type: 'button' | 'submit' | 'reset' = 'button';
	export let neonColor = 'neon-cyan';
	export let neonColorGlow = 'neon-cyanGlow';
	export let badgeColor = 'neon-cyan';
	export let badgeColorGlow = 'neon-cyanGlow';
	export let glowColor = 'neon-cyan';
	export let badgeColorText = 'cyber-bg';


	const baseClasses = 'inline-flex items-center justify-center gap-2 font-medium transition-all duration-fast focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-cyber-bg disabled:opacity-50 disabled:cursor-not-allowed';

	const variantClasses = {
		primary: 'bg-neon-cyan text-cyber-bg hover:bg-neon-cyan/90 active:bg-neon-cyan focus-visible:ring-neon-cyan shadow-neon-cyan/20',
		secondary: 'bg-cyber-bgTerritary text-neutral-100 border border-cyber-border hover:bg-cyber-borderBright active:bg-cyber-border focus-visible:ring-cyber-borderBright',
		danger: 'bg-neon-red text-cyber-bg hover:bg-neon-red/90 active:bg-neon-red focus-visible:ring-neon-red shadow-neon-red/20',
		ghost: 'bg-transparent text-neutral-300 hover:bg-cyber-bgTerritary hover:text-neutral-100 active:bg-cyber-border focus-visible:ring-cyber-borderBright',
		neon: 'relative overflow-hidden text-cyber-bg font-bold',
	};

	const sizeClasses = {
		sm: 'px-3 py-1.5 text-sm gap-1.5',
		md: 'px-4 py-2 text-base gap-2',
		lg: 'px-6 py-3 text-lg gap-2',
		xl: 'px-8 py-4 text-xl gap-3',
	};

	const widthClasses = {
		true: 'w-full',
		false: '',
	};

	const neonStyle = `
		background: linear-gradient(135deg, var(--tw-color-${neonColor}), var(--tw-color-${neonColorGlow}));
	` as const;
</script>

<button
    on:click
	{type}
	{disabled}
	class={cn(
		baseClasses,
		variantClasses[variant],
		sizeClasses[size],
		fullWidth && "w-full",
		variant === 'neon' && 'neon-btn',
		loading && 'opacity-75 cursor-wait'
	)}
	style={variant === 'neon' ? neonStyle : undefined}
	{...$$restProps}
>
	{#if loading}
		<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
			<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
			<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
		</svg>
		<span>Carregando...</span>
	{:else}
		<slot />
	{/if}
</button>

<style>
	.neon-btn {
		background: linear-gradient(135deg, var(--tw-color-neon-cyan), var(--tw-color-neon-cyanGlow));
	}

	.neon-btn::before {
		content: '';
		position: absolute;
		inset: 0;
		background: linear-gradient(135deg, var(--tw-color-neon-cyan), var(--tw-color-neon-cyanGlow));
		opacity: 0;
		transition: opacity 0.3s ease;
	}

	.neon-btn:hover::before {
		opacity: 1;
	}

	.neon-btn span {
		position: relative;
		z-index: 1;
	}
</style>