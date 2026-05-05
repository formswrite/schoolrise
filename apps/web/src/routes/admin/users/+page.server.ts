import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { listUsers } from '$lib/server/admin';

export const load: PageServerLoad = async ({ cookies, locals }) => {
	if (!locals.user) {
		throw redirect(303, '/login');
	}

	const token = cookies.get('schoolrise_session') ?? '';
	const users = await listUsers({ token });

	return { users };
};
