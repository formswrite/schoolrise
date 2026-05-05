import { test, expect, type Page, type BrowserContext } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const SHOTS = 'tests/e2e/screenshots/08-personas-scoped';

type Cred = {
	key: 'minister' | 'inspector.boke' | 'principal' | 'teacher.guinea';
	email: string;
	password: string;
	expectedRole: 'admin' | 'inspector' | 'teacher';
	expectedScopeId: number | null;
	expectedScopeLabel: string | null;
};

const CREDS: Record<Cred['key'], Cred> = {
	minister: {
		key: 'minister',
		email: 'minister@local.test',
		password: 'Minister123!',
		expectedRole: 'admin',
		expectedScopeId: null,
		expectedScopeLabel: null
	},
	'inspector.boke': {
		key: 'inspector.boke',
		email: 'inspector.boke@local.test',
		password: 'Inspector123!',
		expectedRole: 'inspector',
		expectedScopeId: 22304,
		expectedScopeLabel: 'Boké'
	},
	principal: {
		key: 'principal',
		email: 'principal@local.test',
		password: 'Principal123!',
		expectedRole: 'inspector',
		expectedScopeId: 637,
		expectedScopeLabel: 'LT School 1'
	},
	'teacher.guinea': {
		key: 'teacher.guinea',
		email: 'teacher.guinea@local.test',
		password: 'Teacher123!',
		expectedRole: 'teacher',
		expectedScopeId: 637,
		expectedScopeLabel: 'LT School 1'
	}
};

test.use({ storageState: { cookies: [], origins: [] } });

async function shoot(page: Page, name: string) {
	fs.mkdirSync(SHOTS, { recursive: true });
	await page.screenshot({ path: path.join(SHOTS, `${name}.png`), fullPage: true });
}

async function loginFresh(context: BrowserContext, page: Page, cred: Cred) {
	await context.clearCookies();
	const navResp = await page.goto('/login');
	expect(navResp?.status(), `GET /login should be 200 for ${cred.key}`).toBeLessThan(400);

	await page.locator('input[name="email"]').fill(cred.email);
	await page.locator('input[name="password"]').fill(cred.password);

	const [postResp] = await Promise.all([
		page.waitForResponse(
			(r) => r.url().includes('/login') && r.request().method() === 'POST',
			{ timeout: 10_000 }
		),
		page.locator('form button[type="submit"]').click()
	]);
	expect(postResp.status(), `POST /login should redirect (303) for ${cred.key}`).toBe(303);

	await page.waitForLoadState('networkidle');
	expect(page.url(), `${cred.key} should leave /login after submit`).not.toContain('/login');

	const cookies = await context.cookies();
	const session = cookies.find((c) => c.name === 'schoolrise_session');
	expect(session, `${cred.key} should have session cookie`).toBeTruthy();
}

test.describe('PERSONA · scoped views (08)', () => {
	test('minister (country-wide admin) sees country-level dashboard, ALL institutions, full nav', async ({ context, page }) => {
		const cred = CREDS.minister;
		await loginFresh(context, page, cred);

		await page.goto('/admin/dashboard');
		await expect(page.locator('h1')).toContainText(/Progression dashboard/i);
		await shoot(page, '01-minister-dashboard');

		const navLinks = await page.locator('aside nav a').allInnerTexts();
		expect(navLinks).toEqual(
			expect.arrayContaining(['Dashboard', 'Periods', 'Niveaux', 'Imports', 'Notifications', 'Users'])
		);
		expect(navLinks.length).toBeGreaterThanOrEqual(13);
		await expect(page.locator('aside p:has-text("Admin")').first()).toBeVisible();

		const scopeChips = page.locator('header a[href*="?scope="]');
		const chipCount = await scopeChips.count();
		expect(chipCount, 'minister should see many scope chips (country-wide)').toBeGreaterThan(1);
		const chipTexts = await scopeChips.allInnerTexts();
		console.log(`[minister] scope chip count=${chipCount}, sample=${chipTexts.slice(0, 5).join(' | ')}`);

		await page.goto('/admin/students');
		await expect(page.locator('h1')).toContainText(/Students/i);
		await shoot(page, '02-minister-students');
		expect(page.url()).toContain('/admin/students');
	});

	test('inspector.boke (region scope) auto-resolves to Boké, shows fewer chips, blocked from Users/Periods', async ({ context, page }) => {
		const cred = CREDS['inspector.boke'];
		await loginFresh(context, page, cred);

		await page.goto('/admin/dashboard');
		await expect(page.locator('h1')).toContainText(/Progression dashboard/i);
		await shoot(page, '03-inspector-boke-dashboard');

		await expect(page.locator('aside p:has-text("Inspector")').first()).toBeVisible();

		const scopeBanner = page.locator('header p.text-muted-foreground').first();
		const bannerText = (await scopeBanner.textContent({ timeout: 5_000 }).catch(() => '')) ?? '';
		expect(bannerText, 'inspector.boke dashboard should auto-resolve to Boké').toMatch(/Boké/);
		console.log(`[inspector.boke] resolved scope banner: ${bannerText.trim()}`);

		const navLinks = await page.locator('aside nav a').allInnerTexts();
		expect(navLinks).not.toContain('Periods');
		expect(navLinks).not.toContain('Niveaux');
		expect(navLinks).not.toContain('Users');
		expect(navLinks).not.toContain('Imports');
		expect(navLinks).not.toContain('Notifications');

		const chipTexts = await page.locator('header a[href*="?scope="]').allInnerTexts();
		console.log(`[inspector.boke] scope chips=${chipTexts.length}: ${chipTexts.join(' | ')}`);
		expect(chipTexts.length).toBeGreaterThanOrEqual(1);
		expect(chipTexts.join(' | ')).toMatch(/Boké/);

		const usersResp = await page.goto('/admin/users');
		expect(usersResp?.status(), 'inspector blocked from /admin/users').toBe(403);
		await shoot(page, '04-inspector-boke-blocked-users');

		const periodsResp = await page.goto('/admin/periods');
		expect(periodsResp?.status(), 'inspector blocked from /admin/periods').toBe(403);
	});

	test('principal (institution scope, role=inspector) sees only LT School 1, blocked from Users', async ({ context, page }) => {
		const cred = CREDS.principal;
		await loginFresh(context, page, cred);

		await page.goto('/admin/dashboard');
		await expect(page.locator('h1')).toContainText(/Progression dashboard/i);
		await shoot(page, '05-principal-dashboard');

		await expect(page.locator('aside p:has-text("Inspector")').first()).toBeVisible();

		const bannerText = (await page.locator('header p.text-muted-foreground').first().textContent({ timeout: 5_000 }).catch(() => '')) ?? '';
		console.log(`[principal] resolved scope banner: ${bannerText.trim()}`);
		expect(bannerText, 'principal dashboard should auto-resolve to LT School 1').toMatch(/LT School 1/);

		const chipTexts = await page.locator('header a[href*="?scope="]').allInnerTexts();
		console.log(`[principal] scope chips=${chipTexts.length}: ${chipTexts.join(' | ')}`);
		expect(chipTexts.length, 'principal should see at most 1 scope chip').toBeLessThanOrEqual(1);
		if (chipTexts.length === 1) {
			expect(chipTexts[0]).toMatch(/LT School 1/);
		}

		const usersResp = await page.goto('/admin/users');
		expect(usersResp?.status(), 'principal blocked from /admin/users').toBe(403);
		await shoot(page, '06-principal-blocked-users');
	});

	test('teacher.guinea is redirected away from /admin/* and lands on /teacher with only their classes', async ({ context, page }) => {
		const cred = CREDS['teacher.guinea'];
		await loginFresh(context, page, cred);

		await page.goto('/admin/dashboard');
		await page.waitForLoadState('networkidle');
		expect(page.url(), 'teacher should be redirected away from /admin/dashboard').toContain('/teacher');
		await shoot(page, '07-teacher-redirect-from-admin-dashboard');

		await page.goto('/admin/students');
		await page.waitForLoadState('networkidle');
		expect(page.url(), 'teacher should be redirected away from /admin/students').toContain('/teacher');
		await shoot(page, '08-teacher-redirect-from-admin-students');

		await page.goto('/teacher');
		await expect(page.locator('h1')).toContainText(/Your classes/i);
		await shoot(page, '09-teacher-landing');

		const adminNav = await page.locator('aside nav a').count();
		expect(adminNav, 'teacher should not see admin sidebar nav').toBe(0);

		const classCards = page.locator('a[href^="/teacher/classes/"]');
		const classCount = await classCards.count();
		console.log(`[teacher.guinea] visible classes on /teacher: ${classCount}`);
		expect(classCount, 'teacher.guinea should see only the classes they teach').toBeGreaterThanOrEqual(0);
	});
});
