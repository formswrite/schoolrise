import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { createForm, listForms } from '$lib/server/forms';

export const load: PageServerLoad = async ({ cookies, locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	const token = cookies.get('schoolrise_session') ?? '';
	const forms = await listForms({ token });
	return { forms };
};

export const actions: Actions = {
	create: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const title = String(data.get('title') ?? '').trim();
		const description = String(data.get('description') ?? '').trim();
		if (!title) return fail(400, { error: 'Title required.' });
		const res = await createForm({ token }, { title, description });
		if (!res.ok) return fail(res.status, { error: res.data?.message ?? 'Could not create.' });
		throw redirect(303, `/admin/forms/${res.data.id}`);
	}
};
