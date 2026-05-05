import { encoreApiUrl } from './encore';

export type Campaign = {
	id: number;
	public_id: string;
	title: string;
	scale_code: string;
	form_id: number;
	form_version_id: number;
	period_id: number;
	scope_node_id: number;
	status: 'draft' | 'open' | 'closed';
	opens_at?: string;
	closes_at?: string;
	created_at: string;
};

export type Assignment = {
	id: number;
	campaign_id: number;
	student_id: number;
	access_token: string;
	submitted_at?: string;
	created_at: string;
};

export type Score = {
	id: number;
	response_id: number;
	campaign_id: number;
	student_id: number;
	raw_score: number;
	band_code: string;
	band_ordinal: number;
	finalized_at: string;
};

export type Scale = { code: string; label: string };

type Opts = { token: string };

async function req(path: string, token: string, init?: RequestInit) {
	return fetch(`${encoreApiUrl}${path}`, {
		...init,
		headers: {
			...(init?.headers ?? {}),
			Authorization: `Bearer ${token}`,
			'Content-Type': 'application/json'
		}
	});
}

export async function listCampaigns({ token }: Opts, scopeNodeID: number): Promise<Campaign[]> {
	const res = await req(`/v1/campaigns?scope_node_id=${scopeNodeID}`, token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.campaigns ?? [];
}

export async function getCampaign({ token }: Opts, id: number): Promise<Campaign | null> {
	const res = await req(`/v1/campaigns/${id}`, token);
	if (!res.ok) return null;
	return await res.json();
}

export async function createCampaign(
	{ token }: Opts,
	body: {
		title: string;
		scale_code: string;
		form_id: number;
		form_version_id: number;
		period_id: number;
		scope_node_id: number;
	}
) {
	const res = await req('/v1/campaigns', token, { method: 'POST', body: JSON.stringify(body) });
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function setCampaignStatus({ token }: Opts, id: number, action: 'open' | 'close') {
	const res = await req(`/v1/campaigns/${id}/${action}`, token, { method: 'POST' });
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function listAssignments({ token }: Opts, campaignID: number): Promise<Assignment[]> {
	const res = await req(`/v1/campaigns/${campaignID}/assignments`, token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.assignments ?? [];
}

export async function listScores({ token }: Opts, campaignID: number): Promise<Score[]> {
	const res = await req(`/v1/campaigns/${campaignID}/scores`, token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.scores ?? [];
}

export async function assignStudents(
	{ token }: Opts,
	campaignID: number,
	body: { student_ids: number[]; notify_by_email: boolean }
) {
	const res = await req(`/v1/campaigns/${campaignID}/assign`, token, {
		method: 'POST',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function listScales({ token }: Opts): Promise<Scale[]> {
	const res = await req('/v1/scales', token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.scales ?? [];
}
