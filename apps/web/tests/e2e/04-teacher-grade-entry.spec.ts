import { test, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const SHOTS_DIR = 'tests/e2e/screenshots/04-teacher-grade-entry';

test.describe.configure({ mode: 'serial' });
test.use({ storageState: 'tests/e2e/.auth/teacher.json' });

test.describe('Teacher proctored grade entry', () => {
	test.beforeAll(() => {
		fs.mkdirSync(SHOTS_DIR, { recursive: true });
	});

	test('step 1 — teacher lands on /teacher and sees their classes', async ({ page }) => {
		await page.goto('/teacher');
		await expect(page.locator('h1')).toContainText(/Your classes/i);
		await expect(page.locator('text=CE1-A')).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '01-class-picker.png'), fullPage: true });
	});

	test('step 2 — picks a class and sees campaigns', async ({ page }) => {
		await page.goto('/teacher');
		await page.locator('text=CE1-A').first().click();
		await page.waitForURL(/\/teacher\/classes\/\d+/);
		await expect(page.locator('text=/Open campaigns/i')).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '02-campaign-picker.png'), fullPage: true });
	});

	test('step 3 — opens grade entry and sees roster', async ({ page }) => {
		await page.goto('/teacher');
		await page.locator('text=CE1-A').first().click();
		await page.waitForURL(/\/teacher\/classes\/\d+/);

		const enterBtn = page.getByRole('link', { name: /Enter scores|Continue/ }).first();
		await enterBtn.click();
		await page.waitForURL(/\/teacher\/classes\/\d+\/campaigns\/\d+/);
		await expect(page.locator('text=/Roster/i')).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '03-roster-empty.png'), fullPage: true });
	});

	test('step 4 — enters scores for all students and submits', async ({ page }) => {
		await page.goto('/teacher');
		await page.locator('text=CE1-A').first().click();
		await page
			.getByRole('link', { name: /Enter scores|Continue/ })
			.first()
			.click();
		await page.waitForURL(/\/teacher\/classes\/\d+\/campaigns\/\d+/);

		const inputs = await page.locator('input[name^="score_"]').all();
		const targetScores = [25, 55, 88];
		for (let i = 0; i < inputs.length && i < targetScores.length; i++) {
			await inputs[i].fill(String(targetScores[i]));
		}
		await page.screenshot({ path: path.join(SHOTS_DIR, '04-roster-filled.png'), fullPage: true });

		await page.getByRole('button', { name: /Submit batch/ }).click();
		await page.waitForLoadState('networkidle');
		await expect(page.locator('text=/Saved/i').first()).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '05-after-submit.png'), fullPage: true });
	});

	test('step 5 — reload shows persisted scores with Saved chips and bands', async ({ page }) => {
		await page.goto('/teacher');
		await page.locator('text=CE1-A').first().click();
		await page
			.getByRole('link', { name: /Continue|Enter scores/ })
			.first()
			.click();
		await page.waitForLoadState('networkidle');

		const savedChips = await page.locator('text=Saved').count();
		expect(savedChips).toBeGreaterThan(0);
		await page.screenshot({
			path: path.join(SHOTS_DIR, '06-reload-persisted.png'),
			fullPage: true
		});
	});
});
