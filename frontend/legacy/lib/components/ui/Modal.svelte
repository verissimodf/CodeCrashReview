<script lang="ts">
	import { onMount } from 'svelte';
	import { cn } from '$lib/utils';
	import { X } from 'lucide-svelte';

	export let open = false;
	export let title: string;
	export let description: string;
	export let size: 'sm' | 'md' | 'lg' | 'xl' | 'full' = 'md';
	export let closeOnOverlay = true;
	export let closeOnEscape = true;

	let contentRef: HTMLDivElement;

	const sizeClasses = {
		sm: 'max-w-sm',
		md: 'max-w-md',
		lg: 'max-w-lg',
		xl: 'max-w-xl',
		full: 'max-w-4xl',
	};

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && closeOnEscape) {
			close();
		}
	}

	function handleOverlayClick(event: MouseEvent) {
		if (closeOnOverlay && event.target === event.currentTarget) {
			close();
		}
	}

	function close() {
		open = false;
	}

	function handleContentClick(event: MouseEvent) {
		event.stopPropagation();
	}

	onMount(() => {
		if (open) {
			document.addEventListener('keydown', handleKeydown);
			document.body.style.overflow = 'hidden';
		}
		return () => {
			document.removeEventListener('keydown', handleKeydown);
			document.body.style.overflow = '';
		};
	});

	$: if (open) {
		document.addEventListener('keydown', handleKeydown);
		document.body.style.overflow = 'hidden';
	} else {
		document.removeEventListener('keydown', handleKeydown);
		document.body.style.overflow = '';
	}
</script>

{#if open}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4"
		on:click={handleOverlayClick}
		role="dialog"
		aria-modal="true"
		aria-labelledby={title ? 'modal-title' : undefined}
		aria-describedby={description ? 'modal-description' : undefined}
	>
		<div class="modal-overlay animate-fade-in" />

		<div
			class={`
				modal-content
				${sizeClasses[size]}
				w-full
				max-h-[90vh]
				overflow-y-auto
			`}
			bind:this={contentRef}
			on:click={handleContentClick}
		>
			<div class="flex items-start justify-between p-6 border-b border-cyber-border">
				<div>
					{#if title}
						<h2 id="modal-title" class="font-display text-xl font-bold text-neutral-100">
							{title}
						</h2>
					{/if}
					{#if description}
						<p id="modal-description" class="mt-1 text-sm text-neutral-400">
							{description}
						</p>
					{/if}
				</div>
				<button
					class="btn-ghost btn-sm p-2 text-neutral-400 hover:text-neon-red"
					on:click={close}
					aria-label="Fechar"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<div class="p-6">
				<slot />
			</div>
		</div>
	</div>
{/if}