import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { getDrilldown, getProgression, listCampaignsByScope, refreshSnapshot } from '$lib/server/progression';
import { resolveDefaultScope } from '$lib/server/default-scope';

function parseId(value: string | null): number | null {
	if (!value) return null;
	const n = Number(value);
	return Number.isFinite(n) && n > 0 ? n : null;
}

export const load: PageServerLoad = async ({ url, cookies, locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	const token = cookies.get('schoolrise_session') ?? '';

	const requestedScope = parseId(url.searchParams.get('scope'));
	const requestedCampaign = parseId(url.searchParams.get('campaign'));
	const requestedPeriod = parseId(url.searchParams.get('period'));

	const resolved = await resolveDefaultScope(token, locals.user, requestedScope);
	if (!resolved) {
		return {
			scope: null,
			scopeNodeId: null,
			scopeOptions: [],
			campaigns: [],
			campaignId: null,
			periodId: null,
			progression: null,
			drilldown: null
		};
	}

	const { scopeNodeId, scope, options } = resolved;
	const campaigns = await listCampaignsByScope({ token }, scopeNodeId);

	let campaignId = requestedCampaign;
	let periodId = requestedPeriod;
	if (!campaignId && campaigns.length > 0) {
		const open = campaigns.find((c) => c.status === 'open') ?? campaigns[0];
		campaignId = open.id;
		periodId = open.period_id;
	}

	if (!campaignId || !periodId) {
		return {
			scope,
			scopeNodeId,
			scopeOptions: options,
			campaigns,
			campaignId: null,
			periodId: null,
			progression: null,
			drilldown: null
		};
	}

	const [progression, drilldown] = await Promise.all([
		getProgression({ token }, scopeNodeId, periodId, campaignId),
		getDrilldown({ token }, scopeNodeId, periodId, campaignId)
	]);

	return {
		scope,
		scopeNodeId,
		scopeOptions: options,
		campaigns,
		campaignId,
		periodId,
		progression,
		drilldown
	};
};

export const actions: Actions = {
	refresh: async ({ cookies, url }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const scope = parseId(url.searchParams.get('scope'));
		const campaign = parseId(url.searchParams.get('campaign'));
		const period = parseId(url.searchParams.get('period'));
		if (!scope || !campaign || !period) {
			return fail(400, { error: 'scope, period and campaign required' });
		}
		const res = await refreshSnapshot({ token }, scope, period, campaign);
		if (!res.ok) return fail(res.status, { error: 'Refresh failed.' });
		return { success: true };
	}
};
