import { get } from 'svelte/store';
import { auth } from '$lib/auth';

export interface ArenaChallenge {
    id: string;
    title: string;
    description: string;
    starter_code: string;
}

export interface ArenaMatch {
    id: string;
    status: string;
    challenge_id: string;
    winner_id: string | null;
    participants: {user_id: string; username: string; persona: string; score: number | null; status: string}[];
    submissions: {username: string; score: number; feedback: string; files: Record<string,string>}[];
}

export const personas = ['sentinel', 'optimizer', 'refactor', 'guardian', 'docubot', 'polyglot'];
export const statusLabel: Record<string,string> = {
    waiting: 'Aguardando adversário', pending: 'Aguardando adversário', active: 'Em andamento',
    in_progress: 'Em andamento', completed: 'Encerrada', submitted: 'Entregue', joined: 'Na partida'
};

export async function request<T>(path: string, body?: unknown): Promise<T> {
    const token = get(auth).tokens?.access_token;
    if (!token) throw new Error('Entre na sua conta para continuar.');
    const response = await fetch('/api/' + path, {
        method: body === undefined ? 'GET' : 'POST',
        headers: {Authorization: 'Bearer ' + token, ...(body === undefined ? {} : {'Content-Type': 'application/json'})},
        body: body === undefined ? undefined : JSON.stringify(body)
    });
    const payload = await response.json().catch(() => null);
    if (!response.ok) throw new Error(payload?.error?.message || payload?.error || 'Falha na solicitação (' + response.status + ')');
    if (payload === null) throw new Error('O servidor retornou uma resposta inválida.');
    return payload.data ?? payload;
}

export function errorText(error: unknown) {
    return error instanceof Error ? error.message : String(error);
}
