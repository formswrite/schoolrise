import { test, expect } from '@playwright/test';
import { adminCtx, buildPublishedFormWithToken } from './lib/forms-helpers';

test.describe.configure({ mode: 'default' });

test.describe('Public renderer — logic rules hide/show reactively', () => {
	test.use({ storageState: { cookies: [], origins: [] } });

	let token: string;

	test.beforeAll(async () => {
		const ctx = await adminCtx();
		const built = await buildPublishedFormWithToken(
			ctx,
			'E2E logic-render',
			[
				{ type: 'YES_NO', title: 'Avez-vous un téléphone ?', required: true },
				{ type: 'SHORT_ANSWER', title: 'Quel est votre numéro ?', required: false }
			],
			[
				{
					id: 'r_show_phone',
					target_question_client_id: '__Q1__',
					operator: 'show_if',
					conditions: [{ source_question_client_id: '__Q0__', op: 'equals', value: 'yes' }]
				}
			]
		);
		token = built.token;
		await ctx.dispose();
	});

	test('Q2 hidden initially; appears after Q1 = yes', async ({ page }) => {
		await page.goto(`/r/${token}`);
		await page.waitForLoadState('networkidle');

		await expect(page.locator('text=/Avez-vous un téléphone/')).toBeVisible();
		await expect(page.locator('text=/Quel est votre numéro/')).not.toBeVisible();

		await page.locator('input[type="radio"][value="yes"]').first().check();
		await expect(page.locator('text=/Quel est votre numéro/')).toBeVisible();

		await page.locator('input[type="radio"][value="no"]').first().check();
		await expect(page.locator('text=/Quel est votre numéro/')).not.toBeVisible();
	});
});
