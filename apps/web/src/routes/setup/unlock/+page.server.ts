import type { Actions } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { dev } from '$app/environment';
import { setupUnlock } from '$lib/server/setup';

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const data = await request.formData();
		const token = String(data.get('install_token') ?? '').trim();

		if (!token) {
			return fail(400, { error: 'Install token required.' });
		}

		const result = await setupUnlock(token);
		if (!result.ok) {
			return fail(result.status === 502 ? 502 : 400, { error: result.message });
		}

		const expiresAtMs = Date.parse(result.expiresAt);
		const maxAge = Math.max(0, Math.floor((expiresAtMs - Date.now()) / 1000));

		cookies.set('schoolrise_setup_session', result.sessionToken, {
			path: '/setup',
			httpOnly: true,
			secure: !dev,
			sameSite: 'lax',
			maxAge
		});

		throw redirect(303, '/setup/admin');
	}
};
