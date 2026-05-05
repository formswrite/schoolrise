import { encoreApiUrl } from './encore';

export type Enrollment = {
	id: number;
	student_id: number;
	institution_id: number;
	period_id: number;
	status: 'active' | 'transferred' | 'dropped' | 'graduated' | 'reinstated';
	enrolled_on: string;
	ended_on?: string;
};

export type Coverage = {
	scope_node_id: number;
	period_id: number;
	total_enrolled: number;
	male: number;
	female: number;
	other: number;
	unknown: number;
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

export async function listEnrollments(
	{ token }: Opts,
	institutionId: number,
	periodId: number,
	includeInactive = false
): Promise<Enrollment[]> {
	const url = `/v1/enrollment/enrollments?institution_id=${institutionId}&period_id=${periodId}&include_inactive=${includeInactive}`;
	const res = await req(url, token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.enrollments ?? [];
}

export async function getCoverage(
	{ token }: Opts,
	scopeNodeId: number,
	periodId: number
): Promise<Coverage | null> {
	const res = await req(
		`/v1/enrollment/coverage?scope_node_id=${scopeNodeId}&period_id=${periodId}`,
		token
	);
	if (!res.ok) return null;
	return (await res.json()) as Coverage;
}

export async function createEnrollment(
	{ token }: Opts,
	body: {
		student_id: number;
		institution_id: number;
		period_id: number;
		enrolled_on: string;
		note?: string;
	}
) {
	const res = await req('/v1/enrollment/enrollments', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function transferEnrollment(
	{ token }: Opts,
	body: {
		student_id: number;
		period_id: number;
		to_institution_id: number;
		effective_on?: string;
		note?: string;
	}
) {
	const res = await req('/v1/enrollment/transfers', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function dropEnrollment(
	{ token }: Opts,
	body: { enrollment_id: number; ended_on?: string; note?: string }
) {
	const res = await req('/v1/enrollment/drop', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}
