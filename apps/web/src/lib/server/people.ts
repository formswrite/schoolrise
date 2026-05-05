import { encoreApiUrl } from './encore';

export type Person = {
	id: number;
	fullName: string;
	givenName: string;
	familyName: string;
	email: string;
	phone: string;
	gender: string;
};

export type Student = {
	id: number;
	personId: number;
	institutionId: number;
	studentCode: string;
	enrollmentDate: string | null;
	person: Person;
	createdAt: string;
};

export type Staff = {
	id: number;
	personId: number;
	scopeNodeId: number;
	position: string;
	staffCode: string;
	hireDate: string | null;
	person: Person;
	createdAt: string;
};

type Auth = { token: string };

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

export async function listStudents({ token }: Auth, institutionId: number): Promise<Student[]> {
	const res = await authedFetch(`/v1/people/students?institutionId=${institutionId}`, token);
	if (!res.ok) return [];

	const data = await res.json();
	return (data.students ?? []) as Student[];
}

export async function getStudent({ token }: Auth, id: number): Promise<Student | null> {
	const res = await authedFetch(`/v1/people/students/${id}`, token);
	if (!res.ok) return null;

	const data = await res.json();
	return (data.student ?? null) as Student | null;
}

export type CreateStudentInput = {
	institutionId: number;
	studentCode?: string;
	enrollmentDate?: string;
	fullName: string;
	givenName?: string;
	familyName?: string;
	dateOfBirth?: string;
	gender?: string;
	email?: string;
	phone?: string;
};

export async function createStudent(
	{ token }: Auth,
	input: CreateStudentInput
): Promise<{ ok: boolean; status: number; message?: string; student?: Student }> {
	const body: Record<string, unknown> = {
		InstitutionId: input.institutionId,
		FullName: input.fullName
	};

	if (input.studentCode) body.StudentCode = input.studentCode;
	if (input.enrollmentDate) body.EnrollmentDate = input.enrollmentDate;
	if (input.givenName) body.GivenName = input.givenName;
	if (input.familyName) body.FamilyName = input.familyName;
	if (input.dateOfBirth) body.DateOfBirth = input.dateOfBirth;
	if (input.gender) body.Gender = input.gender;
	if (input.email) body.Email = input.email;
	if (input.phone) body.Phone = input.phone;

	const res = await authedFetch('/v1/people/students', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});

	if (res.ok) {
		const data = await res.json();
		return { ok: true, status: res.status, student: data.student as Student };
	}

	let message = `Request failed (${res.status})`;
	try {
		const data = await res.json();
		if (data?.message) message = String(data.message);
	} catch {}

	return { ok: false, status: res.status, message };
}

export async function deleteStudent(
	{ token }: Auth,
	id: number
): Promise<{ ok: boolean; message?: string }> {
	const res = await authedFetch(`/v1/people/students/${id}`, token, { method: 'DELETE' });
	if (res.ok) return { ok: true };

	let message = `Delete failed (${res.status})`;
	try {
		const data = await res.json();
		if (data?.message) message = String(data.message);
	} catch {}

	return { ok: false, message };
}

export async function listStaff({ token }: Auth, scopeNodeId: number): Promise<Staff[]> {
	const res = await authedFetch(`/v1/people/staff?scopeNodeId=${scopeNodeId}`, token);
	if (!res.ok) return [];

	const data = await res.json();
	return (data.staff ?? []) as Staff[];
}

export type CreateStaffInput = {
	scopeNodeId: number;
	position?: string;
	staffCode?: string;
	hireDate?: string;
	fullName: string;
	givenName?: string;
	familyName?: string;
	email?: string;
	phone?: string;
};

export async function createStaff(
	{ token }: Auth,
	input: CreateStaffInput
): Promise<{ ok: boolean; status: number; message?: string; staff?: Staff }> {
	const body: Record<string, unknown> = {
		ScopeNodeId: input.scopeNodeId,
		FullName: input.fullName
	};

	if (input.position) body.Position = input.position;
	if (input.staffCode) body.StaffCode = input.staffCode;
	if (input.hireDate) body.HireDate = input.hireDate;
	if (input.givenName) body.GivenName = input.givenName;
	if (input.familyName) body.FamilyName = input.familyName;
	if (input.email) body.Email = input.email;
	if (input.phone) body.Phone = input.phone;

	const res = await authedFetch('/v1/people/staff', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});

	if (res.ok) {
		const data = await res.json();
		return { ok: true, status: res.status, staff: data.staff as Staff };
	}

	let message = `Request failed (${res.status})`;
	try {
		const data = await res.json();
		if (data?.message) message = String(data.message);
	} catch {}

	return { ok: false, status: res.status, message };
}

export async function deleteStaff(
	{ token }: Auth,
	id: number
): Promise<{ ok: boolean; message?: string }> {
	const res = await authedFetch(`/v1/people/staff/${id}`, token, { method: 'DELETE' });
	if (res.ok) return { ok: true };

	let message = `Delete failed (${res.status})`;
	try {
		const data = await res.json();
		if (data?.message) message = String(data.message);
	} catch {}

	return { ok: false, message };
}
