<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { cn } from '$lib/utils';
	import { X, CheckCircle, AlertCircle, AlertTriangle, Info } from 'lucide-svelte';

	export let id: string;
	export let type: 'success' | 'error' | 'warning' | 'info' = 'info';
	export let title: string;
	export let message: string;
	export let duration = 5000;
	export let action: { label: string; onClick: () => void };
	export let onDismiss: (id: string) => void;

	let visible = false;

	const icons = {
		success: CheckCircle,
		error: AlertCircle,
		warning: AlertTriangle,
		info: Info,
	};

	const iconColors = {
		success: 'text-neon-green',
		error: 'text-neon-red',
		warning: 'text-neon-amber',
		info: 'text-neon-cyan',
	};

	const borderColors = {
		success: 'border-neon-green',
		error: 'border-neon-red',
		warning: 'border-neon-amber',
		info: 'border-neon-cyan',
	};

	let timeout: ReturnType<typeof setTimeout>;

	onMount(() => {
		visible = true;
		if (duration !== 0) {
			timeout = setTimeout(() => dismiss(), duration);
		}
	});

	onDestroy(() => {
		if (timeout) clearTimeout(timeout);
	});

	function dismiss() {
		if (onDismiss) {
			onDismiss(id);
		}
	}
</script>

{#if visible}
	<div
		class={`
			glass-strong rounded-lg border p-4 min-w-[300px] max-w-md shadow-xl animate-slide-up
			flex items-start gap-3
			border-l-4
			${borderColors[type]}
		`}
		role="alert"
		aria-live="polite"
	>
		<div class={iconColors[type]} flex-shrink-0 mt-0.5>
			<svelte:component this={icons[type]} class="w-5 h-5" />
		</div>

		<div class="flex-1 min-w-0">
			<p class="font-medium text-neutral-100">{title}</p>
			{#if message}
				<p class="text-sm text-neutral-400 mt-1">{message}</p>
			{/if}
			{#if action}
				<button
					class="mt-2 text-sm font-medium text-neon-cyan hover:text-neon-cyan/70 underline"
					on:click={action.onClick}
				>
					{action.label}
				</button>
			{/if}
		</div>

		<button
			class="text-neutral-400 hover:text-neon-cyan flex-shrink-0"
			on:click={dismiss}
			aria-label="Fechar"
		>
			<X class="w-4 h-4" />
		</button>
	</div>
{/if}