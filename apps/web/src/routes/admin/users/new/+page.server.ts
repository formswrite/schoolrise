import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { createUser } from '$lib/server/admin';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user) {
		throw redirect(303, '/login');
	}

	return {};
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const email = String(data.get('email') ?? '').trim();
		const fullName = String(data.get('full_name') ?? '').trim();
		const password = String(data.get('password') ?? '');
		const role = String(data.get('role') ?? 'teacher').trim();
		const mustChange = data.get('must_change_password') === 'on';

		if (!email || !fullName || !password) {
			return fail(400, { email, fullName, role, error: 'All fields required.' });
		}

		const result = await createUser({ token }, email, fullName, password, role, mustChange);
		if (!result.ok) {
			return fail(result.status, {
				email,
				fullName,
				role,
				error: result.message ?? 'Could not create user.'
			});
		}

		throw redirect(303, `/admin/users/${result.user!.id}`);
	}
};
