<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { cn } from '$lib/utils';
	import { ChevronDown, ChevronUp, Check } from 'lucide-svelte';

	export let trigger: 'click' | 'hover' = 'click';
	export let align: 'left' | 'right' = 'right';
	export let items: { label: string; value: string; icon?: any; disabled?: boolean; selected?: boolean }[] = [];
	export let value: string;
	export let placeholder = 'Selecionar...';
	export let disabled = false;
	export let searchable = false;
	export let multiple = false;

	let open = false;
	let searchQuery = '';
	let dropdownRef: HTMLDivElement;
	let buttonRef: HTMLButtonElement;

	function toggle() {
		if (!disabled) {
			open = !open;
		}
	}

	function close() {
		open = false;
	}

	function selectItem(item: typeof items[0]) {
		if (item.disabled) return;

		if (multiple) {
			// Handle multiple selection
			value = value.split(',').filter(v => v !== item.value).concat(item.value).join(',');
		} else {
			value = item.value;
			close();
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (!open) return;

		switch (event.key) {
			case 'Escape':
				close();
				break;
			case 'ArrowDown':
				event.preventDefault();
				focusNext();
				break;
			case 'ArrowUp':
				event.preventDefault();
				focusPrev();
				break;
			case 'Enter':
				event.preventDefault();
				const focused = document.activeElement as HTMLElement;
				if (focused.dataset.value) {
					const item = items.find(i => i.value === focused.dataset.value);
					if (item && !item.disabled) selectItem(item);
				}
				break;
		}
	}

	function focusNext() {
		const items_list = dropdownRef?.querySelectorAll('[data-value]') as NodeListOf<HTMLElement>;
		const current = document.activeElement as HTMLElement;
		let found = false;
		for (const item of Array.from(items_list)) {
			if (found) {
				item.focus();
				return;
			}
			if (item === document.activeElement) found = true;
		}
		items_list[0]?.focus();
	}

	function focusPrev() {
		const items_list = dropdownRef?.querySelectorAll('[data-value]') as NodeListOf<HTMLElement>;
		let prev: HTMLElement | null = null;
		for (const item of Array.from(items_list)) {
			if (item === document.activeElement) {
				if (prev) {
					prev.focus();
					return;
				}
				break;
			}
			prev = item;
		}
		if (prev) prev.focus();
	}

	function handleClickOutside(event: MouseEvent) {
		if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
			close();
		}
	}

	function handleButtonClick(event: MouseEvent) {
		event.stopPropagation();
		toggle();
	}

	// Handle click outside
	let clickOutsideHandler: (event: MouseEvent) => void;

	onMount(() => {
		clickOutsideHandler = (event: MouseEvent) => {
			if (dropdownRef && !dropdownRef.contains(event.target as Node) && buttonRef && !buttonRef.contains(event.target as Node)) {
				close();
			}
		};
		document.addEventListener('mousedown', clickOutsideHandler);
		document.addEventListener('keydown', handleKeydown);
	});

	onDestroy(() => {
		document.removeEventListener('mousedown', clickOutsideHandler);
		document.removeEventListener('keydown', handleKeydown);
	});
</script>

<div class="relative inline-block" bind:this={dropdownRef}>
	<button
		bind:this={buttonRef}
		class={`
			btn-secondary w-full justify-between
			${disabled ? 'opacity-50 cursor-not-allowed' : ''}
		`}
		on:click={handleButtonClick}
		disabled={disabled}
		aria-haspopup="listbox"
		aria-expanded={open}
		aria-label={placeholder}
		type="button"
	>
		<span class="flex-1 text-left truncate">
			{#if value}
				{#if multiple}
					<span class="flex flex-wrap gap-1">
						{#each value.split(',') as val}
							{#each items.filter(i => i.value === val) as item}
								<Badge variant="outline" size="sm" removable onRemove={() => {
									value = value.split(',').filter(v => v !== item.value).join(',');
								}}>
									{item.label}
								</Badge>
							{/each}
						{/each}
					</span>
				{:else}
					{#each items.filter(i => i.value === value) as item}
						<span class="font-medium">{item.label}</span>
					{/each}
				{/if}
			{:else}
				<span class="text-neutral-500">{placeholder}</span>
			{/if}
		</span>
		{#if open}
			<ChevronUp class="w-4 h-4 text-neutral-400" />
		{:else}
			<ChevronDown class="w-4 h-4 text-neutral-400" />
		{/if}
	</button>

	{#if open}
		<div
			class={`
				dropdown
				${align === 'right' ? 'right-0' : 'left-0'}
				${searchable ? 'max-h-[400px]' : 'max-h-60'}
			`}
			role="listbox"
			aria-multiselectable={multiple}
		>
			{#if searchable}
				<div class="p-2 border-b border-cyber-border">
					<Input
						type="search"
						placeholder="Buscar..."
						bind:value={searchQuery}
						class="w-full"
					/>
				</div>
			{/if}

			<div class="max-h-[300px] overflow-y-auto">
				{#each items.filter(item => !searchQuery || item.label.toLowerCase().includes(searchQuery.toLowerCase())) as item}
					<button
						type="button"
						role="option"
						aria-selected={value?.includes(item.value) || value === item.value}
						aria-disabled={item.disabled}
						data-value={item.value}
						class={`
							dropdown-item w-full
							${item.disabled ? 'opacity-50 cursor-not-allowed' : ''}
							${value?.includes(item.value) || value === item.value ? 'bg-neon-cyan/10 text-neon-cyan' : ''}
						`}
						on:click={() => selectItem(item)}
						disabled={item.disabled}
					>
						<div class="flex items-center gap-2">
							{#if item.icon}
								<svelte:component this={item.icon} class="w-4 h-4" />
							{/if}
							<span class="flex-1 text-left truncate">{item.label}</span>
							{#if value?.includes(item.value) || value === item.value}
								<Check class="w-4 h-4 text-neon-cyan" />
							{/if}
						</div>
					</button>
				{/each}

				{#if items.filter(item => !searchQuery || item.label.toLowerCase().includes(searchQuery.toLowerCase())).length === 0}
					<div class="dropdown-item text-center text-neutral-500 py-4">
						Nenhum resultado encontrado
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>