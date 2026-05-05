import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { setupCreateAdmin } from '$lib/server/setup';

export const load: PageServerLoad = async ({ cookies }) => {
	if (!cookies.get('schoolrise_setup_session')) {
		throw redirect(303, '/setup/unlock');
	}

	return {};
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const session = cookies.get('schoolrise_setup_session') ?? '';
		if (!session) {
			throw redirect(303, '/setup/unlock');
		}

		const data = await request.formData();
		const email = String(data.get('email') ?? '').trim();
		const fullName = String(data.get('full_name') ?? '').trim();
		const password = String(data.get('password') ?? '');
		const confirm = String(data.get('confirm_password') ?? '');

		if (!email || !fullName || !password) {
			return fail(400, { email, fullName, error: 'All fields required.' });
		}

		if (password !== confirm) {
			return fail(400, { email, fullName, error: 'Passwords do not match.' });
		}

		if (password.length < 8) {
			return fail(400, { email, fullName, error: 'Password must be at least 8 characters.' });
		}

		const result = await setupCreateAdmin(session, email, fullName, password);
		if (!result.ok) {
			return fail(result.status, { email, fullName, error: result.message ?? 'Could not create admin.' });
		}

		throw redirect(303, '/setup/system');
	}
};
