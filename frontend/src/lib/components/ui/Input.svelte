<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLInputAttributes, HTMLTextareaAttributes, HTMLSelectAttributes } from 'svelte/elements';

	export let label: string;
	export let error: string;
	export let hint: string;
	export let disabled = false;
	export let required = false;
	export let type: 'text' | 'email' | 'password' | 'number' | 'url' | 'tel' = 'text';
	export let placeholder = '';
	export let value = '';
	export let name: string;
	export let id: string;
	export let autocomplete: string;
	export let rows = 4;
	export let as: 'input' | 'textarea' | 'select' = 'input';
	export let options: { value: string; label: string }[] = [];

	$$restProps: HTMLInputAttributes | HTMLTextareaAttributes | HTMLSelectAttributes;

	const inputId = id || name || `input-${Math.random().toString(36).substr(2, 9)}`;
	const errorId = error ? `${inputId}-error` : undefined;
	const hintId = hint ? `${inputId}-hint` : undefined;

	const describedBy = [errorId, hintId].filter(Boolean).join(' ') || undefined;

	$: inputClasses = `
		w-full px-4 py-2.5 bg-cyber-bgSecondary border rounded-lg text-neutral-100 placeholder-neutral-500
		transition-all duration-fast focus:border-neon-cyan focus:ring-2 focus:ring-neon-cyan/20 focus:outline-none
		disabled:bg-cyber-bgTerritary disabled:text-neutral-500 disabled:cursor-not-allowed
		${error ? 'border-neon-red focus:border-neon-red focus:ring-neon-red/20' : 'border-cyber-border'}
	`;
</script>

<div class="w-full">
	{#if label}
		<label for={inputId} class="label">
			{label}
			{#if required}
				<span class="text-neon-red ml-1" aria-hidden="true">*</span>
			{/if}
		</label>
	{/if}

	{#if as === 'textarea'}
		<textarea
			{id}
			{name}
			{value}
			{placeholder}
			{disabled}
			{required}
			{rows}
			{autocomplete}
			{describedBy}
			class={inputClasses}
			{...$$restProps}
		/>
	{:else if as === 'select'}
		<select
			{id}
			{name}
			{value}
			{disabled}
			{required}
			{describedBy}
			class={inputClasses}
			{...$$restProps}
		>
			{#if placeholder}
				<option value="" disabled selected>{placeholder}</option>
			{/if}
			{#each options as option}
				<option value={option.value}>{option.label}</option>
			{/each}
		</select>
	{:else}
		<input
			type={type}
			{id}
			{name}
			{value}
			{placeholder}
			{disabled}
			{required}
			{autocomplete}
			{describedBy}
			class={inputClasses}
			{...$$restProps}
		/>
	{/if}

	{#if error}
		<p id={errorId} class="mt-1.5 text-sm text-neon-red flex items-center gap-1">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
			</svg>
			{error}
		</p>
	{:else if hint}
		<p id={hintId} class="mt-1.5 text-sm text-neutral-500">{hint}</p>
	{/if}
</div>