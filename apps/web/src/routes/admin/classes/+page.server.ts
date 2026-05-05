import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import {
	createClass,
	deleteClass,
	listClassesByInstitution,
	listNiveaux,
	listPeriods
} from '$lib/server/academics';
import { resolveDefaultInstitution } from '$lib/server/default-scope';

function parseId(value: string | null): number | null {
	if (!value) return null;
	const n = Number(value);
	return Number.isFinite(n) && n > 0 ? n : null;
}

export const load: PageServerLoad = async ({ url, cookies, locals }) => {
	if (!locals.user) {
		throw redirect(303, '/login');
	}
	const token = cookies.get('schoolrise_session') ?? '';
	const requested = parseId(url.searchParams.get('institution'));

	const [periods, niveaux, resolved] = await Promise.all([
		listPeriods({ token }),
		listNiveaux({ token }),
		resolveDefaultInstitution(token, locals.user, requested)
	]);

	if (!resolved) {
		return {
			institution: null,
			institutionId: null,
			institutionOptions: [],
			classes: [],
			periods,
			niveaux
		};
	}

	const classes = await listClassesByInstitution({ token }, resolved.institutionId);
	return {
		institution: resolved.institution,
		institutionId: resolved.institutionId,
		institutionOptions: resolved.options,
		classes,
		periods,
		niveaux
	};
};

export const actions: Actions = {
	create: async ({ request, cookies, url }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const institutionId = parseId(url.searchParams.get('institution'));
		if (!institutionId) {
			return fail(400, { error: 'Pick an institution first.' });
		}
		const body = {
			period_id: Number(data.get('period_id')),
			niveau_id: Number(data.get('niveau_id')),
			institution_id: institutionId,
			code: String(data.get('code') ?? '').trim(),
			label: String(data.get('label') ?? '').trim(),
			capacity: Number(data.get('capacity') ?? 0)
		};
		if (!body.code || !body.label || !body.period_id || !body.niveau_id) {
			return fail(400, { error: 'Period, niveau, code, and label are required.' });
		}
		const res = await createClass({ token }, body);
		if (!res.ok) {
			return fail(res.status || 500, { error: res.data?.message ?? 'Could not create class.' });
		}
		return { success: true };
	},

	delete: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const id = Number(data.get('id'));
		if (!Number.isFinite(id) || id <= 0) {
			return fail(400, { error: 'Invalid class id.' });
		}
		const res = await deleteClass({ token }, id);
		if (!res.ok) {
			return fail(res.status || 500, { error: 'Could not delete class.' });
		}
		return { success: true };
	}
};
