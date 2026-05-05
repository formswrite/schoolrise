import { test, expect, request } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const SHOTS_DIR = 'tests/e2e/screenshots/03-public-renderer';
const GATEWAY_URL = process.env.E2E_GATEWAY_URL ?? 'http://localhost:8080';
const ADMIN_EMAIL = process.env.E2E_ADMIN_EMAIL ?? 'admin@local.test';
const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD ?? 'ChangeMe123!';

async function loginToken(): Promise<string> {
	const ctx = await request.newContext();
	const res = await ctx.post(`${GATEWAY_URL}/v1/auth/login`, {
		data: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD }
	});
	const body = await res.json();
	await ctx.dispose();
	return body.SessionToken as string;
}

async function newAssignmentToken(): Promise<{ token: string; campaignTitle: string }> {
	const sessionToken = await loginToken();
	const ctx = await request.newContext({
		extraHTTPHeaders: { Authorization: `Bearer ${sessionToken}` }
	});

	const stamp = Date.now().toString().slice(-6);
	const studentRes = await ctx.post(`${GATEWAY_URL}/v1/people/students`, {
		data: { institutionId: 2, fullName: `E2E Student ${stamp}` }
	});
	const student = (await studentRes.json()).student;

	const assignRes = await ctx.post(`${GATEWAY_URL}/v1/campaigns/1/assign`, {
		data: { student_ids: [student.id], notify_by_email: false }
	});
	const assign = await assignRes.json();
	const access = assign.created?.[0]?.access_token as string | undefined;

	const campRes = await ctx.get(`${GATEWAY_URL}/v1/campaigns/1`);
	const camp = await campRes.json();

	await ctx.dispose();
	if (!access) throw new Error('Could not get an access token from /v1/campaigns/1/assign');
	return { token: access, campaignTitle: camp.title };
}

test.describe('Public student renderer', () => {
	test.use({ storageState: { cookies: [], origins: [] } });

	test.beforeAll(() => {
		fs.mkdirSync(SHOTS_DIR, { recursive: true });
	});

	test('invalid token shows graceful error', async ({ page }) => {
		await page.goto('/r/totally-bogus-token');
		await expect(page.locator('text=/Invalid link/i')).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '01-invalid-token.png'), fullPage: true });
	});

	test('valid token renders the form, student submits, dashboard reflects it', async ({ page }) => {
		const { token } = await newAssignmentToken();

		await page.goto(`/r/${token}`);
		await page.waitForLoadState('networkidle');
		await expect(page.locator('text=/SchoolRise assessment/i')).toBeVisible();
		await expect(page.locator('form button[type="submit"]').last()).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '02-form-rendered.png'), fullPage: true });

		const inputs = await page
			.locator('form input[name^="q_"], form textarea[name^="q_"], form select[name^="q_"]')
			.all();
		for (const input of inputs) {
			const tag = await input.evaluate((el) => el.tagName.toLowerCase());
			const type = await input.getAttribute('type');
			if (tag === 'textarea') {
				await input.fill('Pilot answer');
			} else if (type === 'radio') {
				if (!(await input.isChecked())) await input.check();
			} else if (type === 'checkbox') {
				await input.check();
			} else if (type === 'date') {
				await input.fill('2025-09-15');
			} else if (type === 'time') {
				await input.fill('09:00');
			} else if (type === 'number') {
				await input.fill('42');
			} else if (type === 'email') {
				await input.fill('e2e@example.com');
			} else if (type === 'tel') {
				await input.fill('555-0100');
			} else if (type === 'range') {
				continue;
			} else {
				await input.fill('Pilot answer');
			}
		}
		await page.screenshot({ path: path.join(SHOTS_DIR, '03-form-filled.png'), fullPage: true });

		await Promise.all([
			page.waitForURL(/\/r\/.+\/done/),
			page.locator('form button[type="submit"]').last().click()
		]);
		await expect(page.locator('text=/Submitted|Thank you/i').first()).toBeVisible();
		await page.screenshot({ path: path.join(SHOTS_DIR, '04-thank-you.png'), fullPage: true });
	});

	test('already-submitted token shows the already-submitted message', async ({ page }) => {
		const { token } = await newAssignmentToken();
		await page.goto(`/r/${token}`);
		await Promise.all([
			page.waitForURL(/\/r\/.+\/done/),
			page.locator('form button[type="submit"]').last().click()
		]);

		await page.goto(`/r/${token}`);
		await expect(page.locator('text=/Already submitted/i')).toBeVisible();
		await page.screenshot({
			path: path.join(SHOTS_DIR, '05-already-submitted.png'),
			fullPage: true
		});
	});
});
