import type { Actions, PageServerLoad } from './$types';
import { error, fail, redirect } from '@sveltejs/kit';
import { createStaff } from '$lib/server/people';
import { getNode } from '$lib/server/tenancy';

export const load: PageServerLoad = async ({ url, cookies, locals }) => {
	if (!locals.user) {
		throw redirect(303, '/login');
	}

	const scopeNodeId = Number(url.searchParams.get('scope'));

	if (!Number.isFinite(scopeNodeId) || scopeNodeId <= 0) {
		throw error(400, 'scope query parameter required');
	}

	const token = cookies.get('schoolrise_session') ?? '';
	const node = await getNode({ token }, scopeNodeId);

	return { scopeNodeId, node };
};

export const actions: Actions = {
	default: async ({ request, cookies, url }) => {
		const scopeNodeId = Number(url.searchParams.get('scope'));

		if (!Number.isFinite(scopeNodeId) || scopeNodeId <= 0) {
			return fail(400, { error: 'Missing scope.' });
		}

		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const fullName = String(data.get('full_name') ?? '').trim();

		if (!fullName) {
			return fail(400, { error: 'Full name required.' });
		}

		const result = await createStaff(
			{ token },
			{
				scopeNodeId,
				fullName,
				position: String(data.get('position') ?? '').trim() || undefined,
				staffCode: String(data.get('staff_code') ?? '').trim() || undefined,
				hireDate: String(data.get('hire_date') ?? '').trim() || undefined,
				givenName: String(data.get('given_name') ?? '').trim() || undefined,
				familyName: String(data.get('family_name') ?? '').trim() || undefined,
				email: String(data.get('email') ?? '').trim() || undefined,
				phone: String(data.get('phone') ?? '').trim() || undefined
			}
		);

		if (!result.ok) {
			return fail(result.status, { error: result.message ?? 'Could not create staff.' });
		}

		throw redirect(303, `/admin/staff?scope=${scopeNodeId}`);
	}
};
