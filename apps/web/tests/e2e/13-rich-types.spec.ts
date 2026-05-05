import { test, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const SHOTS_DIR = 'tests/e2e/screenshots/13-rich-types';
const API_BASE = 'http://localhost:8080';

test.describe.configure({ mode: 'default' });

test.describe('Rich field types — editor settings + previews persist', () => {
	test.beforeAll(() => {
		fs.mkdirSync(SHOTS_DIR, { recursive: true });
	});

	let formId: number;
	let token: string;

	test.beforeAll(async ({ request }) => {
		const login = await request.post(`${API_BASE}/v1/auth/login`, {
			data: { email: 'admin@local.test', password: 'ChangeMe123!' }
		});
		token = (await login.json()).SessionToken;

		const formRes = await request.post(`${API_BASE}/v1/forms`, {
			headers: { Authorization: `Bearer ${token}` },
			data: { title: 'E2E rich-types test', description: 'Phase 3' }
		});
		formId = (await formRes.json()).form?.id ?? (await formRes.json()).id;
		expect(formId).toBeGreaterThan(0);
	});

	test('palette shows β badge only for HOTSPOT (the last pending type)', async ({ page }) => {
		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');
		const palette = page.locator('aside').filter({ hasText: 'Add a field' });
		const badges = await palette.locator('text=/^β$/').count();
		expect(badges).toBe(1);
		await page.screenshot({
			path: path.join(SHOTS_DIR, '01-palette-after-phase-3.png'),
			fullPage: true
		});
	});

	test('EQUATION: stores LaTeX in extra; grading.correct_value persists', async ({ request }) => {
		const addRes = await request.post(`${API_BASE}/v1/forms/items/${formId}/questions`, {
			headers: { Authorization: `Bearer ${token}` },
			data: {
				client_id: 'eq_' + Math.random().toString(36).slice(2, 10),
				sort_order: 10,
				title: "Résoudre l'équation",
				type: 'EQUATION',
				required: true,
				extra: { latex: 'x^2 + 2x + 1 = 0' },
				grading: { points_max: 2, correct_value: '-1' }
			}
		});
		expect(addRes.ok()).toBeTruthy();

		const getRes = await request.get(`${API_BASE}/v1/forms/items/${formId}`, {
			headers: { Authorization: `Bearer ${token}` }
		});
		const body = await getRes.json();
		const eq = body.questions.find((q: { type: string }) => q.type === 'EQUATION');
		expect(eq.extra.latex).toBe('x^2 + 2x + 1 = 0');
		expect(eq.grading.correct_value).toBe('-1');
		expect(eq.grading.points_max).toBe(2);
	});

	test('FILL_IN_BLANK: template + answers persist', async ({ request }) => {
		const addRes = await request.post(`${API_BASE}/v1/forms/items/${formId}/questions`, {
			headers: { Authorization: `Bearer ${token}` },
			data: {
				client_id: 'fib_' + Math.random().toString(36).slice(2, 10),
				sort_order: 20,
				title: 'Compléter les phrases',
				type: 'FILL_IN_BLANK',
				required: true,
				extra: { template: 'Le mot manquant est [[1]] et [[2]] aussi.' },
				grading: { answers: ['premier', 'second'], points_max: 2 }
			}
		});
		expect(addRes.ok()).toBeTruthy();

		const getRes = await request.get(`${API_BASE}/v1/forms/items/${formId}`, {
			headers: { Authorization: `Bearer ${token}` }
		});
		const fib = (await getRes.json()).questions.find(
			(q: { type: string }) => q.type === 'FILL_IN_BLANK'
		);
		expect(fib.extra.template).toContain('[[1]]');
		expect(fib.grading.answers).toEqual(['premier', 'second']);
	});

	test('TABLE: rows + columns persist', async ({ request }) => {
		const addRes = await request.post(`${API_BASE}/v1/forms/items/${formId}/questions`, {
			headers: { Authorization: `Bearer ${token}` },
			data: {
				client_id: 'tbl_' + Math.random().toString(36).slice(2, 10),
				sort_order: 30,
				title: 'Présence par mois',
				type: 'TABLE',
				required: true,
				extra: { rows: ['CP1', 'CP2', 'CE1'], columns: ['Sept', 'Oct', 'Nov'] }
			}
		});
		expect(addRes.ok()).toBeTruthy();

		const tbl = (
			await (
				await request.get(`${API_BASE}/v1/forms/items/${formId}`, {
					headers: { Authorization: `Bearer ${token}` }
				})
			).json()
		).questions.find((q: { type: string }) => q.type === 'TABLE');
		expect(tbl.extra.rows).toEqual(['CP1', 'CP2', 'CE1']);
		expect(tbl.extra.columns).toEqual(['Sept', 'Oct', 'Nov']);
	});

	test('Validation: text min/max_length round-trips', async ({ request }) => {
		const addRes = await request.post(`${API_BASE}/v1/forms/items/${formId}/questions`, {
			headers: { Authorization: `Bearer ${token}` },
			data: {
				client_id: 'val_' + Math.random().toString(36).slice(2, 10),
				sort_order: 40,
				title: 'Bounded text',
				type: 'SHORT_ANSWER',
				required: false,
				validation: { min_length: 3, max_length: 50 }
			}
		});
		expect(addRes.ok()).toBeTruthy();

		const q = (
			await (
				await request.get(`${API_BASE}/v1/forms/items/${formId}`, {
					headers: { Authorization: `Bearer ${token}` }
				})
			).json()
		).questions.find((qi: { title: string }) => qi.title === 'Bounded text');
		expect(q.validation.min_length).toBe(3);
		expect(q.validation.max_length).toBe(50);
	});

	test('Settings drawer renders type-specific extra fields', async ({ page }) => {
		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');

		const eqRow = page
			.locator('div')
			.filter({ hasText: /^[0-9]+\.\s*Résoudre/ })
			.first();
		await eqRow.click();
		const drawer = page.locator('aside').filter({ hasText: 'Question settings' });
		await expect(drawer).toBeVisible();
		await expect(drawer.locator('input[name="extra_latex"]')).toBeVisible();
		await expect(drawer.locator('input[name="extra_latex"]')).toHaveValue('x^2 + 2x + 1 = 0');
		await expect(drawer.locator('summary:has-text("Grading")')).toBeVisible();
	});
});
