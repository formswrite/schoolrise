import { test, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const SHOTS_DIR = 'tests/e2e/screenshots/02-pilot-flow';

const stamp = Date.now().toString().slice(-6);
const PERIOD_CODE = `pilot-${stamp}`;
const NIVEAU_CODE = `CE1-${stamp}`;
const FORM_TITLE = `Pilot Form ${stamp}`;

test.describe.configure({ mode: 'serial' });

test.describe('Pilot flow — admin walks the §10 happy path through the UI', () => {
	test.beforeAll(() => {
		fs.mkdirSync(SHOTS_DIR, { recursive: true });
	});

	test('step 1 — create academic period and mark it current', async ({ page }) => {
		await page.goto('/admin/periods');
		await page.getByRole('button', { name: /\+ New period/ }).click();
		await page.locator('input[name="code"]').fill(PERIOD_CODE);
		await page.locator('input[name="label"]').fill(`Pilot Year ${stamp}`);
		await page.locator('input[name="starts_on"]').fill('2025-09-01');
		await page.locator('input[name="ends_on"]').fill('2026-06-30');
		await page.locator('label[for="is_current"]').click();
		await Promise.all([
			page.waitForURL('**/admin/periods**'),
			page.getByRole('button', { name: /^Create$/ }).click()
		]);
		await expect(page.locator('text=' + PERIOD_CODE).first()).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '01-periods-created.png'), fullPage: true });
	});

	test('step 2 — create niveau', async ({ page }) => {
		await page.goto('/admin/niveaux');
		await page.getByRole('button', { name: /\+ New niveau/ }).click();
		await page.locator('input[name="code"]').fill(NIVEAU_CODE);
		await page.locator('input[name="label"]').fill(`Cours Élémentaire 1 ${stamp}`);
		await page.locator('input[name="sort_order"]').fill('10');
		await page.getByRole('button', { name: /^Create$/ }).click();
		await page.waitForLoadState('networkidle');
		await expect(page.locator('text=' + NIVEAU_CODE).first()).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '02-niveau-created.png'), fullPage: true });
	});

	test('step 3 — create a form via the builder UI', async ({ page }) => {
		await page.goto('/admin/forms');
		await page.getByRole('button', { name: /\+ New form/ }).click();
		await page.locator('input[name="title"]').fill(FORM_TITLE);
		await page.locator('input[name="description"]').fill('End-to-end pilot');
		await page.getByRole('button', { name: /^Create form$/ }).click();
		await page.waitForURL(/\/admin\/forms\/\d+/);
		await page.screenshot({ path: path.join(SHOTS_DIR, '03-form-created.png'), fullPage: true });
	});

	test('step 4 — add 3 questions and publish', async ({ page }) => {
		await page.goto('/admin/forms');
		await page.locator(`text=${FORM_TITLE}`).first().click();
		await page.waitForURL(/\/admin\/forms\/\d+/);

		const addQuestion = async (type: string, title: string) => {
			const typeSelect = page.locator('select[name="type"]');
			if (!(await typeSelect.isVisible())) {
				await page.getByRole('button', { name: /\+ Add question/ }).click();
				await typeSelect.waitFor({ state: 'visible' });
			}
			await typeSelect.selectOption(type);
			await page.locator('input[name="title"]').fill(title);
			await page.getByRole('button', { name: /^Add$/ }).click();
			await page.waitForLoadState('networkidle');
		};

		await addQuestion('SHORT_ANSWER', 'What is your name?');
		await addQuestion('PARAGRAPH', 'Describe your favourite story.');
		await addQuestion('LINEAR_SCALE', 'How confident are you with reading?');

		await page.screenshot({
			path: path.join(SHOTS_DIR, '04-form-with-questions.png'),
			fullPage: true
		});

		await page.getByRole('button', { name: /Publish version/ }).click();
		await page.waitForLoadState('networkidle');
		await expect(page.locator('text=/Published as version/i')).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '05-form-published.png'), fullPage: true });
	});

	test('step 5 — visit dashboard for region 1', async ({ page }) => {
		await page.goto('/admin/dashboard?scope=1');
		await page.waitForLoadState('networkidle');
		await expect(page.locator('h1')).toContainText(/Progression dashboard/i);
		await expect(
			page.locator('text=/Pick a campaign|No campaigns at this scope/i').first()
		).toBeVisible();
		await page.screenshot({
			path: path.join(SHOTS_DIR, '06-dashboard-pick-campaign.png'),
			fullPage: true
		});
	});

	test('step 6 — drill into existing campaign 1 (when seeded)', async ({ page }) => {
		await page.goto('/admin/dashboard?scope=1&campaign=1&period=1');
		await page.waitForLoadState('networkidle');
		const bandHeader = page.locator('text=/Band distribution/i');
		if (await bandHeader.count()) {
			await expect(bandHeader.first()).toBeVisible();
		} else {
			await expect(page.locator('h1')).toContainText(/Progression dashboard/i);
		}
		await page.screenshot({
			path: path.join(SHOTS_DIR, '07-dashboard-with-bands.png'),
			fullPage: true
		});
	});

	test('step 7 — refresh snapshot (when seeded)', async ({ page }) => {
		await page.goto('/admin/dashboard?scope=1&campaign=1&period=1');
		await page.waitForLoadState('networkidle');
		const refresh = page.getByRole('button', { name: /Refresh snapshot/ });
		if (await refresh.count()) {
			await refresh.click();
			await page.waitForLoadState('networkidle');
		}
		await page.screenshot({
			path: path.join(SHOTS_DIR, '08-dashboard-after-refresh.png'),
			fullPage: true
		});
	});

	test('step 8 — drill into a child school (when seeded)', async ({ page }) => {
		await page.goto('/admin/dashboard?scope=1&campaign=1&period=1');
		await page.waitForLoadState('networkidle');
		const drillLink = page.locator('a:has-text("Drill in")').first();
		if (await drillLink.count()) {
			await drillLink.click();
			await page.waitForLoadState('networkidle');
			await page.screenshot({
				path: path.join(SHOTS_DIR, '09-dashboard-child-school.png'),
				fullPage: true
			});
		}
	});

	test('step 9 — open notifications outbox', async ({ page }) => {
		await page.goto('/admin/notifications');
		await expect(page.locator('text=/Provider/i').first()).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '10-notifications.png'), fullPage: true });
	});
});
