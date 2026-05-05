import { test, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const SHOTS_DIR = 'tests/e2e/screenshots/07-admin-assessment';

test.describe.configure({ mode: 'default' });

test.describe('Admin assessment surface — campaigns, enrollment, forms, periods', () => {
	test.beforeAll(() => {
		fs.mkdirSync(SHOTS_DIR, { recursive: true });
	});

	test('campaigns list — renders, scope chips visible, create modal opens', async ({ page }) => {
		await page.goto('/admin/campaigns');
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1')).toContainText(/Campaigns/i);

		const scopeLabel = page.locator('text=/^Scope:$/i');
		await expect(scopeLabel.first()).toBeVisible();

		const newButton = page.getByRole('button', { name: /\+ New campaign/ });
		await expect(newButton).toBeVisible();
		await newButton.click();
		await expect(page.locator('text=/^New campaign$/').first()).toBeVisible();
		await expect(page.locator('input[name="title"]')).toBeVisible();
		await page.screenshot({
			path: path.join(SHOTS_DIR, '01-campaigns-create-modal.png'),
			fullPage: true
		});

		await page.getByRole('button', { name: /Cancel/ }).click();
	});

	test('campaigns at scope=2 — E2E Personas Campaign visible', async ({ page }) => {
		await page.goto('/admin/campaigns?scope=2');
		await page.waitForLoadState('networkidle');
		await page.screenshot({
			path: path.join(SHOTS_DIR, '02-campaigns-scope-2.png'),
			fullPage: true
		});
		await expect(page.locator('text=/E2E Personas Campaign/i').first()).toBeVisible();
	});

	test('campaigns at scope=590 — Load Test Campaign visible', async ({ page }) => {
		await page.goto('/admin/campaigns?scope=590');
		await page.waitForLoadState('networkidle');
		await page.screenshot({
			path: path.join(SHOTS_DIR, '03-campaigns-scope-590.png'),
			fullPage: true
		});
		await expect(page.locator('text=/Load Test Campaign/i').first()).toBeVisible();
	});

	test('campaign detail page (campaign 1) — renders without error', async ({ page }) => {
		const errors: string[] = [];
		page.on('pageerror', (err) => errors.push(`pageerror: ${err.message}`));
		page.on('console', (msg) => {
			if (msg.type() === 'error') errors.push(`console.error: ${msg.text()}`);
		});

		const resp = await page.goto('/admin/campaigns/1');
		expect(resp?.status(), 'HTTP for /admin/campaigns/1').toBeLessThan(400);
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1').first()).toBeVisible();
		await expect(page.locator('text=/Public ID/i').first()).toBeVisible();
		await page.screenshot({
			path: path.join(SHOTS_DIR, '04-campaign-1-detail.png'),
			fullPage: true
		});
		expect(errors, 'Console/page errors on /admin/campaigns/1').toEqual([]);
	});

	test('campaign detail page (campaign 5 — Load Test) — renders without error', async ({
		page
	}) => {
		const errors: string[] = [];
		page.on('pageerror', (err) => errors.push(`pageerror: ${err.message}`));
		page.on('console', (msg) => {
			if (msg.type() === 'error') errors.push(`console.error: ${msg.text()}`);
		});

		const resp = await page.goto('/admin/campaigns/5');
		expect(resp?.status(), 'HTTP for /admin/campaigns/5').toBeLessThan(400);
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1').first()).toBeVisible();
		await page.screenshot({
			path: path.join(SHOTS_DIR, '05-campaign-5-detail.png'),
			fullPage: true
		});
		expect(errors, 'Console/page errors on /admin/campaigns/5').toEqual([]);
	});

	test('enrollment at scope=22304 — coverage tiles populate, table has rows', async ({ page }) => {
		await page.goto('/admin/enrollment?scope=22304');
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1')).toContainText(/Enrollment/i);

		await expect(page.locator('text=/^Total enrolled$/i').first()).toBeVisible();
		await expect(page.locator('text=/^Female$/i').first()).toBeVisible();
		await expect(page.locator('text=/^Male$/i').first()).toBeVisible();

		const totalCard = page
			.locator('div')
			.filter({ hasText: /^Total enrolled$/i })
			.first();
		await expect(totalCard).toBeVisible();

		const totalText = await page.locator('p.text-3xl.font-bold').first().textContent();
		const totalNum = Number((totalText ?? '0').replace(/\D/g, ''));
		expect(totalNum, 'Total enrolled count should be > 0 after 4.3M insert').toBeGreaterThan(0);

		const tableRows = page.locator('table tbody tr');
		await expect(tableRows.first()).toBeVisible();
		const rowCount = await tableRows.count();
		expect(rowCount, 'Enrollment rows should render').toBeGreaterThan(0);

		await page.screenshot({
			path: path.join(SHOTS_DIR, '06-enrollment-scope-22304.png'),
			fullPage: true
		});
	});

	test('enrollment period selector — switching period reloads page', async ({ page }) => {
		await page.goto('/admin/enrollment?scope=22304');
		await page.waitForLoadState('networkidle');

		await page.goto('/admin/enrollment?scope=22304&period=10');
		await page.waitForLoadState('networkidle');
		await expect(page.locator('text=/Period #10/i').first()).toBeVisible();
		await page.screenshot({
			path: path.join(SHOTS_DIR, '07-enrollment-period-10.png'),
			fullPage: true
		});
	});

	test('forms list — at least 6 forms; expected titles visible', async ({ page }) => {
		await page.goto('/admin/forms');
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1')).toContainText(/Forms/i);

		const rows = page.locator('table tbody tr');
		await expect(rows.first()).toBeVisible();
		const count = await rows.count();
		expect(count, 'Forms count').toBeGreaterThanOrEqual(6);

		await expect(page.locator('text=/E2E Personas Form/i').first()).toBeVisible();
		await expect(page.locator('text=/Pilot Form/i').first()).toBeVisible();
		await expect(page.locator('text=/Load Test Assessment/i').first()).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '08-forms-list.png'), fullPage: true });
	});

	test('clicking a published form opens detail (with version history if any)', async ({ page }) => {
		await page.goto('/admin/forms');
		await page.waitForLoadState('networkidle');

		const publishedRow = page
			.locator('table tbody tr')
			.filter({ has: page.locator('text=/^Published$/') })
			.first();
		await expect(publishedRow).toBeVisible();
		await publishedRow.locator('a').first().click();

		await page.waitForURL(/\/admin\/forms\/\d+/);
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1').first()).toBeVisible();
		await expect(page.locator('text=/Public ID:/i').first()).toBeVisible();
		await page.screenshot({
			path: path.join(SHOTS_DIR, '09-published-form-detail.png'),
			fullPage: true
		});
	});

	test('form detail page (form 1) — questions list renders', async ({ page }) => {
		const errors: string[] = [];
		page.on('pageerror', (err) => errors.push(`pageerror: ${err.message}`));
		page.on('console', (msg) => {
			if (msg.type() === 'error') errors.push(`console.error: ${msg.text()}`);
		});

		const resp = await page.goto('/admin/forms/1');
		expect(resp?.status(), 'HTTP for /admin/forms/1').toBeLessThan(400);
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1').first()).toBeVisible();
		await expect(page.locator('text=/Public ID:/i').first()).toBeVisible();

		const hasQuestions =
			(await page.locator('table tbody tr').count()) > 0 ||
			(await page.locator('text=/No questions yet/i').count()) > 0;
		expect(hasQuestions, 'Either questions table or empty state should render').toBe(true);

		await page.screenshot({ path: path.join(SHOTS_DIR, '10-form-1-detail.png'), fullPage: true });
		expect(errors, 'Console/page errors on /admin/forms/1').toEqual([]);
	});

	test('periods — 8 academic periods; "Pilot Year 562522" is current', async ({ page }) => {
		await page.goto('/admin/periods');
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1')).toContainText(/Academic periods/i);

		const rows = page.locator('table tbody tr');
		await expect(rows.first()).toBeVisible();
		const count = await rows.count();
		expect(count, 'Period row count').toBe(8);

		const pilotRow = rows.filter({ hasText: /Pilot Year 562522/i }).first();
		await expect(pilotRow).toBeVisible();
		await expect(pilotRow.locator('text=/^Current$/i')).toBeVisible();

		await page.screenshot({ path: path.join(SHOTS_DIR, '11-periods.png'), fullPage: true });
	});
});
