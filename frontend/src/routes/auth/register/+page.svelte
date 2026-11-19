<script lang="ts">
	import { auth } from '$lib/stores';
	import { goto } from '$app/navigation';
	import { Code, Mail, Lock, User, AlertCircle, CheckCircle, Loader2, Eye, EyeOff } from 'lucide-svelte';

	let username = '';
	let email = '';
	let password = '';
	let confirmPassword = '';
	let showPassword = false;
	let error = '';
	let loading = false;

	let usernameError = '';
	let emailError = '';
	let passwordError = '';
	let confirmPasswordError = '';

	function validateEmail(email: string): boolean {
		return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
	}

	function validateUsername(username: string): boolean {
		return /^[a-zA-Z0-9_-]{3,20}$/.test(username);
	}

	function validatePassword(password: string): { valid: boolean; errors: string[] } {
		const errors: string[] = [];
		if (password.length < 8) errors.push('Mínimo 8 caracteres');
		if (!/[A-Z]/.test(password)) errors.push('Pelo menos uma maiúscula');
		if (!/[a-z]/.test(password)) errors.push('Pelo menos uma minúscula');
		if (!/[0-9]/.test(password)) errors.push('Pelo menos um número');
		if (!/[!@#$%^&*]/.test(password)) errors.push('Pelo menos um caractere especial');
		return { valid: errors.length === 0, errors };
	}

	function validateForm(): boolean {
		usernameError = '';
		emailError = '';
		passwordError = '';
		confirmPasswordError = '';

		if (!username.trim()) {
			usernameError = 'Nome de usuário é obrigatório';
		} else if (!validateUsername(username)) {
			usernameError = '3-20 caracteres, apenas letras, números, _ e -';
		}

		if (!email.trim()) {
			emailError = 'Email é obrigatório';
		} else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
			emailError = 'Email inválido';
		}

		const passwordValidation = validatePassword(password);
		if (!passwordValidation.valid) {
			passwordError = passwordValidation.errors.join(', ');
		}

		if (password !== confirmPassword) {
			confirmPasswordError = 'As senhas não coincidem';
		}

		return !usernameError && !emailError && !passwordError && !confirmPasswordError;
	}

	async function handleSubmit() {
		if (!validateForm()) return;

		loading = true;
		error = '';

		try {
			await auth.register(username, email, password);
			await goto('/arena');
		} catch (err: any) {
			error = err.message || 'Erro ao registrar';
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center p-4">
	<div class="w-full max-w-md">
		<div class="absolute inset-0 bg-grid-animate opacity-20 pointer-events-none" />

		<div class="relative card p-8">
			<div class="text-center mb-8">
				<div class="w-16 h-16 mx-auto mb-4 rounded-2xl flex items-center justify-center font-display text-2xl font-bold text-cyber-bg"
					style="background: linear-gradient(135deg, var(--tw-color-neon-cyan), var(--tw-color-neon-magenta));">
					<Code class="w-8 h-8" />
				</div>
				<h1 class="font-display text-2xl font-bold text-neutral-100">Crie sua conta</h1>
				<p class="text-neutral-400 mt-2">Junte-se à arena CodeCrashReview</p>
			</div>

			{#if error}
				<div class="mb-6 p-4 bg-neon-red/10 border border-neon-red/30 rounded-lg flex items-center gap-3 text-neon-red">
					<AlertCircle class="w-5 h-5 flex-shrink-0" />
					<p class="text-sm">{error}</p>
				</div>
			{/if}

			<form on:submit|preventDefault={handleSubmit} class="space-y-5">
				<div>
					<label for="username" class="label">Nome de Usuário</label>
					<div class="relative">
						<User class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-neutral-500" />
						<input
							type="text"
							id="username"
							bind:value={username}
							class="input pl-10"
							placeholder="usuario123"
							autocomplete="username"
							disabled={loading}
						/>
					</div>
					{#if usernameError}
						<p class="mt-1 text-sm text-neon-red flex items-center gap-1">
							<AlertCircle class="w-3 h-3" />
							{usernameError}
						</p>
					{/if}
				</div>

				<div>
					<label for="email" class="label">Email</label>
					<div class="relative">
						<Mail class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-neutral-500" />
						<input
							type="email"
							id="email"
							bind:value={email}
							class="input pl-10"
							placeholder="seu@email.com"
							autocomplete="email"
							disabled={loading}
						/>
					</div>
					{#if emailError}
						<p class="mt-1 text-sm text-neon-red flex items-center gap-1">
							<AlertCircle class="w-3 h-3" />
							{emailError}
						</p>
					{/if}
				</div>

				<div>
					<label for="password" class="label">Senha</label>
					<div class="relative">
						<Lock class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-neutral-500" />
						<input
							type={showPassword ? 'text' : 'password'}
							id="password"
							value={password} on:input={event => password = event.currentTarget.value}
							class="input pl-10 pr-10"
							placeholder="••••••••"
							autocomplete="new-password"
							disabled={loading}
						/>
						<button
							type="button"
							class="absolute right-3 top-1/2 -translate-y-1/2 text-neutral-500 hover:text-neon-cyan"
							on:click={() => showPassword = !showPassword}
						>
							{#if showPassword}<EyeOff class="w-5 h-5" />{:else}<Eye class="w-5 h-5" />{/if}
						</button>
					</div>
					{#if passwordError}
						<p class="mt-1 text-sm text-neon-red flex items-center gap-1">
							<AlertCircle class="w-3 h-3" />
							{passwordError}
						</p>
					{/if}
				</div>

				<div>
					<label for="confirmPassword" class="label">Confirmar Senha</label>
					<div class="relative">
						<Lock class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-neutral-500" />
						<input
							type={showPassword ? 'text' : 'password'}
							id="confirmPassword"
							value={confirmPassword} on:input={event => confirmPassword = event.currentTarget.value}
							class="input pl-10"
							placeholder="••••••••"
							autocomplete="new-password"
							disabled={loading}
						/>
					</div>
					{#if confirmPasswordError}
						<p class="mt-1 text-sm text-neon-red flex items-center gap-1">
							<AlertCircle class="w-3 h-3" />
							{confirmPasswordError}
						</p>
					{/if}
				</div>

				<div class="space-y-2 text-sm text-neutral-400">
					<p class="flex items-center gap-1"><CheckCircle class="w-4 h-4" /> Mínimo 8 caracteres</p>
					<p class="flex items-center gap-1"><CheckCircle class="w-4 h-4" /> Pelo menos uma maiúscula</p>
					<p class="flex items-center gap-1"><CheckCircle class="w-4 h-4" /> Pelo menos uma minúscula</p>
					<p class="flex items-center gap-1"><CheckCircle class="w-4 h-4" /> Pelo menos um número</p>
					<p class="flex items-center gap-1"><CheckCircle class="w-4 h-4" /> Pelo menos um caractere especial (!@#$%^&*)</p>
				</div>

				<button
					type="submit"
					class="btn-primary btn-lg w-full"
					disabled={loading}
				>
					{#if loading}
						<Loader2 class="w-5 h-5 mr-2 animate-spin" />
							Criando conta...
					{:else}
						Criar Conta
					{/if}
				</button>
			</form>

			<div class="mt-6 text-center">
				<p class="text-neutral-400">
					Já tem conta?
					<a href="/auth/login" class="text-neon-cyan hover:text-neon-cyan/70 font-medium ml-1">
						Entrar
					</a>
				</p>
			</div>
		</div>
	</div>
</div>
