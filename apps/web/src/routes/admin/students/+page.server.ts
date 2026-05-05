import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { deleteStudent, listStudents } from '$lib/server/people';
import { getNode } from '$lib/server/tenancy';

function parseInstitutionId(value: string | null): number | null {
	if (!value) return null;
	const n = Number(value);
	return Number.isFinite(n) && n > 0 ? n : null;
}

export const load: PageServerLoad = async ({ url, cookies, locals }) => {
	if (!locals.user) {
		throw redirect(303, '/login');
	}

	const token = cookies.get('schoolrise_session') ?? '';
	const institutionId = parseInstitutionId(url.searchParams.get('institution'));

	if (institutionId === null) {
		return { institution: null, students: [] };
	}

	const [institution, students] = await Promise.all([
		getNode({ token }, institutionId),
		listStudents({ token }, institutionId)
	]);

	return { institution, institutionId, students };
};

export const actions: Actions = {
	delete: async ({ request, cookies, url }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const id = Number(data.get('id'));

		if (!Number.isFinite(id) || id <= 0) {
			return fail(400, { error: 'Invalid student id.' });
		}

		const result = await deleteStudent({ token }, id);
		if (!result.ok) {
			return fail(400, { error: result.message ?? 'Could not delete.' });
		}

		throw redirect(303, url.pathname + url.search);
	}
};
