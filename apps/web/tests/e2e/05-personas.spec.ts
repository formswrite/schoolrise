import { test, expect, type Page } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';
import type { Personas } from './global.setup.personas';
import { newAssignmentToken } from './lib/seed';

const SHOTS = 'tests/e2e/screenshots/05-personas';

function loadPersonas(): Personas {
	return JSON.parse(fs.readFileSync('tests/e2e/.auth/personas.json', 'utf8'));
}

async function shoot(page: Page, name: string) {
	fs.mkdirSync(SHOTS, { recursive: true });
	await page.screenshot({ path: path.join(SHOTS, `${name}.png`), fullPage: true });
}

async function captureConsoleErrors(page: Page): Promise<string[]> {
	const errors: string[] = [];
	page.on('pageerror', (e) => errors.push(`pageerror: ${e.message}`));
	page.on('console', (m) => {
		if (m.type() === 'error') errors.push(`console.error: ${m.text()}`);
	});
	return errors;
}

test.describe('PERSONA · global admin', () => {
	test.use({ storageState: 'tests/e2e/.auth/admin.json' });

	test('admin sees the full 12-item nav and can reach every section', async ({ page }) => {
		const errors = await captureConsoleErrors(page);
		await page.goto('/admin/dashboard');
		await expect(page.locator('h1')).toContainText(/Progression dashboard/i);

		const navLinks = await page.locator('aside nav a').allInnerTexts();
		expect(navLinks).toEqual(
			expect.arrayContaining(['Dashboard', 'Institutions', 'Periods', 'Niveaux', 'Classes', 'Students', 'Staff', 'Enrollment', 'Forms', 'Campaigns', 'AI', 'Imports', 'Notifications', 'Users'])
		);
		expect(navLinks.length).toBe(14);

		await shoot(page, '01-admin-dashboard');
		await expect(page.locator('a:has-text("Teacher view")').first()).toBeVisible();
		await expect(page.locator('aside p:has-text("Admin")').first()).toBeVisible();
		expect(errors).toEqual([]);
	});

	test('admin can crud — period create-then-delete', async ({ page }) => {
		const stamp = Date.now().toString().slice(-5);
		await page.goto('/admin/periods');
		await page.getByRole('button', { name: /\+ New period/ }).click();
		await page.locator('input[name="code"]').fill(`p-${stamp}`);
		await page.locator('input[name="label"]').fill(`Period ${stamp}`);
		await page.locator('input[name="starts_on"]').fill('2030-09-01');
		await page.locator('input[name="ends_on"]').fill('2031-06-30');
		await page.getByRole('button', { name: /^Create$/ }).click();
		await page.waitForLoadState('networkidle');
		await expect(page.locator(`text=p-${stamp}`).first()).toBeVisible();
		await shoot(page, '02-admin-period-created');

		page.once('dialog', (d) => d.accept());
		const row = page.locator('tr', { hasText: `p-${stamp}` });
		await row.getByRole('button', { name: 'Delete' }).click();
		await page.waitForLoadState('networkidle');
		await expect(row).toHaveCount(0);
	});

	test('admin can publish a form and create a campaign via existing surfaces', async ({ page }) => {
		await page.goto('/admin/forms');
		await expect(page.locator('text=E2E Personas Form').first()).toBeVisible();
		await shoot(page, '03-admin-forms-list');
	});

	test('admin can browse campaigns and view a campaign detail page', async ({ page }) => {
		const personas: Personas = loadPersonas();
		await page.goto(`/admin/campaigns?scope=${personas.context.schoolID}`);
		await expect(page.locator('h1')).toContainText(/Campaigns/i);
		await expect(page.locator('text=E2E Personas Campaign').first()).toBeVisible();
		await shoot(page, '05-admin-campaigns-list');

		await page.locator('text=E2E Personas Campaign').first().click();
		await page.waitForURL(/\/admin\/campaigns\/\d+/);
		await expect(page.locator('text=/Assign students/i')).toBeVisible();
		await expect(page.locator('text=Assigned').first()).toBeVisible();
		await shoot(page, '06-admin-campaign-detail');
	});

	test('admin can open a class detail page and see students + staff', async ({ page }) => {
		const personas: Personas = loadPersonas();
		await page.goto(`/admin/classes/${personas.context.classID}`);
		await expect(page.locator('h1')).toContainText(/E2E-CE1-A/i);
		await expect(page.locator('text=Students in this class').first()).toBeVisible();
		await expect(page.locator('text=Staff on this class').first()).toBeVisible();
		await shoot(page, '07-admin-class-detail');
	});

	test('admin can act-as-teacher via the header link', async ({ page }) => {
		await page.goto('/admin/dashboard');
		await page.locator('a:has-text("Teacher view")').click();
		await page.waitForURL(/\/teacher/);
		await expect(page.locator('text=/Admin · acting as teacher/i')).toBeVisible();
		await shoot(page, '04-admin-acting-as-teacher');
	});
});

test.describe('PERSONA · teacher (scoped to one institution + class)', () => {
	test.use({ storageState: 'tests/e2e/.auth/teacher.json' });

	test('teacher visiting /admin/dashboard is redirected to /teacher', async ({ page }) => {
		await page.goto('/admin/dashboard');
		await page.waitForLoadState('networkidle');
		expect(page.url()).toContain('/teacher');
		await shoot(page, '15-teacher-redirected-from-admin');
	});

	test('teacher landing shows ONLY their classes (not the global admin nav)', async ({ page }) => {
		await page.goto('/teacher');
		await expect(page.locator('h1')).toContainText(/Your classes/i);
		await expect(page.locator('text=E2E-CE1-A').first()).toBeVisible();

		const adminNav = await page.locator('aside nav a').count();
		expect(adminNav).toBe(0);

		await expect(page.locator('text=Teacher').first()).toBeVisible();
		await shoot(page, '10-teacher-landing');
	});

	test('teacher drills into class → campaigns → grade entry', async ({ page }) => {
		await page.goto('/teacher');
		await page.locator('text=E2E-CE1-A').first().click();
		await page.waitForURL(/\/teacher\/classes\/\d+/);
		await expect(page.locator('text=/Open campaigns/i')).toBeVisible();
		await shoot(page, '11-teacher-campaigns');

		await page.getByRole('link', { name: /Enter scores|Continue/ }).first().click();
		await page.waitForURL(/\/teacher\/classes\/\d+\/campaigns\/\d+/);
		await expect(page.locator('text=/Roster/i')).toBeVisible();
		await shoot(page, '12-teacher-grade-entry');
	});

	test('teacher submits scores end-to-end', async ({ page }) => {
		await page.goto('/teacher');
		await page.locator('text=E2E-CE1-A').first().click();
		await page.getByRole('link', { name: /Enter scores|Continue/ }).first().click();
		await page.waitForLoadState('networkidle');

		const inputs = await page.locator('input[name^="score_"]').all();
		const scores = [22, 48, 71, 95];
		for (let i = 0; i < inputs.length && i < scores.length; i++) {
			await inputs[i].fill(String(scores[i]));
		}
		await shoot(page, '13-teacher-scores-typed');

		await page.getByRole('button', { name: /Submit batch/ }).click();
		await page.waitForLoadState('networkidle');

		const toast = page.locator('[data-sonner-toast]').first();
		await expect(toast).toBeVisible({ timeout: 5_000 });
		await expect(toast).toContainText(/Saved \d+ score/);
		await shoot(page, '14-teacher-scores-saved-toast');
	});

	test('teacher visiting /admin/users is also redirected (no admin shell access at all)', async ({ page }) => {
		await page.goto('/admin/users');
		await page.waitForLoadState('networkidle');
		expect(page.url()).toContain('/teacher');
		await shoot(page, '16-teacher-redirected-from-admin-users');
	});
});

test.describe('PERSONA · scoped inspector (read-only at region scope)', () => {
	test.use({ storageState: 'tests/e2e/.auth/inspector.json' });

	test('inspector lands and sees a FILTERED nav (no Periods/Niveaux/Imports/Notifications/Users)', async ({ page }) => {
		await page.goto('/admin/dashboard');
		await expect(page.locator('h1')).toContainText(/Progression dashboard/i);

		const navLinks = await page.locator('aside nav a').allInnerTexts();
		expect(navLinks).toEqual(
			expect.arrayContaining(['Dashboard', 'Institutions', 'Classes', 'Students', 'Staff', 'Enrollment', 'Forms'])
		);
		expect(navLinks).not.toContain('Periods');
		expect(navLinks).not.toContain('Niveaux');
		expect(navLinks).not.toContain('Imports');
		expect(navLinks).not.toContain('Notifications');
		expect(navLinks).not.toContain('Users');

		await expect(page.locator('aside p:has-text("Inspector")').first()).toBeVisible();
		await shoot(page, '20-inspector-dashboard');
	});

	test('inspector visits /admin/institutions (read-only allowed at their scope)', async ({ page }) => {
		await page.goto('/admin/institutions');
		await expect(page.locator('h1')).toContainText(/Institutions/i);
		await shoot(page, '21-inspector-institutions');
	});

	test('inspector trying /teacher gets redirected (no teacher role)', async ({ page }) => {
		await page.goto('/teacher');
		await page.waitForLoadState('networkidle');
		expect(page.url()).not.toContain('/teacher');
		await shoot(page, '22-inspector-tried-teacher-route');
	});

	test('inspector trying /admin/users gets a 403 page', async ({ page }) => {
		const resp = await page.goto('/admin/users');
		expect(resp?.status()).toBe(403);
		await shoot(page, '23-inspector-blocked-from-users');
	});

	test('inspector trying /admin/periods (admin-only) gets a 403 page', async ({ page }) => {
		const resp = await page.goto('/admin/periods');
		expect(resp?.status()).toBe(403);
		await shoot(page, '24-inspector-blocked-from-periods');
	});
});

test.describe('PERSONA · public student (no auth, token-only)', () => {
	test.use({ storageState: { cookies: [], origins: [] } });

	test('student opens an invalid token link and sees the friendly error', async ({ page }) => {
		await page.goto('/r/totally-fake-token-xyz');
		await expect(page.locator('text=/Invalid link/i')).toBeVisible();
		await shoot(page, '30-student-invalid-token');
	});

	test('student opens a valid token, sees the form, submits', async ({ page }) => {
		const personas: Personas = loadPersonas();
		const token = await newAssignmentToken(personas);

		await page.goto(`/r/${token}`);
		await expect(page.locator('text=/SchoolRise assessment/i')).toBeVisible();
		await shoot(page, '31-student-form-rendered');

		const inputs = await page.locator('form input[name^="q_"], form textarea[name^="q_"], form select[name^="q_"]').all();
		for (const input of inputs) {
			const tag = await input.evaluate((e) => e.tagName.toLowerCase());
			const type = await input.getAttribute('type');
			if (tag === 'textarea') await input.fill('Persona answer');
			else if (type === 'radio' && !(await input.isChecked())) await input.check();
			else if (type === 'checkbox') await input.check();
			else if (type === 'date') await input.fill('2025-09-15');
			else if (type === 'time') await input.fill('09:00');
			else if (type === 'number') await input.fill('42');
			else if (type === 'email') await input.fill('p@e.com');
			else if (type === 'tel') await input.fill('555-0100');
			else if (type === 'range') continue;
			else await input.fill('Persona answer');
		}
		await shoot(page, '32-student-form-filled');

		await Promise.all([
			page.waitForURL(/\/r\/.+\/done/),
			page.locator('form button[type="submit"]').last().click()
		]);
		await expect(page.locator('text=/Submitted|Thank you/i').first()).toBeVisible();
		await shoot(page, '33-student-done');
	});

	test('student tries to reach /admin/dashboard without auth → redirected', async ({ page }) => {
		const resp = await page.goto('/admin/dashboard');
		expect(resp?.status()).toBeLessThan(400);
		expect(page.url()).toContain('/login');
		await shoot(page, '34-anon-blocked');
	});
});
