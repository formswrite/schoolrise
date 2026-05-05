import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { dev } from '$app/environment';
import { loginRequest } from '$lib/server/encore';

export const load: PageServerLoad = async ({ locals }) => {
	if (locals.user) {
		throw redirect(303, '/');
	}

	return {};
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const data = await request.formData();
		const email = String(data.get('email') ?? '').trim();
		const password = String(data.get('password') ?? '');

		if (!email || !password) {
			return fail(400, { email, error: 'Email and password are required.' });
		}

		const result = await loginRequest(email, password);
		if (!result.ok) {
			return fail(result.status === 502 ? 502 : 401, { email, error: result.message });
		}

		const expiresAtMs = Date.parse(result.expiresAt);
		const maxAge = Math.max(0, Math.floor((expiresAtMs - Date.now()) / 1000));

		cookies.set('schoolrise_session', result.sessionToken, {
			path: '/',
			httpOnly: true,
			secure: !dev,
			sameSite: 'lax',
			maxAge
		});

		throw redirect(303, result.mustChangePassword ? '/change-password' : '/');
	}
};
