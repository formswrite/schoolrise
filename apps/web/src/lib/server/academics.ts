import { encoreApiUrl } from './encore';

export type Period = {
	id: number;
	code: string;
	label: string;
	starts_on: string;
	ends_on: string;
	is_current: boolean;
};

export type Niveau = {
	id: number;
	code: string;
	label: string;
	sort_order: number;
};

export type Class = {
	id: number;
	period_id: number;
	niveau_id: number;
	institution_id: number;
	code: string;
	label: string;
	capacity?: number;
};

type Opts = { token: string };

async function authedFetch(path: string, token: string, init?: RequestInit) {
	return fetch(`${encoreApiUrl}${path}`, {
		...init,
		headers: {
			...(init?.headers ?? {}),
			Authorization: `Bearer ${token}`,
			'Content-Type': 'application/json'
		}
	});
}

export async function listPeriods({ token }: Opts): Promise<Period[]> {
	const res = await authedFetch('/v1/academics/periods', token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.periods ?? [];
}

export async function createPeriod(
	{ token }: Opts,
	body: { code: string; label: string; starts_on: string; ends_on: string; is_current: boolean }
) {
	const res = await authedFetch('/v1/academics/periods', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function listNiveaux({ token }: Opts): Promise<Niveau[]> {
	const res = await authedFetch('/v1/academics/niveaux', token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.niveaux ?? [];
}

export async function createNiveau(
	{ token }: Opts,
	body: { code: string; label: string; sort_order: number }
) {
	const res = await authedFetch('/v1/academics/niveaux', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function listClassesByInstitution(
	{ token }: Opts,
	institutionId: number
): Promise<Class[]> {
	const res = await authedFetch(`/v1/academics/classes?institution_id=${institutionId}`, token);
	if (!res.ok) return [];
	const data = await res.json();
	return data.classes ?? [];
}

export async function createClass(
	{ token }: Opts,
	body: {
		period_id: number;
		niveau_id: number;
		institution_id: number;
		code: string;
		label: string;
		capacity: number;
	}
) {
	const res = await authedFetch('/v1/academics/classes', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	return { ok: res.ok, status: res.status, data } as const;
}

export async function deleteClass({ token }: Opts, id: number) {
	const res = await authedFetch(`/v1/academics/classes/${id}`, token, { method: 'DELETE' });
	return { ok: res.ok, status: res.status } as const;
}

export async function getClass({ token }: Opts, id: number): Promise<Class | null> {
	const res = await authedFetch(`/v1/academics/classes/${id}`, token);
	if (!res.ok) return null;
	return await res.json();
}

export type ClassRoster = {
	student_ids: number[];
	staff: Array<{ staff_id: number; role: string }>;
};

export async function getClassRoster(
	{ token }: Opts,
	classID: number
): Promise<ClassRoster | null> {
	const res = await authedFetch(`/v1/academics/classes/${classID}/roster`, token);
	if (!res.ok) return null;
	return await res.json();
}

export async function addStudentToClass({ token }: Opts, classID: number, studentID: number) {
	const res = await authedFetch(`/v1/academics/classes/${classID}/students`, token, {
		method: 'POST',
		body: JSON.stringify({ student_id: studentID })
	});
	return { ok: res.ok, status: res.status } as const;
}

export async function removeStudentFromClass({ token }: Opts, classID: number, studentID: number) {
	const res = await authedFetch(`/v1/academics/classes/${classID}/students/${studentID}`, token, {
		method: 'DELETE'
	});
	return { ok: res.ok, status: res.status } as const;
}

export async function addStaffToClass(
	{ token }: Opts,
	classID: number,
	staffID: number,
	role = 'teacher'
) {
	const res = await authedFetch(`/v1/academics/classes/${classID}/staff`, token, {
		method: 'POST',
		body: JSON.stringify({ staff_id: staffID, role })
	});
	return { ok: res.ok, status: res.status } as const;
}

export async function removeStaffFromClass(
	{ token }: Opts,
	classID: number,
	staffID: number,
	role: string
) {
	const res = await authedFetch(
		`/v1/academics/classes/${classID}/staff/${staffID}/${encodeURIComponent(role)}`,
		token,
		{ method: 'DELETE' }
	);
	return { ok: res.ok, status: res.status } as const;
}
