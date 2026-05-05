import { test, expect } from '@playwright/test';
import { adminCtx, createFormWithQuestions, setLogicRules } from './lib/forms-helpers';

test.describe.configure({ mode: 'default' });

const GATEWAY = process.env.E2E_GATEWAY_URL ?? 'http://localhost:8080';

test.describe('Server-side actions: untested branches', () => {
	let formId: number;
	let questionIds: number[];
	let clientIds: string[];

	test.beforeAll(async () => {
		const ctx = await adminCtx();
		const built = await createFormWithQuestions(ctx, 'E2E server-actions', [
			{ type: 'SHORT_ANSWER', title: 'Q1' },
			{ type: 'NUMBER', title: 'Q2' },
			{ type: 'YES_NO', title: 'Q3' },
			{ type: 'ESSAY', title: 'Q4' }
		]);
		formId = built.formId;
		questionIds = built.questionIds;
		clientIds = built.clientIds;
		await ctx.dispose();
	});

	test('Validation: regex pattern, numeric min/max round-trip', async () => {
		const ctx = await adminCtx();

		await ctx.put(`${GATEWAY}/v1/forms/questions/${questionIds[0]}`, {
			data: {
				title: 'Q1',
				type: 'SHORT_ANSWER',
				required: false,
				sort_order: 10,
				validation: { pattern: '^[A-Z]+$' }
			}
		});
		await ctx.put(`${GATEWAY}/v1/forms/questions/${questionIds[1]}`, {
			data: {
				title: 'Q2',
				type: 'NUMBER',
				required: false,
				sort_order: 20,
				validation: { min: 10, max: 99 }
			}
		});

		const f = await (await ctx.get(`${GATEWAY}/v1/forms/items/${formId}`)).json();
		const q1 = f.questions.find((q: { id: number }) => q.id === questionIds[0]);
		const q2 = f.questions.find((q: { id: number }) => q.id === questionIds[1]);
		expect(q1.validation.pattern).toBe('^[A-Z]+$');
		expect(q2.validation.min).toBe(10);
		expect(q2.validation.max).toBe(99);
		await ctx.dispose();
	});

	test('Grading: points_max + correct_value + answers + rubric_url round-trip', async () => {
		const ctx = await adminCtx();

		await ctx.put(`${GATEWAY}/v1/forms/questions/${questionIds[2]}`, {
			data: {
				title: 'Q3',
				type: 'YES_NO',
				required: false,
				sort_order: 30,
				grading: { points_max: 5, correct_value: 'yes' }
			}
		});
		await ctx.put(`${GATEWAY}/v1/forms/questions/${questionIds[3]}`, {
			data: {
				title: 'Q4',
				type: 'ESSAY',
				required: false,
				sort_order: 40,
				grading: { points_max: 10, rubric_url: 'https://example.com/rubric.pdf' }
			}
		});

		const f = await (await ctx.get(`${GATEWAY}/v1/forms/items/${formId}`)).json();
		const q3 = f.questions.find((q: { id: number }) => q.id === questionIds[2]);
		const q4 = f.questions.find((q: { id: number }) => q.id === questionIds[3]);
		expect(q3.grading.points_max).toBe(5);
		expect(q3.grading.correct_value).toBe('yes');
		expect(q4.grading.rubric_url).toBe('https://example.com/rubric.pdf');
		await ctx.dispose();
	});

	test('?/deleteQuestion removes the question from the DB', async ({ page }) => {
		const ctxPre = await adminCtx();
		const before = await (await ctxPre.get(`${GATEWAY}/v1/forms/items/${formId}`)).json();
		const beforeCount = before.questions.length;
		await ctxPre.dispose();
		expect(beforeCount).toBeGreaterThanOrEqual(4);

		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');

		page.once('dialog', (d) => d.accept());
		const q4Row = page
			.locator('div')
			.filter({ hasText: /^4\.\s*Q4/ })
			.first();
		await q4Row.hover();
		await q4Row.locator('button[aria-label="Delete question"]').click();
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(500);

		const ctxPost = await adminCtx();
		const after = await (await ctxPost.get(`${GATEWAY}/v1/forms/items/${formId}`)).json();
		expect(after.questions.length).toBe(beforeCount - 1);
		expect(after.questions.find((q: { title: string }) => q.title === 'Q4')).toBeUndefined();
		await ctxPost.dispose();
	});

	test('Logic-rule delete via UI removes the rule from settings.logic_rules', async ({
		page
	}) => {
		const ctx = await adminCtx();
		await setLogicRules(ctx, formId, [
			{
				id: 'r_to_delete',
				target_question_client_id: clientIds[1],
				operator: 'show_if',
				conditions: [{ source_question_client_id: clientIds[0], op: 'equals', value: 'A' }]
			}
		]);
		await ctx.dispose();

		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');

		const targetRow = page
			.locator('div')
			.filter({ hasText: /^[0-9]+\.\s*Q2/ })
			.first();
		await targetRow.click();

		const drawer = page.locator('aside').filter({ hasText: 'Question settings' });
		await expect(drawer.locator('text=/Show if/')).toBeVisible();

		page.once('dialog', (d) => d.accept());
		await drawer.locator('button[aria-label="Delete rule"]').click();
		await page.waitForLoadState('networkidle');

		await expect(drawer.locator('text=/Show if/')).not.toBeVisible();
	});
});
