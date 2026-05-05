import { test, expect } from '@playwright/test';
import { adminCtx, buildPublishedFormWithToken } from './lib/forms-helpers';

test.describe.configure({ mode: 'default' });

test.describe('Public renderer — every rich type renders correctly', () => {
	test.use({ storageState: { cookies: [], origins: [] } });

	let token: string;

	test.beforeAll(async () => {
		const ctx = await adminCtx();
		const built = await buildPublishedFormWithToken(ctx, 'E2E renderer-types', [
			{ type: 'EQUATION', title: 'Solve', extra: { latex: 'x^2 + 2x + 1 = 0' } },
			{
				type: 'FILL_IN_BLANK',
				title: 'Fill blanks',
				extra: { template: 'Hello [[1]] world [[2]] end' }
			},
			{
				type: 'TABLE',
				title: 'Monthly grid',
				extra: { rows: ['CP1', 'CP2'], columns: ['Sept', 'Oct'] }
			},
			{
				type: 'ORDERING',
				title: 'Order steps',
				options: [
					{ label: 'First', value: 'a' },
					{ label: 'Second', value: 'b' },
					{ label: 'Third', value: 'c' }
				]
			},
			{
				type: 'MATCHING',
				title: 'Match',
				extra: {
					pairs: [
						{ left: 'Lion', right: 'Mammal' },
						{ left: 'Eagle', right: 'Bird' }
					]
				}
			},
			{
				type: 'ADDRESS',
				title: 'Where',
				extra: { fields: ['Quartier', 'Commune'] }
			},
			{
				type: 'COUNTRY_REGION',
				title: 'Region',
				extra: { regions: ['Boké', 'Conakry', 'Faranah'] }
			},
			{ type: 'CODE_BLOCK', title: 'Code', extra: { language: 'python' } }
		]);
		token = built.token;
		await ctx.dispose();
	});

	test('EQUATION renders KaTeX HTML', async ({ page }) => {
		await page.goto(`/r/${token}`);
		await expect(page.locator('.katex').first()).toBeVisible();
	});

	test('FILL_IN_BLANK renders inline inputs from [[N]] template', async ({ page }) => {
		await page.goto(`/r/${token}`);
		const blanks = page.locator('input[name^="q_"][name*="_blank_"]');
		await expect(blanks.first()).toBeVisible();
		expect(await blanks.count()).toBeGreaterThanOrEqual(2);
	});

	test('TABLE renders a grid with headers and cells', async ({ page }) => {
		await page.goto(`/r/${token}`);
		await expect(page.locator('th:has-text("Sept")').first()).toBeVisible();
		await expect(page.locator('th:has-text("Oct")').first()).toBeVisible();
		await expect(page.locator('td:has-text("CP1")').first()).toBeVisible();
		const cells = page.locator('input[name*="_0_0"]');
		await expect(cells.first()).toBeVisible();
	});

	test('ORDERING renders position inputs per item', async ({ page }) => {
		await page.goto(`/r/${token}`);
		const pos = page.locator('input[name*="_pos_"]');
		expect(await pos.count()).toBeGreaterThanOrEqual(3);
	});

	test('MATCHING renders one select per left item with all rights', async ({ page }) => {
		await page.goto(`/r/${token}`);
		const selects = page.locator('select[name*="_match_"]');
		expect(await selects.count()).toBeGreaterThanOrEqual(2);
		await expect(page.locator('option:has-text("Mammal")').first()).toBeAttached();
		await expect(page.locator('option:has-text("Bird")').first()).toBeAttached();
	});

	test('ADDRESS renders one input per sub-field', async ({ page }) => {
		await page.goto(`/r/${token}`);
		await expect(page.locator('input[placeholder="Quartier"]').first()).toBeVisible();
		await expect(page.locator('input[placeholder="Commune"]').first()).toBeVisible();
	});

	test('COUNTRY_REGION renders Guinea + region options', async ({ page }) => {
		await page.goto(`/r/${token}`);
		await expect(page.locator('option:has-text("Guinée")').first()).toBeAttached();
		await expect(page.locator('option:has-text("Conakry")').first()).toBeAttached();
	});

	test('CODE_BLOCK renders a monospace textarea', async ({ page }) => {
		await page.goto(`/r/${token}`);
		const ta = page.locator('textarea[name^="q_"][placeholder*="python"]');
		await expect(ta.first()).toBeVisible();
	});
});
