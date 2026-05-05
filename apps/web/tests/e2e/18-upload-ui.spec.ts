import { test, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';
import { adminCtx, buildPublishedFormWithToken } from './lib/forms-helpers';

test.describe.configure({ mode: 'default' });

const FIXTURES = 'tests/e2e/fixtures';

test.describe('FileUploadInput component — real file upload through the public renderer', () => {
	test.use({ storageState: { cookies: [], origins: [] } });

	let token: string;

	test.beforeAll(async () => {
		fs.mkdirSync(FIXTURES, { recursive: true });
		const png = Buffer.from(
			'89504e470d0a1a0a0000000d49484452000000010000000108060000001f15c4890000000d4944415478' +
				'9c63fcffff3f0005fe02fe418600000000049454e44ae426082',
			'hex'
		);
		fs.writeFileSync(path.join(FIXTURES, 'tiny.png'), png);

		const ctx = await adminCtx();
		const built = await buildPublishedFormWithToken(ctx, 'E2E upload-ui', [
			{ type: 'IMAGE', title: 'Photo de la salle de classe', required: false }
		]);
		token = built.token;
		await ctx.dispose();
	});

	test('user picks a PNG, sees progress then saved state, hidden input gets MinIO key', async ({
		page
	}) => {
		await page.goto(`/r/${token}`);
		await page.waitForLoadState('networkidle');

		const fileInput = page.locator('input[type="file"]').first();
		await fileInput.setInputFiles(path.join(FIXTURES, 'tiny.png'));

		await expect(page.locator('text=/tiny\\.png/')).toBeVisible({ timeout: 10000 });

		const hidden = page.locator('input[type="hidden"][name^="q_"]').first();
		const key = await hidden.getAttribute('value');
		expect(key).toMatch(/^uploads\/\d{4}\/\d{2}\/.+\.png$/);

		const dl = await page.request.get(`/api/uploads/${key}`);
		expect(dl.ok()).toBeTruthy();
		expect((await dl.body()).length).toBeGreaterThan(0);
	});
});
