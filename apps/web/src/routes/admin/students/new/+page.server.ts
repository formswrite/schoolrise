import type { Actions, PageServerLoad } from './$types';
import { error, fail, redirect } from '@sveltejs/kit';
import { createStudent } from '$lib/server/people';
import { getNode } from '$lib/server/tenancy';

export const load: PageServerLoad = async ({ url, cookies, locals }) => {
	if (!locals.user) {
		throw redirect(303, '/login');
	}

	const institutionId = Number(url.searchParams.get('institution'));

	if (!Number.isFinite(institutionId) || institutionId <= 0) {
		throw error(400, 'institution query parameter required');
	}

	const token = cookies.get('schoolrise_session') ?? '';
	const institution = await getNode({ token }, institutionId);

	return { institutionId, institution };
};

export const actions: Actions = {
	default: async ({ request, cookies, url }) => {
		const institutionId = Number(url.searchParams.get('institution'));

		if (!Number.isFinite(institutionId) || institutionId <= 0) {
			return fail(400, { error: 'Missing institution id.' });
		}

		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const fullName = String(data.get('full_name') ?? '').trim();

		if (!fullName) {
			return fail(400, { error: 'Full name required.' });
		}

		const result = await createStudent(
			{ token },
			{
				institutionId,
				fullName,
				givenName: String(data.get('given_name') ?? '').trim() || undefined,
				familyName: String(data.get('family_name') ?? '').trim() || undefined,
				studentCode: String(data.get('student_code') ?? '').trim() || undefined,
				enrollmentDate: String(data.get('enrollment_date') ?? '').trim() || undefined,
				gender: String(data.get('gender') ?? '').trim() || undefined,
				email: String(data.get('email') ?? '').trim() || undefined,
				phone: String(data.get('phone') ?? '').trim() || undefined
			}
		);

		if (!result.ok) {
			return fail(result.status, { error: result.message ?? 'Could not create student.' });
		}

		throw redirect(303, `/admin/students?institution=${institutionId}`);
	}
};
