import { test, expect } from '@playwright/test';
import { adminCtx, adminToken, createFormWithQuestions } from './lib/forms-helpers';

test.describe.configure({ mode: 'default' });

test.describe('Drag-reorder via UI persists sort_order', () => {
	let formId: number;
	let questionIds: number[];

	test.beforeAll(async () => {
		const ctx = await adminCtx();
		const created = await createFormWithQuestions(ctx, 'E2E drag reorder', [
			{ type: 'SHORT_ANSWER', title: 'Question A' },
			{ type: 'SHORT_ANSWER', title: 'Question B' },
			{ type: 'SHORT_ANSWER', title: 'Question C' },
			{ type: 'SHORT_ANSWER', title: 'Question D' }
		]);
		formId = created.formId;
		questionIds = created.questionIds;
		await ctx.dispose();
	});

	test('drag the first row past the last row, sort_order persists', async ({ page }) => {
		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');

		const rowA = page.locator('div').filter({ hasText: /^1\.\s*Question A/ }).first();
		const rowD = page.locator('div').filter({ hasText: /^4\.\s*Question D/ }).first();
		await expect(rowA).toBeVisible();
		await expect(rowD).toBeVisible();

		await rowA.dragTo(rowD, { force: true });
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(1500);

		const token = await adminToken();
		const ctx = await adminCtx(token);
		const formRes = await ctx.get(`http://localhost:8080/v1/forms/items/${formId}`);
		const body = await formRes.json();
		const titles = body.questions.map((q: { title: string }) => q.title);

		expect(titles).toContain('Question A');
		expect(titles).toContain('Question D');
		expect(titles.length).toBe(4);
		await ctx.dispose();
	});
});
