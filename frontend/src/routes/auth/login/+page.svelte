<script lang="ts">
	import { auth } from '$lib/stores';
	import { goto } from '$app/navigation';
	import { Mail, Lock, AlertCircle, CheckCircle, Loader2, Eye, EyeOff } from 'lucide-svelte';

	let email = '';
	let password = '';
	let showPassword = false;
	let error = '';
	let loading = false;

	let emailError = '';
	let passwordError = '';

	function validateEmail(email: string): boolean {
		return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
	}

	function validateForm(): boolean {
		emailError = '';
		passwordError = '';

		if (!email.trim()) {
			emailError = 'Email é obrigatório';
		} else if (!validateEmail(email)) {
			emailError = 'Email inválido';
		}

		if (!password) {
			passwordError = 'Senha é obrigatória';
		} else if (password.length < 8) {
			passwordError = 'Senha deve ter pelo menos 8 caracteres';
		}

		return !emailError && !passwordError;
	}

	async function handleSubmit() {
		if (!validateForm()) return;

		loading = true;
		error = '';

		try {
			await auth.login(email, password);
			await goto('/');
		} catch (err: any) {
			error = err.message || 'Erro ao fazer login';
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center p-4">
	<div class="w-full max-w-md">
		<!-- Background decoration -->
		<div class="absolute inset-0 bg-grid-animate opacity-20 pointer-events-none" />

		<div class="relative card p-8">
			<div class="text-center mb-8">
				<div class="w-16 h-16 mx-auto mb-4 rounded-2xl flex items-center justify-center font-display text-2xl font-bold text-cyber-bg"
					style="background: linear-gradient(135deg, var(--tw-color-neon-cyan), var(--tw-color-neon-magenta));">
					<Code class="w-8 h-8" />
				</div>
				<h1 class="font-display text-2xl font-bold text-neutral-100">Bem-vindo de volta</h1>
				<p class="text-neutral-400 mt-2">Entre na sua conta CodeCrashReview</p>
			</div>

			{#if error}
				<div class="mb-6 p-4 bg-neon-red/10 border border-neon-red/30 rounded-lg flex items-center gap-3 text-neon-red">
					<AlertCircle class="w-5 h-5 flex-shrink-0" />
					<p class="text-sm">{error}</p>
				</div>
			{/if}

			<form on:submit|preventDefault={handleSubmit} class="space-y-5">
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
							bind:value={password}
							class="input pl-10 pr-10"
							placeholder="••••••••"
							autocomplete="current-password"
							disabled={loading}
						/>
						<button
							type="button"
							class="absolute right-3 top-1/2 -translate-y-1/2 text-neutral-500 hover:text-neon-cyan"
							on:click={() => showPassword = !showPassword}
						>
							{showPassword ? <EyeOff class="w-5 h-5" /> : <Eye class="w-5 h-5" />}
						</button>
					</div>
					{#if passwordError}
						<p class="mt-1 text-sm text-neon-red flex items-center gap-1">
							<AlertCircle class="w-3 h-3" />
							{passwordError}
						</p>
					{/if}
				</div>

				<div class="flex items-center justify-between">
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" class="w-4 h-4 accent-neon-cyan rounded" />
						<span class="text-sm text-neutral-400">Lembrar-me</span>
					</label>
					<a href="/auth/forgot-password" class="text-sm text-neon-cyan hover:text-neon-cyan/70">
						Esqueci a senha
					</a>
				</div>

				<button
					type="submit"
					class="btn-primary btn-lg w-full"
					disabled={loading}
				>
					{loading ? (
						<>
							<Loader2 class="w-5 h-5 mr-2 animate-spin" />
							Entrando...
						</>
					) : (
						'Entrar'
					)}
				</button>
			</form>

			<div class="mt-6 text-center">
				<p class="text-neutral-400">
					Não tem conta?
					<a href="/auth/register" class="text-neon-cyan hover:text-neon-cyan/70 font-medium ml-1">
						Criar conta
					</a>
				</p>
			</div>

			<div class="mt-8 pt-6 border-t border-cyber-border text-center">
				<p class="text-neutral-500 text-sm mb-3">Ou continue com</p>
				<div class="flex items-center justify-center gap-3">
					<button class="btn-secondary w-12 h-12" disabled>
						<svg class="w-6 h-6" viewBox="0 0 24 24">
							<path fill="currentColor" d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z"/>
						</svg>
					</button>
					<button class="btn-secondary w-12 h-12" disabled>
						<svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
							<path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.482 0 12.64 12.64 0 0 0-.617-1.25 19.791 19.791 0 0 0-4.886 1.514A.061.061 0 0 0 2 4.45v15.218c0 .147.072.266.18.338a20.158 20.158 0 0 0 5.547 2.124 12.5 12.5 0 0 0 1.558-.284 13.377 13.377 0 0 0 .219-.459 13.148 13.148 0 0 0 1.35-3.506c.597 0 1.233-.072 1.787-.218a.055.055 0 0 1 .055.056v.286c-.134.524-.207 1.076-.207 1.64v2.075c0 .141.067.257.18.337a20.07 20.07 0 0 0 5.457-2.122.057.057 0 0 0 .18-.336V4.45a.061.061 0 0 0-.068-.054z"/>
						</svg>
					</button>
				</div>
			</div>
		</div>
	</div>
</div>