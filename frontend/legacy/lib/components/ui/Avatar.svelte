<script lang="ts">
	import { cn } from '$lib/utils';

	export let src: string;
	export let alt: string;
	export let name: string;
	export let color = 'neon-cyan';
	export let glowColor = 'neon-cyanGlow';
	export let size: 'xs' | 'sm' | 'md' | 'lg' | 'xl' = 'md';
	export let shape: 'circle' | 'square' = 'circle';
	export let status: 'online' | 'offline' | 'busy' | 'away' | null = null;
	export let statusPosition: 'bottom-right' | 'bottom-left' | 'top-right' | 'top-left' = 'bottom-right';

	const sizeClasses = {
		xs: 'w-6 h-6 text-[10px]',
		sm: 'w-8 h-8 text-xs',
		md: 'w-10 h-10 text-sm',
		lg: 'w-12 h-12 text-base',
		xl: 'w-16 h-16 text-lg',
	};

	const shapeClasses = {
		circle: 'rounded-full',
		square: 'rounded-lg',
	};

	const statusClasses = {
		online: 'bg-neon-green',
		offline: 'bg-neutral-500',
		busy: 'bg-neon-red',
		away: 'bg-neon-amber',
	};

	const statusSizeClasses = {
		xs: 'w-1.5 h-1.5',
		sm: 'w-2 h-2',
		md: 'w-2.5 h-2.5',
		lg: 'w-3 h-3',
		xl: 'w-4 h-4',
	};

	const statusPositionClasses = {
		'bottom-right': 'bottom-0 right-0',
		'bottom-left': 'bottom-0 left-0',
		'top-right': 'top-0 right-0',
		'top-left': 'top-0 left-0',
	};

	const initials = name
		?.split(' ')
		.map(w => w[0])
		.slice(0, 2)
		.join('')
		.toUpperCase();
</script>

<div
	class={cn(
		'inline-flex items-center justify-center font-medium font-display overflow-hidden relative',
		'bg-gradient-to-br',
		`from-[var(--tw-color-${color})] to-[var(--tw-color-${glowColor})]`,
		sizeClasses[size],
		shapeClasses[shape]
	)}
	role="img"
	aria-label={name || alt}
>
	{#if src}
		<img
			src={src}
			alt={alt || name}
			class="w-full h-full object-cover"
		/>
	{:else}
		<span class="text-cyber-bg select-none">{initials || '?'}</span>
	{/if}

	{#if status}
		<span
			class={cn(
				'absolute border-2 border-cyber-bg rounded-full',
				statusClasses[status],
				statusSizeClasses[size],
				statusPositionClasses[statusPosition]
			)}
			aria-label={status}
		/>
	{/if}
</div>