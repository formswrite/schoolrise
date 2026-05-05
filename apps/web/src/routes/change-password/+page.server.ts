import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { changePasswordRequest } from '$lib/server/encore';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user) {
		throw redirect(303, '/login');
	}

	return { user: locals.user };
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const data = await request.formData();
		const current = String(data.get('current_password') ?? '');
		const next = String(data.get('new_password') ?? '');
		const confirm = String(data.get('confirm_password') ?? '');

		if (!current || !next || !confirm) {
			return fail(400, { error: 'All fields are required.' });
		}

		if (next !== confirm) {
			return fail(400, { error: 'New password and confirmation do not match.' });
		}

		if (next.length < 8) {
			return fail(400, { error: 'New password must be at least 8 characters.' });
		}

		const token = cookies.get('schoolrise_session') ?? '';

		const result = await changePasswordRequest(token, current, next);
		if (!result.ok) {
			return fail(result.status === 502 ? 502 : 400, { error: result.message });
		}

		cookies.delete('schoolrise_session', { path: '/' });

		throw redirect(303, '/login?changed=1');
	}
};
