import { test, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

const SHOTS_DIR = 'tests/e2e/screenshots/14-uploads';
const FIXTURES = 'tests/e2e/fixtures';

test.describe.configure({ mode: 'default' });

test.describe('MinIO uploads — file goes to MinIO and is fetchable back', () => {
	test.beforeAll(() => {
		fs.mkdirSync(SHOTS_DIR, { recursive: true });
		fs.mkdirSync(FIXTURES, { recursive: true });
		const png = Buffer.from(
			'89504e470d0a1a0a0000000d49484452000000010000000108060000001f15c4890000000d4944415478' +
				'9c63fcffff3f0005fe02fe418600000000049454e44ae426082',
			'hex'
		);
		fs.writeFileSync(path.join(FIXTURES, 'tiny.png'), png);
	});

	test('POST /api/uploads stores the file in MinIO and returns key + url', async ({ request }) => {
		const file = fs.readFileSync(path.join(FIXTURES, 'tiny.png'));
		const res = await request.post('http://localhost:3001/api/uploads', {
			headers: { Origin: 'http://localhost:3001' },
			multipart: {
				file: { name: 'tiny.png', mimeType: 'image/png', buffer: file }
			}
		});
		expect(res.ok()).toBeTruthy();
		const body = await res.json();
		expect(body.key).toMatch(/^uploads\/\d{4}\/\d{2}\/.+\.png$/);
		expect(body.url).toMatch(/^\/api\/uploads\/uploads\/\d{4}\/\d{2}\/.+$/);
		expect(body.content_type).toBe('image/png');
		expect(body.size).toBeGreaterThan(0);

		const dl = await request.get(`http://localhost:3001/api/uploads/${body.key}`);
		expect(dl.ok()).toBeTruthy();
		const back = await dl.body();
		expect(back.length).toBe(body.size);
	});

	test('rejects unsupported content-type', async ({ request }) => {
		const res = await request.post('http://localhost:3001/api/uploads', {
			headers: { Origin: 'http://localhost:3001' },
			multipart: {
				file: {
					name: 'evil.exe',
					mimeType: 'application/x-msdownload',
					buffer: Buffer.from([0x4d, 0x5a, 0x90, 0x00])
				}
			}
		});
		expect(res.status()).toBe(415);
	});

	test('public renderer FILE_UPLOAD branch shows the FileUploadInput component', async ({
		page,
		request
	}) => {
		const login = await request.post('http://localhost:8080/v1/auth/login', {
			data: { email: 'admin@local.test', password: 'ChangeMe123!' }
		});
		const token = (await login.json()).SessionToken;

		const formRes = await request.post('http://localhost:8080/v1/forms', {
			headers: { Authorization: `Bearer ${token}` },
			data: { title: 'E2E uploads form', description: 'Phase 4u' }
		});
		const formId = (await formRes.json()).form?.id ?? (await formRes.json()).id;

		await request.post(`http://localhost:8080/v1/forms/items/${formId}/questions`, {
			headers: { Authorization: `Bearer ${token}` },
			data: {
				client_id: 'fu_' + Math.random().toString(36).slice(2, 10),
				sort_order: 10,
				title: 'Téléchargez une preuve',
				type: 'FILE_UPLOAD',
				required: false
			}
		});

		await page.goto(`/admin/forms/${formId}`);
		await page.waitForLoadState('networkidle');
		await expect(page.locator('text=/Téléchargez une preuve/').first()).toBeVisible();
	});
});
