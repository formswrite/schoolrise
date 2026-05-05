import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { listTeacherClasses } from '$lib/server/teacher';

export const load: LayoutServerLoad = async ({ cookies, locals, url }) => {
	if (!locals.user) {
		throw redirect(303, `/login?next=${encodeURIComponent(url.pathname)}`);
	}
	const token = cookies.get('schoolrise_session') ?? '';
	const { role } = await listTeacherClasses({ token });
	if (!role) {
		throw redirect(303, '/admin/dashboard');
	}
	return { user: locals.user, teacherRole: role };
};
