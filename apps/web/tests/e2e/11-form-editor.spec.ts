import { test, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const SHOTS_DIR = 'tests/e2e/screenshots/11-form-editor';

test.describe.configure({ mode: 'default' });

test.describe('Form editor v2 — 3-panel layout, drag-reorder, click-to-edit', () => {
	test.beforeAll(() => {
		fs.mkdirSync(SHOTS_DIR, { recursive: true });
	});

	let formId: number;

	test.beforeAll(async ({ request }) => {
		const login = await request.post('http://localhost:8080/v1/auth/login', {
			data: { email: 'admin@local.test', password: 'ChangeMe123!' }
		});
		const { SessionToken } = await login.json();
		const res = await request.post('http://localhost:8080/v1/forms', {
			headers: { Authorization: `Bearer ${SessionToken}` },
			data: { title: 'E2E editor test form', description: 'Created by 11-form-editor' }
		});
		const body = await res.json();
		formId = body.form?.id ?? body.id;
		expect(formId).toBeGreaterThan(0);
	});

	test('editor renders 3-panel layout with palette + empty canvas', async ({ page }) => {
		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');

		await expect(page.locator('aside').filter({ hasText: 'Add a field' })).toBeVisible();
		await expect(page.locator('text=/^Text$/').first()).toBeVisible();
		await expect(page.locator('text=/^Choice$/').first()).toBeVisible();
		await expect(page.locator('text=/^Assessment$/').first()).toBeVisible();
		await expect(page.locator('text=/No questions yet/')).toBeVisible();
		await page.screenshot({
			path: path.join(SHOTS_DIR, '01-empty-3-panel.png'),
			fullPage: true
		});
	});

	test('renderer-pending types show β badge in palette', async ({ page }) => {
		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');
		const palette = page.locator('aside').filter({ hasText: 'Add a field' });
		const hotspotRow = palette.locator('text=/^Hotspot$/').locator('..');
		await expect(hotspotRow.locator('text=/^β$/')).toBeVisible();
	});

	test('clicking a palette item adds a question to the canvas', async ({ page }) => {
		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');

		await page.locator('aside button:has-text("Short answer")').first().click();
		await page.waitForLoadState('networkidle');

		await expect(page.locator('text=/Short answer/').first()).toBeVisible();
		await expect(page.locator('text=/No questions yet/')).not.toBeVisible();
	});

	test('clicking a question opens the settings drawer; saving persists the title', async ({ page }) => {
		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');

		await page.locator('aside button:has-text("Multiple choice")').first().click();
		await page.waitForLoadState('networkidle');

		const row = page.locator('div').filter({ hasText: /^[0-9]+\.\s*Multiple choice/ }).first();
		await row.click();

		const drawer = page.locator('aside').filter({ hasText: 'Question settings' });
		await expect(drawer).toBeVisible();

		const titleInput = drawer.locator('input[name="title"]');
		await titleInput.fill('Quel est ton plat préféré ?');
		await drawer.getByRole('button', { name: /^Save$/ }).click();
		await page.waitForLoadState('networkidle');

		await expect(page.locator('text=/Quel est ton plat préféré/').first()).toBeVisible();
		await page.screenshot({
			path: path.join(SHOTS_DIR, '02-edited-question.png'),
			fullPage: true
		});
	});

	test('publish creates a new version and the canvas still loads', async ({ page }) => {
		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');

		const publishBtn = page.getByRole('button', { name: /Publish version/ });
		await publishBtn.click();
		await page.waitForLoadState('networkidle');

		await expect(page.locator('text=/Published as version/i')).toBeVisible();
		await page.screenshot({
			path: path.join(SHOTS_DIR, '03-published.png'),
			fullPage: true
		});
	});
});
