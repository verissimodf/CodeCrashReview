<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { auth, ui } from '$lib/stores';
	import { cn } from '$lib/utils';
	import {
		LayoutDashboard,
		Users,
		Trophy,
		Code,
		BookOpen,
		User,
		Settings,
		Menu,
		X,
		Moon,
		Sun,
		Bell,
		LogOut,
		ChevronDown
	} from 'lucide-svelte';

	const navigation = [
		{ href: '/', label: 'Dashboard', icon: LayoutDashboard, exact: true },
		{ href: '/arena', label: 'Arena', icon: Users },
		{ href: '/leaderboard', label: 'Ranking', icon: Trophy },
		{ href: '/challenges', label: 'Desafios', icon: Code },
		{ href: '/snippets', label: 'Snippets', icon: BookOpen },
		{ href: '/profile', label: 'Perfil', icon: User }
	];

	let sidebarOpen = true;
	let userMenuOpen = false;
	let mobileMenuOpen = false;

	$: sidebarOpen = $ui.sidebar_open;

	function toggleSidebar() {
		ui.toggleSidebar();
	}

	function toggleUserMenu() {
		userMenuOpen = !userMenuOpen;
	}

	function toggleMobileMenu() {
		mobileMenuOpen = !mobileMenuOpen;
	}

	function handleLogout() {
		auth.clearAuth();
		userMenuOpen = false;
	}

	function toggleTheme() {
		const newTheme = $ui.theme === 'dark' ? 'light' : 'dark';
		ui.setTheme(newTheme);
	}

	// Close menus on route change
	$effect(() => {
		$page.url.pathname;
		userMenuOpen = false;
		mobileMenuOpen = false;
	});
</script>

<div class="min-h-screen bg-cyber-bg font-ui">
	<!-- Background Effects -->
	<div class="fixed inset-0 bg-grid-animate opacity-50 pointer-events-none" aria-hidden="true" />
	<div class="fixed inset-0 scanlines pointer-events-none" aria-hidden="true" />

	<!-- Mobile Menu Button -->
	<button
		class="fixed top-4 left-4 z-50 lg:hidden btn-ghost btn-sm p-2"
		on:click={toggleMobileMenu}
		aria-label="Menu"
	>
		{#if mobileMenuOpen}
			<X class="w-5 h-5" />
		{:else}
			<Menu class="w-5 h-5" />
		{/if}
	</button>

	<!-- Sidebar -->
	<aside
		class={cn(
			'fixed top-0 left-0 h-screen z-40 glass-strong border-r border-cyber-border transition-all duration-normal',
			sidebarOpen ? 'w-64' : 'w-20',
			'lg:translate-x-0'
		)}
	>
		<!-- Logo -->
		<div class="flex items-center gap-3 h-16 px-4 border-b border-cyber-border">
			<div class="flex items-center justify-center w-10 h-10 rounded-lg bg-neon-cyan/10 border border-neon-cyan/30">
				<Code class="w-5 h-5 text-neon-cyan" />
			</div>
			{#if sidebarOpen}
				<div class="flex flex-col overflow-hidden">
					<span class="font-display text-xl font-bold text-gradient text-gradient-cyan">CodeCrash</span>
					<span class="font-display text-xs font-medium text-neon-cyan/70">Review</span>
				</div>
			{/if}
		</div>

		<!-- Navigation -->
		<nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
			{#each navigation as item}
				<a
					href={item.href}
					class={cn(
						'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-fast',
						'hover:bg-cyber-bgTerritary hover:text-neon-cyan',
						$page.url.pathname === item.href || (item.exact && $page.url.pathname.startsWith(item.href + '/'))
							? 'bg-neon-cyan/10 text-neon-cyan border border-neon-cyan/30'
							: 'text-neutral-400'
					)}
				>
					<item.icon class="w-5 h-5 flex-shrink-0" />
					{#if sidebarOpen}<span>{item.label}</span>{/if}
				</a>
			{/each}
		</nav>

		<!-- User Info (collapsed) -->
		{#if $auth.isAuthenticated && sidebarOpen}
			<div class="absolute bottom-0 left-0 right-0 p-4 border-t border-cyber-border">
				<div class="flex items-center gap-3">
					<div class="w-10 h-10 rounded-full bg-neon-cyan/10 border border-neon-cyan/30 flex items-center justify-center">
						{#if $auth.user?.avatar_url}
							<img src={$auth.user.avatar_url} alt="" class="w-full h-full rounded-full" />
						{:else}
							<span class="font-display text-neon-cyan font-bold">
								{$auth.user?.username?.charAt(0).toUpperCase()}
							</span>
						{/if}
					</div>
					<div class="flex-1 min-w-0">
						<p class="text-sm font-medium text-neutral-100 truncate">{$auth.user?.username}</p>
						<p class="text-xs text-neutral-500">{formatRating($auth.user?.rating)} • {getRatingTier($auth.user?.rating).name}</p>
					</div>
				</div>
			</div>
		{/if}
	</aside>

	<!-- Mobile Overlay -->
	{#if mobileMenuOpen}
		<div class="fixed inset-0 bg-black/50 z-30 lg:hidden" on:click={toggleMobileMenu} />
	{/if}

	<!-- Main Content -->
	<main class={cn('transition-all duration-normal lg:ml-64', sidebarOpen ? 'lg:ml-64' : 'lg:ml-20')}>
		<!-- Top Bar -->
		<header class="sticky top-0 z-30 glass-strong border-b border-cyber-border">
			<div class="flex items-center justify-between h-16 px-6">
				<div class="flex items-center gap-4">
					<h1 class="font-display text-xl font-bold text-gradient text-gradient-cyan">
						{$page.url.pathname.split('/')[1]?.charAt(0).toUpperCase() + $page.url.pathname.split('/')[1]?.slice(1) || 'Dashboard'}
					</h1>
				</div>

				<div class="flex items-center gap-3">
					<!-- Theme Toggle -->
					<button
						class="btn-ghost btn-sm p-2"
						on:click={toggleTheme}
						aria-label="Alternar tema"
					>
						{#if $ui.theme === 'dark'}
							<Sun class="w-5 h-5 text-neon-amber" />
						{:else}
							<Moon class="w-5 h-5 text-neon-blue" />
						{/if}
					</button>

					<!-- Notifications -->
					<button class="btn-ghost btn-sm p-2 relative" aria-label="Notificações">
						<Bell class="w-5 h-5" />
						{#if $ui.notifications.filter(n => !n.read).length > 0}
							<span class="absolute -top-1 -right-1 w-4 h-4 bg-neon-red rounded-full text-[10px] font-bold flex items-center justify-center">
								{$ui.notifications.filter(n => !n.read).length > 9 ? '9+' : $ui.notifications.filter(n => !n.read).length}
							</span>
						{/if}
					</button>

					<!-- User Menu -->
					{#if $auth.isAuthenticated}
						<div class="relative">
							<button
								class="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-cyber-bgTerritary transition-colors"
								on:click={toggleUserMenu}
							>
								<div class="w-8 h-8 rounded-full bg-neon-cyan/10 border border-neon-cyan/30 flex items-center justify-center">
									{#if $auth.user?.avatar_url}
										<img src={$auth.user.avatar_url} alt="" class="w-full h-full rounded-full" />
									{:else}
										<span class="font-display text-neon-cyan font-bold text-sm">
											{$auth.user?.username?.charAt(0).toUpperCase()}
										</span>
									{/if}
								</div>
								<span class="hidden sm:block text-sm font-medium">{$auth.user?.username}</span>
								<ChevronDown class="w-4 h-4 text-neutral-400" />
							</button>

							{#if userMenuOpen}
								<div class="absolute right-0 mt-2 w-48 glass-strong rounded-lg shadow-xl border border-cyber-border py-1 z-50 animate-slide-down">
									<a href="/profile" class="dropdown-item flex items-center gap-2">
										<User class="w-4 h-4" />
										Perfil
									</a>
									<a href="/settings" class="dropdown-item flex items-center gap-2">
										<Settings class="w-4 h-4" />
										Configurações
									</a>
									<hr class="my-1 border-cyber-border" />
									<button class="dropdown-item flex items-center gap-2 w-full text-left" on:click={handleLogout}>
										<LogOut class="w-4 h-4" />
										Sair
									</button>
								</div>
							{/if}
						</div>
					{:else}
						<div class="flex items-center gap-2">
							<a href="/auth/login" class="btn-ghost btn-sm">Entrar</a>
							<a href="/auth/register" class="btn-primary btn-sm">Registrar</a>
						</div>
					{/if}
				</div>
			</div>
		</header>

		<!-- Page Content -->
		<div class="p-6">
			@render()
		</div>
	</main>

	<!-- Toasts -->
	<div class="fixed bottom-6 right-6 z-50 flex flex-col gap-2">
		{#each $ui.toasts as toast (toast.id)}
			<div class={cn(
				'glass-strong rounded-lg border p-4 min-w-[300px] max-w-md shadow-xl animate-slide-up flex items-start gap-3',
				'border-l-4',
				toast.type === 'success' && 'border-neon-green',
				toast.type === 'error' && 'border-neon-red',
				toast.type === 'warning' && 'border-neon-amber',
				toast.type === 'info' && 'border-neon-cyan'
			)}>
				<div class="flex-1">
					<p class="font-medium text-neutral-100">{toast.title}</p>
					{#if toast.message}
						<p class="text-sm text-neutral-400 mt-1">{toast.message}</p>
					{/if}
				</div>
				<button class="text-neutral-400 hover:text-neon-cyan" on:click={() => $ui.dismissToast(toast.id)}>
					<X class="w-4 h-4" />
				</button>
			</div>
		{/if}
	</div>
</div>

<script>
	// Helper functions for template
	function formatRating(rating: number): string {
		return rating.toLocaleString('pt-BR');
	}

	function getRatingTier(rating: number): { name: string; color: string } {
		if (rating >= 2400) return { name: 'Grandmaster', color: '#ff3366' };
		if (rating >= 2200) return { name: 'Master', color: '#ff6b6b' };
		if (rating >= 2000) return { name: 'Diamond', color: '#b9f2ff' };
		if (rating >= 1800) return { name: 'Platinum', color: '#e5e4e2' };
		if (rating >= 1600) return { name: 'Gold', color: '#ffd700' };
		if (rating >= 1400) return { name: 'Silver', color: '#c0c0c0' };
		if (rating >= 1200) return { name: 'Bronze', color: '#cd7f32' };
		return { name: 'Unranked', color: '#8b949e' };
	}
</script>