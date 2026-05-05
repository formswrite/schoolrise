import { encoreApiUrl } from './encore';

type Opts = { token: string };

export type SuggestedItem = {
	type: string;
	title: string;
	required: boolean;
	options?: string[];
};

export type RubricBand = {
	band_code: string;
	description: string;
	min_score: number;
	max_score: number;
};

export type AIJob = {
	id: number;
	kind: string;
	model: string;
	status: 'pending' | 'running' | 'done' | 'failed';
	prompt_summary: string;
	request_tokens: number;
	response_tokens: number;
	latency_ms: number;
	error?: string;
	created_at: string;
	completed_at?: string;
};

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

export async function suggestItems(
	{ token }: Opts,
	body: { topic: string; scale_code: string; niveau_label: string; count: number }
): Promise<{ ok: boolean; items: SuggestedItem[]; error?: string }> {
	const res = await req('/v1/ai/suggest-items', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	if (!res.ok) return { ok: false, items: [], error: data?.message ?? 'Suggest failed.' };
	return { ok: true, items: data.items ?? [] };
}

export async function generateDistractors(
	{ token }: Opts,
	body: { question_title: string; correct_answer: string; count: number }
): Promise<{ ok: boolean; distractors: string[]; error?: string }> {
	const res = await req('/v1/ai/generate-distractors', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	if (!res.ok) return { ok: false, distractors: [], error: data?.message ?? 'Generate failed.' };
	return { ok: true, distractors: data.distractors ?? [] };
}

export async function getProviderStatus({ token }: Opts) {
	const res = await req('/v1/ai/provider', token);
	if (!res.ok) return null;
	return await res.json() as { provider: string; model: string };
}

export async function listJobs({ token }: Opts, limit = 50): Promise<AIJob[]> {
	const res = await req(`/v1/ai/jobs?limit=${limit}`, token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.jobs ?? [];
}
