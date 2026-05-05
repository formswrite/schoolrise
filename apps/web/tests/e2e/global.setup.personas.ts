import { test as setup, expect } from '@playwright/test';
import fs from 'node:fs';
import { seedPersonas, type Personas } from './lib/seed';

const PERSONAS_FILE = 'tests/e2e/.auth/personas.json';

setup('seed personas + log each one in', async ({ browser }) => {
	const personas = await seedPersonas();
	fs.mkdirSync('tests/e2e/.auth', { recursive: true });
	fs.writeFileSync(PERSONAS_FILE, JSON.stringify(personas, null, 2));

	for (const [name, creds] of [
		[
			'admin',
			{ email: personas.admin.email, password: personas.admin.password, file: 'admin.json' }
		],
		[
			'teacher',
			{ email: personas.teacher.email, password: personas.teacher.password, file: 'teacher.json' }
		],
		[
			'inspector',
			{
				email: personas.inspector.email,
				password: personas.inspector.password,
				file: 'inspector.json'
			}
		]
	] as const) {
		const page = await browser.newPage();
		await page.goto('http://localhost:3001/login');
		await page.locator('input[name="email"]').fill(creds.email);
		await page.locator('input[name="password"]').fill(creds.password);
		await Promise.all([
			page.waitForURL((url) => !url.pathname.startsWith('/login'), { timeout: 10_000 }),
			page.locator('form button[type="submit"]').click()
		]);
		await expect(page).not.toHaveURL(/\/login/);
		await page.context().storageState({ path: `tests/e2e/.auth/${creds.file}` });
		await page.close();
	}
});

export type { Personas };
