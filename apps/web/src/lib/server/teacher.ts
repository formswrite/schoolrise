import { encoreApiUrl } from './encore';

export type TeacherClass = {
	id: number;
	period_id: number;
	niveau_id: number;
	institution_id: number;
	code: string;
	label: string;
};

export type TeacherCampaign = {
	id: number;
	public_id: string;
	title: string;
	scale_code: string;
	form_id: number;
	form_version_id: number;
	period_id: number;
	scope_node_id: number;
	status: string;
	assigned: number;
	scored: number;
	created_at: string;
};

export type GradingRosterRow = {
	student_id: number;
	full_name: string;
	student_code: string;
	has_score: boolean;
	raw_score?: number;
	band_code?: string;
	band_ordinal?: number;
	entry_mode?: string;
};

export type GradingRoster = {
	campaign: { id: number; title: string; scale_code: string; status: string };
	class_id: number;
	rows: GradingRosterRow[];
};

export type ProctoredEntry = {
	student_id: number;
	raw_score?: number;
	mode: 'proctored_score' | 'proctored_answers';
	answers?: unknown;
};

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

export async function listTeacherClasses(
	{ token }: Opts
): Promise<{ classes: TeacherClass[]; role: string }> {
	const res = await req('/v1/teacher/classes', token);
	if (!res.ok) return { classes: [], role: '' };
	return await res.json();
}

export async function listTeacherCampaigns(
	{ token }: Opts,
	classID: number
): Promise<TeacherCampaign[]> {
	const res = await req(`/v1/teacher/classes/${classID}/campaigns`, token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.campaigns ?? [];
}

export async function getGradingRoster(
	{ token }: Opts,
	classID: number,
	campaignID: number
): Promise<GradingRoster | null> {
	const res = await req(
		`/v1/teacher/classes/${classID}/campaigns/${campaignID}/roster`,
		token
	);
	if (!res.ok) return null;
	return await res.json();
}

export async function submitProctoredScores(
	{ token }: Opts,
	classID: number,
	campaignID: number,
	entries: ProctoredEntry[]
): Promise<{ ok: boolean; status: number; data: { created?: number; updated?: number; errors?: Array<{ student_id: number; message: string }> } }> {
	const res = await req(
		`/v1/teacher/classes/${classID}/campaigns/${campaignID}/scores`,
		token,
		{ method: 'POST', body: JSON.stringify({ entries }) }
	);
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data };
}
