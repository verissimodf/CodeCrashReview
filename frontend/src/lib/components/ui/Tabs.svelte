<script lang="ts">
	import { cn } from '$lib/utils';

	export let tabs: { value: string; label: string; icon?: any; disabled?: boolean; count?: number }[] = [];
	export let value: string;
	export let variant: 'default' | 'pills' | 'underline' = 'default';
	export let fullWidth = false;

	const variantClasses = {
		default: 'tabs',
		pills: 'tabs',
		underline: 'flex border-b border-cyber-border',
	};

	const tabClasses = {
		default: 'tab',
		pills: 'tab',
		underline: 'px-4 py-3 border-b-2 -mb-px text-sm font-medium transition-all duration-fast',
	};

	const activeClasses = {
		default: 'tab-active',
		pills: 'tab-active',
		underline: 'border-neon-cyan text-neon-cyan',
	};

	const inactiveClasses = {
		default: '',
		pills: '',
		underline: 'text-neutral-400 hover:text-neutral-100',
	};
</script>

<div class={cn('flex', fullWidth && 'w-full', variant === 'underline' && 'border-b border-cyber-border')}>
	{#if variant === 'default' || variant === 'pills'}
		<div class={cn(variantClasses[variant], fullWidth && 'w-full')}>
			{#each tabs as tab}
				<button
					type="button"
					class={cn(
						tabClasses[variant],
						value === tab.value ? activeClasses[variant] : inactiveClasses[variant],
						tab.disabled && 'opacity-50 cursor-not-allowed'
					)}
					on:click={() => !tab.disabled && (value = tab.value)}
					disabled={tab.disabled}
					role="tab"
					aria-selected={value === tab.value}
					aria-disabled={tab.disabled}
				>
					<div class="flex items-center gap-2">
						{#if tab.icon}
							<svelte:component this={tab.icon} class="w-4 h-4" />
						{/if}
						<span>{tab.label}</span>
						{#if tab.count !== undefined}
							<span class="px-1.5 py-0.5 rounded-full text-[10px] font-medium bg-cyber-bgTerritary text-neutral-400">
								{tab.count}
							</span>
						{/if}
					</div>
				</button>
			{/each}
		</div>
	{:else if variant === 'underline'}
		<div class="flex border-b border-cyber-border" role="tablist">
			{#each tabs as tab}
				<button
					type="button"
					class={cn(
						tabClasses.underline,
						value === tab.value ? activeClasses.underline : inactiveClasses.underline,
						tab.disabled && 'opacity-50 cursor-not-allowed'
					)}
					on:click={() => !tab.disabled && (value = tab.value)}
					disabled={tab.disabled}
					role="tab"
					aria-selected={value === tab.value}
					aria-disabled={tab.disabled}
				>
					<div class="flex items-center gap-2">
						{#if tab.icon}
							<svelte:component this={tab.icon} class="w-4 h-4" />
						{/if}
						<span>{tab.label}</span>
						{#if tab.count !== undefined}
							<span class="px-1.5 py-0.5 rounded-full text-[10px] font-medium bg-cyber-bgTerritary text-neutral-400">
								{tab.count}
							</span>
						{/if}
					</div>
				</button>
			{/each}
		</div>
	{/if}
</div>
