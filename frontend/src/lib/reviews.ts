import { get } from 'svelte/store';
import { auth } from '$lib/auth';

export const personas = ['sentinel', 'optimizer', 'refactor', 'guardian', 'docubot', 'polyglot'];
export const reviewStatus: Record<string, string> = {
    queued: 'Na fila', acquiring: 'Criando snapshot', analyzing: 'Dois agentes analisando',
    evaluating: 'Juiz avaliando', awaiting_human: 'Aguardando veredito humano',
    completed: 'Concluída', failed: 'Falhou'
};
export interface ReviewSummary {
    id: string; source_type: 'github'|'local'; source_location: string; source_branch: string;
    objective: string; review_focus: string; agent_a: string; agent_b: string; evaluator_mode: 'ai'|'human'; status: string;
    snapshot_sha: string|null; snapshot_files: number; snapshot_bytes: number; error_message: string|null;
    created_at: string; started_at: string|null; completed_at: string|null;
}
export interface Finding { severity: string; category: string; file: string; line: number; title: string; evidence: string; recommendation: string; }
export interface ReviewDelivery {
    topology: string; design: string; deployment_plan: string[]; approval_criteria: string[];
    final_objective: string; methodology: string[]; effectiveness_evidence: string[]; executed_steps: string[];
}
export interface ProposalVerification {
    proof?: {profile: string; passed: boolean; exit_code: number; source_sha256: string; tests_sha256: string; image: string; output: string; test_count?:number; tests_preserved?:boolean; baseline?:{passed:boolean; exit_code:number; output:string; test_count:number}};
    patch_applicable: boolean; exact_edits_verified: boolean; functional_tests_run: boolean;
    quality_state: string; system_evidence: string[]; unsupported_claims: string[];
}
export interface ReviewProposal {
    id: string; slot: 'A'|'B'; persona: string; model: string; status: string; summary: string|null;
    findings: Finding[]; patch: string; patch_valid: boolean; validation_error: string|null; confidence: number;
    tokens_used: number; latency_ms: number; error_message: string|null;
    delivery: ReviewDelivery;
    verification: ProposalVerification;
}
export interface ReviewVerdict { evaluator_type: 'ai'|'human'; winner: 'A'|'B'|'TIE'; rationale: string; scores: Record<string, unknown>; model: string|null; created_at: string; }
export interface ReviewDetail extends ReviewSummary { focus_paths: string[]; proposals: ReviewProposal[]; verdict: ReviewVerdict|null; }
export async function reviewRequest<T>(path = '', body?: unknown): Promise<T> {
    const token = get(auth).tokens?.access_token;
    if (!token) throw new Error('Entre na sua conta para continuar.');
    const response = await fetch('/api/reviews' + path, {
        method: body === undefined ? 'GET' : 'POST',
        headers: {Authorization: 'Bearer ' + token, ...(body === undefined ? {} : {'Content-Type': 'application/json'})},
        body: body === undefined ? undefined : JSON.stringify(body)
    });
    const payload = await response.json().catch(() => null);
    if (!response.ok) throw new Error(payload?.error || `Falha na solicitação (${response.status})`);
    return payload as T;
}
export function reviewError(error: unknown) { return error instanceof Error ? error.message : String(error); }
