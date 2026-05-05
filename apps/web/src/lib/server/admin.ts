import { encoreApiUrl } from './encore';

export type AdminUser = {
	id: number;
	email: string;
	fullName: string;
	role: string;
	mustChangePassword: boolean;
	lockedAt: string | null;
	lastLoginAt: string | null;
	createdAt: string;
};

export type Assignment = {
	id: number;
	userId: number;
	role: string;
	scopeNodeId: number | null;
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

export async function listUsers({ token }: Auth): Promise<AdminUser[]> {
	const res = await authedFetch('/v1/auth/users', token);
	if (!res.ok) return [];

	const data = await res.json();
	return (data.users ?? []) as AdminUser[];
}

export async function getUser({ token }: Auth, id: number): Promise<AdminUser | null> {
	const res = await authedFetch(`/v1/auth/users/${id}`, token);
	if (!res.ok) return null;

	const data = await res.json();
	return (data.user ?? null) as AdminUser | null;
}

export async function createUser(
	{ token }: Auth,
	email: string,
	fullName: string,
	password: string,
	role: string,
	mustChangePassword: boolean
): Promise<{ ok: boolean; status: number; message?: string; user?: AdminUser }> {
	const res = await authedFetch('/v1/auth/users', token, {
		method: 'POST',
		body: JSON.stringify({
			Email: email,
			FullName: fullName,
			Password: password,
			Role: role,
			MustChangePassword: mustChangePassword
		})
	});

	if (res.ok) {
		const data = await res.json();
		return { ok: true, status: res.status, user: data.user as AdminUser };
	}

	let message = `Request failed (${res.status})`;
	try {
		const data = await res.json();
		if (data?.message) message = String(data.message);
	} catch {
	}

	return { ok: false, status: res.status, message };
}

export async function listUserAssignments({ token }: Auth, userId: number): Promise<Assignment[]> {
	const res = await authedFetch(`/v1/auth/users/${userId}/assignments`, token);
	if (!res.ok) return [];

	const data = await res.json();
	return (data.assignments ?? []) as Assignment[];
}

export async function createAssignment(
	{ token }: Auth,
	userId: number,
	role: string,
	scopeNodeId: number | null
): Promise<{ ok: boolean; status: number; message?: string }> {
	const body: Record<string, unknown> = { UserId: userId, Role: role };
	if (scopeNodeId !== null) body.ScopeNodeId = scopeNodeId;

	const res = await authedFetch('/v1/auth/assignments', token, {
		method: 'POST',
		body: JSON.stringify(body)
	});

	if (res.ok) return { ok: true, status: res.status };

	let message = `Request failed (${res.status})`;
	try {
		const data = await res.json();
		if (data?.message) message = String(data.message);
	} catch {
	}

	return { ok: false, status: res.status, message };
}

export async function deleteAssignment({ token }: Auth, id: number): Promise<{ ok: boolean; message?: string }> {
	const res = await authedFetch(`/v1/auth/assignments/${id}`, token, { method: 'DELETE' });
	if (res.ok) return { ok: true };

	let message = `Delete failed (${res.status})`;
	try {
		const data = await res.json();
		if (data?.message) message = String(data.message);
	} catch {
	}

	return { ok: false, message };
}
