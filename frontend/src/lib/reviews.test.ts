import { beforeEach, expect, test, vi } from 'vitest';
const state = vi.hoisted(() => ({tokens: null as {access_token:string}|null}));
vi.mock('$lib/auth', () => ({auth: {subscribe: (fn: (value:typeof state) => void) => {fn(state); return () => {};}}}));
import { reviewRequest } from './reviews';

beforeEach(() => { state.tokens=null; vi.unstubAllGlobals(); });
test('an unauthenticated review never sends a request', async () => {
    const fetch = vi.fn(); vi.stubGlobal('fetch',fetch);
    await expect(reviewRequest()).rejects.toThrow('Entre');
    expect(fetch).not.toHaveBeenCalled();
});
test('API refusal is shown instead of reporting success', async () => {
    state.tokens={access_token:'fixture'};
    vi.stubGlobal('fetch',vi.fn().mockResolvedValue({ok:false,status:409,json:async()=>({error:'Functional validation required'})}));
    await expect(reviewRequest('/id/verdict',{winner:'A'})).rejects.toThrow('Functional validation required');
});
test('persisted proof and scorecard reach the interface unchanged', async () => {
    state.tokens={access_token:'fixture'};
    const data={id:'fixture',proposals:[{verification:{proof:{passed:true,exit_code:0}}}],verdict:{winner:'A',scores:{A:{total:90},B:{total:70}}}};
    vi.stubGlobal('fetch',vi.fn().mockResolvedValue({ok:true,json:async()=>data}));
    expect(await reviewRequest('/fixture')).toEqual(data);
});
