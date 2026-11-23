import { test, expect } from '@playwright/test';
import { readFileSync, writeFileSync } from 'node:fs';

test('public interface, real login, persisted proof and downloadable correction', async ({ page }) => {
    const session = JSON.parse(readFileSync('../artifacts/validation/.session.json', 'utf8'));
    const review = JSON.parse(readFileSync('../artifacts/validation/review-e2e.json', 'utf8'));
    expect(review.status).toBe('completed');
    const errors: string[] = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.goto('/');
    await expect(page.getByRole('heading', {level: 1})).toContainText('Duas leituras');
    await page.getByRole('link', {name:'Entrar na arena'}).click();
    await page.getByLabel('E-mail', {exact:true}).fill(session.account.email);
    await page.getByLabel('Senha', {exact:true}).fill(session.account.password);
    await page.getByRole('button', {name:'Entrar',exact:true}).click();
    await expect(page).toHaveURL(/\/arena$/);
    await expect(page.getByRole('heading', {name:'Nova revisão'})).toBeVisible();
    await page.locator(`a[href="/arena/${review.id}"]`).click();
    await expect(page.getByRole('heading', {name:/Veredito:/})).toBeVisible({timeout: 30_000});
    await expect(page.getByText('testes funcionais aprovados', {exact:false}).first()).toBeVisible();
    if (review.verdict.winner !== 'TIE') {
        const downloading = page.waitForEvent('download');
        await page.getByRole('button', {name:'Baixar patch vencedor'}).click();
        const download = await downloading;
        const path = '../artifacts/validation/browser-winner.patch';
        await download.saveAs(path);
        const selected = review.proposals.find((p: {slot:string}) => p.slot === review.verdict.winner);
        expect(readFileSync(path, 'utf8').trim()).toBe(selected.patch.trim());
    }
    await page.evaluate(() => window.scrollTo(0,0));
    await page.screenshot({path:'../artifacts/validation/review-desktop.png', fullPage:true});
    await page.screenshot({path:'../artifacts/validation/review-viewport.png'});
    await page.setViewportSize({width:390,height:844});
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
    await page.evaluate(() => window.scrollTo(0,0));
    await page.screenshot({path:'../artifacts/validation/review-mobile.png', fullPage:true});
    await page.screenshot({path:'../artifacts/validation/review-mobile-viewport.png'});
    await page.getByRole('button', {name:/Sair/}).click();
    await expect(page).toHaveURL(/\/auth\/login$/);
    expect(errors).toEqual([]);
    writeFileSync('../artifacts/validation/browser-summary.json', JSON.stringify({review_id:review.id, login:true, evidence_visible:true, patch_download:review.verdict.winner!=='TIE', mobile_no_overflow:true, javascript_errors:errors},null,2));
});
