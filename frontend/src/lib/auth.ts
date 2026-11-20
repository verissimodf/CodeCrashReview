import { browser } from '$app/environment';
import { writable } from 'svelte/store';

export interface User {
    id: string;
    username: string;
    email: string;
}

export interface AuthTokens {
    access_token: string;
    refresh_token: string;
    expires_in: number;
    token_type: string;
}

interface AuthState {
    user: User|null;
    tokens: AuthTokens|null;
    isAuthenticated: boolean;
    isLoading: boolean;
    error: string|null;
}

function createAuthStore() {
    const initial: AuthState = {user: null, tokens: null, isAuthenticated: false, isLoading: browser, error: null};
    const {subscribe, set} = writable(initial);

    function clearAuth() {
        if (browser) {
            localStorage.removeItem('ccr_user');
            localStorage.removeItem('ccr_tokens');
        }
        set({...initial, isLoading: false});
    }

    function setAuth(user: User, tokens: AuthTokens) {
        if (browser) {
            localStorage.setItem('ccr_user', JSON.stringify(user));
            localStorage.setItem('ccr_tokens', JSON.stringify(tokens));
        }
        set({user, tokens, isAuthenticated: true, isLoading: false, error: null});
    }

    async function authenticate(path: 'login'|'register', fields: Record<string,string>) {
        const response = await fetch('/api/auth/' + path, {
            method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(fields)
        });
        const payload = await response.json().catch(() => ({}));
        if (!response.ok) throw new Error(payload.error || 'Falha na autenticação.');
        const session = payload.data ?? payload;
        setAuth(session.user, session.tokens);
    }

    async function refreshToken(refreshToken: string) {
        try {
            const response = await fetch('/api/auth/refresh', {
                method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({refresh_token: refreshToken})
            });
            const payload = await response.json();
            if (!response.ok) return clearAuth();
            const session = payload.data ?? payload;
            setAuth(session.user, session.tokens);
        } catch { clearAuth(); }
    }

    if (browser) {
        try {
            const user = JSON.parse(localStorage.getItem('ccr_user') || 'null') as User|null;
            const tokens = JSON.parse(localStorage.getItem('ccr_tokens') || 'null') as AuthTokens|null;
            if (!user || !tokens) clearAuth();
            else {
                const part = tokens.access_token.split('.')[1];
                if (!part) clearAuth();
                else {
                    const payload = JSON.parse(atob(part.replace(/-/g, '+').replace(/_/g, '/')));
                    if (payload.exp * 1000 > Date.now()) setAuth(user, tokens);
                    else void refreshToken(tokens.refresh_token);
                }
            }
        } catch { clearAuth(); }
    }

    return {
        subscribe, clearAuth,
        login: (email: string, password: string) => authenticate('login', {email, password}),
        register: (username: string, email: string, password: string) => authenticate('register', {username, email, password})
    };
}

export const auth = createAuthStore();
