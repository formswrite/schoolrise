import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { createPeriod, listPeriods } from '$lib/server/academics';
import { encoreApiUrl } from '$lib/server/encore';

export const load: PageServerLoad = async ({ cookies, locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	const token = cookies.get('schoolrise_session') ?? '';
	const periods = await listPeriods({ token });
	return { periods };
};

export const actions: Actions = {
	create: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const body = {
			code: String(data.get('code') ?? '').trim(),
			label: String(data.get('label') ?? '').trim(),
			starts_on: String(data.get('starts_on') ?? '').trim(),
			ends_on: String(data.get('ends_on') ?? '').trim(),
			is_current: data.get('is_current') === 'on'
		};
		if (!body.code || !body.label || !body.starts_on || !body.ends_on) {
			return fail(400, { error: 'All fields except "is_current" are required.' });
		}
		const res = await createPeriod({ token }, body);
		if (!res.ok) {
			return fail(res.status || 500, { error: res.data?.message ?? 'Could not create period.' });
		}
		return { success: true };
	},

	setCurrent: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const id = Number(data.get('id'));
		if (!Number.isFinite(id) || id <= 0) {
			return fail(400, { error: 'Invalid period id.' });
		}
		const res = await fetch(`${encoreApiUrl}/v1/academics/periods/${id}/current`, {
			method: 'POST',
			headers: { Authorization: `Bearer ${token}` }
		});
		if (!res.ok) {
			return fail(res.status, { error: 'Could not change current period.' });
		}
		return { success: true };
	},

	delete: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const id = Number(data.get('id'));
		if (!Number.isFinite(id) || id <= 0) {
			return fail(400, { error: 'Invalid period id.' });
		}
		const res = await fetch(`${encoreApiUrl}/v1/academics/periods/${id}`, {
			method: 'DELETE',
			headers: { Authorization: `Bearer ${token}` }
		});
		if (!res.ok) {
			return fail(res.status, { error: 'Could not delete period.' });
		}
		return { success: true };
	}
};
