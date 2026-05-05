import { test, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const SHOTS_DIR = 'tests/e2e/screenshots/12-form-logic';
const API_BASE = 'http://localhost:8080';

test.describe.configure({ mode: 'default' });

test.describe('Form logic engine — show/hide rules persist + render', () => {
	test.beforeAll(() => {
		fs.mkdirSync(SHOTS_DIR, { recursive: true });
	});

	let formId: number;
	let q1Id: number; // Multiple choice (source)
	let q2Id: number; // Short answer (target)
	let q1ClientId: string;
	let q2ClientId: string;
	let token: string;

	test.beforeAll(async ({ request }) => {
		const login = await request.post(`${API_BASE}/v1/auth/login`, {
			data: { email: 'admin@local.test', password: 'ChangeMe123!' }
		});
		token = (await login.json()).SessionToken;

		const formRes = await request.post(`${API_BASE}/v1/forms`, {
			headers: { Authorization: `Bearer ${token}` },
			data: { title: 'E2E logic test form', description: 'Created by 12-form-logic' }
		});
		formId = (await formRes.json()).form?.id ?? (await formRes.json()).id;
		expect(formId).toBeGreaterThan(0);

		q1ClientId = 'src_q1_' + Math.random().toString(36).slice(2, 10);
		q2ClientId = 'tgt_q2_' + Math.random().toString(36).slice(2, 10);

		const q1Res = await request.post(`${API_BASE}/v1/forms/items/${formId}/questions`, {
			headers: { Authorization: `Bearer ${token}` },
			data: {
				client_id: q1ClientId,
				sort_order: 10,
				title: 'Choisis une option',
				type: 'MULTIPLE_CHOICE',
				required: true,
				options: [
					{ label: 'Option A', value: 'A' },
					{ label: 'Option B', value: 'B' }
				]
			}
		});
		q1Id = (await q1Res.json()).id;

		const q2Res = await request.post(`${API_BASE}/v1/forms/items/${formId}/questions`, {
			headers: { Authorization: `Bearer ${token}` },
			data: {
				client_id: q2ClientId,
				sort_order: 20,
				title: 'Question révélée',
				type: 'SHORT_ANSWER',
				required: false
			}
		});
		q2Id = (await q2Res.json()).id;
	});

	test('logic panel persists a show_if rule into forms.settings.logic_rules', async ({
		page,
		request
	}) => {
		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');

		const targetRow = page
			.locator('div')
			.filter({ hasText: /^[0-9]+\.\s*Question révélée/ })
			.first();
		await targetRow.click();

		const drawer = page.locator('aside').filter({ hasText: 'Question settings' });
		await expect(drawer).toBeVisible();

		await drawer.locator('button:has-text("+ Add rule")').click();
		await drawer.locator('select[name="rule_operator"]').selectOption('show_if');
		await drawer.locator('select[name="source_client_id"]').selectOption(q1ClientId);
		await drawer.locator('select[name="cond_op"]').selectOption('equals');
		await drawer.locator('select[name="cond_value"]').selectOption('A');
		await drawer.getByRole('button', { name: /^Save rule$/ }).click();
		await page.waitForLoadState('networkidle');

		await expect(drawer.locator('text=/Show if/')).toBeVisible();

		const formRes = await request.get(`${API_BASE}/v1/forms/items/${formId}`, {
			headers: { Authorization: `Bearer ${token}` }
		});
		const settings = (await formRes.json()).form.settings;
		expect(settings.logic_rules).toBeDefined();
		expect(settings.logic_rules.length).toBe(1);
		expect(settings.logic_rules[0].target_question_client_id).toBe(q2ClientId);
		expect(settings.logic_rules[0].conditions[0].source_question_client_id).toBe(q1ClientId);
		expect(settings.logic_rules[0].conditions[0].value).toBe('A');

		await page.screenshot({ path: path.join(SHOTS_DIR, '01-rule-saved.png'), fullPage: true });
	});

	test('publishing snapshots logic_rules so the public renderer can evaluate offline', async ({
		request
	}) => {
		const pubRes = await request.post(`${API_BASE}/v1/forms/items/${formId}/publish`, {
			headers: { Authorization: `Bearer ${token}` }
		});
		expect(pubRes.ok()).toBeTruthy();
		const pubBody = await pubRes.json();
		const versionId = pubBody.version?.id;
		expect(versionId).toBeGreaterThan(0);

		const verRes = await request.get(
			`${API_BASE}/v1/forms/public-versions/${versionId}`
		);
		expect(verRes.ok()).toBeTruthy();
		const v = await verRes.json();
		expect(v.snapshot.settings.logic_rules).toBeDefined();
		expect(v.snapshot.settings.logic_rules.length).toBe(1);
		expect(v.snapshot.settings.logic_rules[0].target_question_client_id).toBe(q2ClientId);
	});

	test('pure logic evaluator: hidden when condition not met, visible when met', async () => {
		const { computeVisibleQuestions } = await import('../../src/lib/forms/logic.ts');
		const questions = [
			{ client_id: q1ClientId, type: 'MULTIPLE_CHOICE', title: 'Source', sort_order: 10, required: true },
			{ client_id: q2ClientId, type: 'SHORT_ANSWER', title: 'Target', sort_order: 20, required: false }
		] as never;
		const rules = [
			{
				id: 'r_1',
				target_question_client_id: q2ClientId,
				operator: 'show_if' as const,
				conditions: [
					{ source_question_client_id: q1ClientId, op: 'equals' as const, value: 'A' }
				]
			}
		];

		const hidden = computeVisibleQuestions(questions, rules, { [q1ClientId]: 'B' });
		expect(hidden.map((q: { client_id: string }) => q.client_id)).toEqual([q1ClientId]);

		const visible = computeVisibleQuestions(questions, rules, { [q1ClientId]: 'A' });
		expect(visible.map((q: { client_id: string }) => q.client_id)).toEqual([q1ClientId, q2ClientId]);
	});
});
