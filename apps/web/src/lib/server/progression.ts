import { encoreApiUrl } from './encore';

export type BandRow = {
	band_code: string;
	band_ordinal: number;
	band_label: string;
	student_count: number;
	percentage: number;
};

export type ScopeProgression = {
	scope_node_id: number;
	period_id: number;
	campaign_id: number;
	total_scored: number;
	bands: BandRow[];
	generated_at: string;
};

export type Drilldown = {
	scope: ScopeProgression;
	children: Array<{
		node_id: number;
		code: string;
		label: string;
		level: string;
		bands: BandRow[];
		total: number;
	}>;
};

type Opts = { token: string };

async function req(path: string, token: string) {
	return fetch(`${encoreApiUrl}${path}`, {
		headers: { Authorization: `Bearer ${token}` }
	});
}

export async function getProgression(
	{ token }: Opts,
	scope: number,
	period: number,
	campaign: number
): Promise<ScopeProgression | null> {
	const res = await req(
		`/v1/progression?scope_node_id=${scope}&period_id=${period}&campaign_id=${campaign}`,
		token
	);
	if (!res.ok) return null;
	return await res.json();
}

export async function getDrilldown(
	{ token }: Opts,
	scope: number,
	period: number,
	campaign: number
): Promise<Drilldown | null> {
	const res = await req(
		`/v1/progression/drilldown?scope_node_id=${scope}&period_id=${period}&campaign_id=${campaign}`,
		token
	);
	if (!res.ok) return null;
	return await res.json();
}

export async function refreshSnapshot(
	{ token }: Opts,
	scope: number,
	period: number,
	campaign: number
) {
	const res = await fetch(
		`${encoreApiUrl}/v1/progression/refresh?scope_node_id=${scope}&period_id=${period}&campaign_id=${campaign}`,
		{
			method: 'POST',
			headers: { Authorization: `Bearer ${token}` }
		}
	);
	return { ok: res.ok, status: res.status } as const;
}

export type Campaign = {
	id: number;
	public_id: string;
	title: string;
	scale_code: string;
	form_id: number;
	form_version_id: number;
	period_id: number;
	scope_node_id: number;
	status: string;
	created_at: string;
};

export async function listCampaignsByScope({ token }: Opts, scope: number): Promise<Campaign[]> {
	const res = await req(`/v1/campaigns?scope_node_id=${scope}`, token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.campaigns ?? [];
}
