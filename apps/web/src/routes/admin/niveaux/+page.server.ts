import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { createNiveau, listNiveaux } from '$lib/server/academics';
import { encoreApiUrl } from '$lib/server/encore';

export const load: PageServerLoad = async ({ cookies, locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	const token = cookies.get('schoolrise_session') ?? '';
	const niveaux = await listNiveaux({ token });
	return { niveaux };
};

export const actions: Actions = {
	create: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const body = {
			code: String(data.get('code') ?? '').trim(),
			label: String(data.get('label') ?? '').trim(),
			sort_order: Number(data.get('sort_order') ?? 0)
		};
		if (!body.code || !body.label) {
			return fail(400, { error: 'Code and label are required.' });
		}
		const res = await createNiveau({ token }, body);
		if (!res.ok) {
			return fail(res.status || 500, { error: res.data?.message ?? 'Could not create niveau.' });
		}
		return { success: true };
	},

	delete: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const id = Number(data.get('id'));
		if (!Number.isFinite(id) || id <= 0) {
			return fail(400, { error: 'Invalid niveau id.' });
		}
		const res = await fetch(`${encoreApiUrl}/v1/academics/niveaux/${id}`, {
			method: 'DELETE',
			headers: { Authorization: `Bearer ${token}` }
		});
		if (!res.ok) {
			return fail(res.status, { error: 'Could not delete niveau.' });
		}
		return { success: true };
	}
};
