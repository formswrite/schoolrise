import { test, expect } from '@playwright/test';
import { adminCtx, buildPublishedFormWithToken } from './lib/forms-helpers';

test.describe.configure({ mode: 'default' });

test.describe('SignaturePad — draw on canvas, save, MinIO upload', () => {
	test.use({ storageState: { cookies: [], origins: [] } });

	let token: string;

	test.beforeAll(async () => {
		const ctx = await adminCtx();
		const built = await buildPublishedFormWithToken(ctx, 'E2E signature', [
			{ type: 'SIGNATURE', title: 'Signez ici', required: false }
		]);
		token = built.token;
		await ctx.dispose();
	});

	test('user draws on canvas, saves, hidden input gets MinIO key', async ({ page }) => {
		await page.goto(`/r/${token}`);
		await page.waitForLoadState('networkidle');

		const canvas = page.locator('canvas').first();
		await expect(canvas).toBeVisible();

		const box = await canvas.boundingBox();
		if (!box) throw new Error('canvas has no bounding box');

		await page.mouse.move(box.x + 20, box.y + 20);
		await page.mouse.down();
		await page.mouse.move(box.x + 80, box.y + 60, { steps: 10 });
		await page.mouse.move(box.x + 200, box.y + 100, { steps: 10 });
		await page.mouse.up();

		const saveBtn = page.locator('button:has-text("Save signature")');
		await expect(saveBtn).toBeEnabled();
		await saveBtn.click();

		await expect(page.locator('text=/Signature saved/')).toBeVisible({ timeout: 10000 });

		const hidden = page.locator('input[type="hidden"][name^="q_"]').first();
		const key = await hidden.getAttribute('value');
		expect(key).toMatch(/^uploads\/\d{4}\/\d{2}\/.+\.png$/);
	});
});
