import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { createCampaign, listCampaigns, listScales } from '$lib/server/campaigns';
import { listForms } from '$lib/server/forms';
import { listPeriods } from '$lib/server/academics';
import { resolveDefaultScope } from '$lib/server/default-scope';

function parseId(v: string | null): number | null {
	if (!v) return null;
	const n = Number(v);
	return Number.isFinite(n) && n > 0 ? n : null;
}

export const load: PageServerLoad = async ({ url, cookies, locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	const token = cookies.get('schoolrise_session') ?? '';
	const requested = parseId(url.searchParams.get('scope'));

	const [periods, scales, forms, resolved] = await Promise.all([
		listPeriods({ token }),
		listScales({ token }),
		listForms({ token }),
		resolveDefaultScope(token, locals.user, requested)
	]);

	if (!resolved) {
		return {
			scope: null,
			scopeNodeID: null,
			scopeOptions: [],
			campaigns: [],
			periods,
			scales,
			forms
		};
	}

	const campaigns = await listCampaigns({ token }, resolved.scopeNodeId);
	return {
		scope: resolved.scope,
		scopeNodeID: resolved.scopeNodeId,
		scopeOptions: resolved.options,
		campaigns,
		periods,
		scales,
		forms
	};
};

export const actions: Actions = {
	create: async ({ request, cookies, url }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const scopeNodeID = parseId(url.searchParams.get('scope'));
		if (!scopeNodeID) return fail(400, { error: 'Pick a scope first.' });

		const body = {
			title: String(data.get('title') ?? '').trim(),
			scale_code: String(data.get('scale_code') ?? '').trim(),
			form_id: Number(data.get('form_id')),
			form_version_id: Number(data.get('form_version_id')),
			period_id: Number(data.get('period_id')),
			scope_node_id: scopeNodeID
		};
		if (
			!body.title ||
			!body.scale_code ||
			!body.form_id ||
			!body.form_version_id ||
			!body.period_id
		) {
			return fail(400, { error: 'All fields are required.' });
		}

		const res = await createCampaign({ token }, body);
		if (!res.ok) {
			return fail(res.status, { error: res.data?.message ?? 'Could not create campaign.' });
		}
		throw redirect(303, `/admin/campaigns/${res.data.id}`);
	}
};
