import { encoreApiUrl } from './encore';

export type Form = {
	id: number;
	public_id: string;
	owner_id: number;
	title: string;
	description: string;
	status: 'draft' | 'published' | 'closed';
	settings: Record<string, unknown>;
	published_at?: string;
	created_at: string;
	updated_at: string;
};

export type Question = {
	id: number;
	form_id: number;
	client_id: string;
	sort_order: number;
	title: string;
	description: string;
	type: string;
	required: boolean;
	options: unknown[];
	scale_min?: number;
	scale_max?: number;
	scale_labels: Record<string, unknown>;
	validation: Record<string, unknown>;
	grading: Record<string, unknown>;
	extra: Record<string, unknown>;
};

export type FormVersion = {
	id: number;
	form_id: number;
	version_num: number;
	title: string;
	description: string;
	snapshot: {
		title: string;
		description: string;
		settings: Record<string, unknown>;
		questions: Question[];
	};
	published_at: string;
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

export async function listForms({ token }: Opts): Promise<Form[]> {
	const res = await req('/v1/forms', token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.forms ?? [];
}

export async function createForm({ token }: Opts, body: { title: string; description: string }) {
	const res = await req('/v1/forms', token, { method: 'POST', body: JSON.stringify(body) });
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function getForm({ token }: Opts, id: number): Promise<{ form: Form; questions: Question[] } | null> {
	const res = await req(`/v1/forms/items/${id}`, token);
	if (!res.ok) return null;
	return await res.json();
}

export async function addQuestion(
	{ token }: Opts,
	formId: number,
	body: Partial<Question> & { client_id: string; type: string }
) {
	const res = await req(`/v1/forms/items/${formId}/questions`, token, {
		method: 'POST',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function updateQuestion(
	{ token }: Opts,
	questionId: number,
	body: Partial<Question>
) {
	const res = await req(`/v1/forms/questions/${questionId}`, token, {
		method: 'PUT',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function deleteQuestion({ token }: Opts, questionId: number) {
	const res = await req(`/v1/forms/questions/${questionId}`, token, { method: 'DELETE' });
	return { ok: res.ok, status: res.status } as const;
}

export async function updateForm(
	{ token }: Opts,
	formId: number,
	body: { title?: string; description?: string; settings?: Record<string, unknown> }
) {
	const res = await req(`/v1/forms/items/${formId}`, token, {
		method: 'PUT',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function publishForm({ token }: Opts, formId: number) {
	const res = await req(`/v1/forms/items/${formId}/publish`, token, { method: 'POST' });
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function getPublicVersion(versionId: number): Promise<FormVersion | null> {
	const res = await fetch(`${encoreApiUrl}/v1/forms/public-versions/${versionId}`);
	if (!res.ok) return null;
	return await res.json();
}

export type AssignmentLookup = {
	assignment: {
		id: number;
		campaign_id: number;
		student_id: number;
		access_token: string;
		submitted_at?: string;
	};
	campaign: {
		id: number;
		title: string;
		scale_code: string;
		status: string;
	};
	form_version_id: number;
};

export async function lookupAssignment(token: string): Promise<AssignmentLookup | null> {
	const res = await fetch(`${encoreApiUrl}/v1/responses/lookup?token=${encodeURIComponent(token)}`);
	if (!res.ok) return null;
	return await res.json();
}

export async function submitResponse(body: {
	access_token: string;
	payload: Record<string, unknown>;
	raw_score: number;
}) {
	const res = await fetch(`${encoreApiUrl}/v1/responses`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}
