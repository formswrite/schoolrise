import { test, expect, type Page } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const SHOTS_DIR = 'tests/e2e/screenshots/01-smoke';

const PAGES: Array<{ slug: string; url: string; expectH1?: RegExp; description: string }> = [
	{ slug: '00-dashboard',           url: '/admin/dashboard',                                  expectH1: /Progression dashboard/i, description: 'Empty dashboard (no scope picked)' },
	{ slug: '01-dashboard-with-data', url: '/admin/dashboard?scope=1&campaign=1&period=1',      expectH1: /Progression dashboard/i, description: 'Dashboard with band distribution + drilldown' },
	{ slug: '02-institutions-root',   url: '/admin/institutions',                               expectH1: /Institutions/i,         description: 'Hierarchy browser at root' },
	{ slug: '03-institutions-region', url: '/admin/institutions?parent=1',                      expectH1: /Institutions/i,         description: 'Drilled into region 1' },
	{ slug: '04-periods',             url: '/admin/periods',                                    expectH1: /Academic periods/i,     description: 'Period registry' },
	{ slug: '05-niveaux',             url: '/admin/niveaux',                                    expectH1: /Niveaux/i,              description: 'Grade level registry' },
	{ slug: '06-classes-empty',       url: '/admin/classes',                                    expectH1: /Classes/i,              description: 'Classes (no institution picked)' },
	{ slug: '07-classes',             url: '/admin/classes?institution=2',                     expectH1: /Classes/i,              description: 'Classes at school 2' },
	{ slug: '08-students-empty',      url: '/admin/students',                                   expectH1: /Students/i,             description: 'Students (no institution picked)' },
	{ slug: '09-students',            url: '/admin/students?institution=2',                    expectH1: /Students/i,             description: 'Students at school 2' },
	{ slug: '10-staff-empty',         url: '/admin/staff',                                      expectH1: /Staff/i,                description: 'Staff (no scope picked)' },
	{ slug: '11-staff',               url: '/admin/staff?scope=2',                              expectH1: /Staff/i,                description: 'Staff at school 2' },
	{ slug: '12-enrollment-empty',    url: '/admin/enrollment',                                 expectH1: /Enrollment/i,           description: 'Enrollment (no scope)' },
	{ slug: '13-enrollment',          url: '/admin/enrollment?scope=2',                        expectH1: /Enrollment/i,           description: 'Enrollment with coverage tiles' },
	{ slug: '14-forms-list',          url: '/admin/forms',                                      expectH1: /Forms/i,                description: 'Form list' },
	{ slug: '15-forms-edit',          url: '/admin/forms/1',                                    expectH1: /.+/,                    description: 'Form builder for "Email Test"' },
	{ slug: '16-imports',             url: '/admin/imports',                                    expectH1: /CSV import/i,           description: 'CSV upload' },
	{ slug: '17-notifications',       url: '/admin/notifications',                              expectH1: /Notifications/i,        description: 'Email outbox + provider' },
	{ slug: '18-users',               url: '/admin/users',                                      expectH1: /Users/i,                description: 'User accounts' }
];

test.describe('UI smoke walk — admin', () => {
	test.beforeAll(() => {
		fs.mkdirSync(SHOTS_DIR, { recursive: true });
	});

	for (const p of PAGES) {
		test(`${p.slug} — ${p.description}`, async ({ page }) => {
			const errors: string[] = [];
			page.on('pageerror', (err) => errors.push(`pageerror: ${err.message}`));
			page.on('console', (msg) => {
				if (msg.type() === 'error') errors.push(`console.error: ${msg.text()}`);
			});

			const resp = await page.goto(p.url, { waitUntil: 'networkidle' });
			expect(resp?.status(), `HTTP for ${p.url}`).toBeLessThan(400);

			if (p.expectH1) {
				await expect(page.locator('h1, h2').first()).toBeVisible();
			}

			await page.waitForTimeout(400);
			const file = path.join(SHOTS_DIR, `${p.slug}.png`);
			await page.screenshot({ path: file, fullPage: true });

			expect(errors, `Console/page errors on ${p.url}`).toEqual([]);
		});
	}
});

test.describe('Public-facing pages (no auth)', () => {
	test.use({ storageState: { cookies: [], origins: [] } });

	const PUBLIC: Array<{ slug: string; url: string }> = [
		{ slug: '90-login',        url: '/login' },
		{ slug: '91-r-invalid',    url: '/r/this-is-not-a-real-token' },
		{ slug: '92-r-done',       url: '/r/this-is-not-a-real-token/done' }
	];

	for (const p of PUBLIC) {
		test(`${p.slug} — ${p.url}`, async ({ page }: { page: Page }) => {
			await page.goto(p.url, { waitUntil: 'networkidle' });
			fs.mkdirSync(SHOTS_DIR, { recursive: true });
			await page.screenshot({
				path: path.join(SHOTS_DIR, `${p.slug}.png`),
				fullPage: true
			});
		});
	}
});
