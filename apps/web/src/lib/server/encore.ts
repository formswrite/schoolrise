import { env } from '$env/dynamic/private';

export const encoreApiUrl = env.ENCORE_API_URL ?? 'http://localhost:8080';

export type RoleAssignment = {
	role: string;
	scopeNodeId: number | null;
};

export type Session = {
	userId: number;
	email: string;
	fullName: string;
	role: string;
	mustChangePassword: boolean;
	assignments: RoleAssignment[];
};

export async function fetchSession(token: string): Promise<Session | null> {
	if (!token) return null;

	try {
		const res = await fetch(`${encoreApiUrl}/v1/auth/me`, {
			headers: { Authorization: `Bearer ${token}` }
		});

		if (!res.ok) return null;

		const data = await res.json();
		const assignments: RoleAssignment[] = Array.isArray(data.Assignments)
			? data.Assignments.map((a: { role?: string; scopeNodeId?: number | null; Role?: string; ScopeNodeID?: number | null }) => ({
					role: (a.role ?? a.Role ?? '').toString(),
					scopeNodeId: a.scopeNodeId ?? a.ScopeNodeID ?? null
				}))
			: [];
		return {
			userId: data.UserID,
			email: data.Email,
			fullName: data.FullName,
			role: data.Role,
			mustChangePassword: data.MustChangePassword,
			assignments
		};
	} catch {
		return null;
	}
}

export type RoleSummary = {
	isGlobalAdmin: boolean;
	isInspector: boolean;
	isTeacher: boolean;
};

export function summarizeRoles(session: Session | null): RoleSummary {
	const out: RoleSummary = { isGlobalAdmin: false, isInspector: false, isTeacher: false };
	if (!session) return out;
	for (const a of session.assignments) {
		if (a.role === 'admin' && a.scopeNodeId === null) out.isGlobalAdmin = true;
		if (a.role === 'inspector') out.isInspector = true;
		if (a.role === 'teacher') out.isTeacher = true;
	}
	return out;
}

export type LoginResult =
	| { ok: true; sessionToken: string; expiresAt: string; mustChangePassword: boolean }
	| { ok: false; status: number; message: string };

export async function loginRequest(email: string, password: string): Promise<LoginResult> {
	try {
		const res = await fetch(`${encoreApiUrl}/v1/auth/login`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ Email: email, Password: password })
		});

		if (res.ok) {
			const data = await res.json();
			return {
				ok: true,
				sessionToken: data.SessionToken,
				expiresAt: data.ExpiresAt,
				mustChangePassword: data.MustChangePassword
			};
		}

		const message =
			res.status === 401
				? 'Invalid credentials'
				: res.status === 412
					? 'Account locked. Contact your administrator.'
					: 'Login failed.';

		return { ok: false, status: res.status, message };
	} catch {
		return { ok: false, status: 502, message: 'Authentication service unreachable.' };
	}
}

export type ChangePasswordResult =
	| { ok: true }
	| { ok: false; status: number; message: string };

export async function changePasswordRequest(
	token: string,
	currentPassword: string,
	newPassword: string
): Promise<ChangePasswordResult> {
	try {
		const res = await fetch(`${encoreApiUrl}/v1/auth/change-password`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${token}`
			},
			body: JSON.stringify({ CurrentPassword: currentPassword, NewPassword: newPassword })
		});

		if (res.ok) return { ok: true };

		const message =
			res.status === 401
				? 'Current password is incorrect.'
				: res.status === 400
					? 'New password must be at least 8 characters.'
					: 'Could not change password.';

		return { ok: false, status: res.status, message };
	} catch {
		return { ok: false, status: 502, message: 'Authentication service unreachable.' };
	}
}

export async function logoutRequest(token: string): Promise<void> {
	if (!token) return;

	try {
		await fetch(`${encoreApiUrl}/v1/auth/logout`, {
			method: 'POST',
			headers: { Authorization: `Bearer ${token}` }
		});
	} catch {
	}
}
