import type { Actions, PageServerLoad } from './$types';
import { error, fail, redirect } from '@sveltejs/kit';
import {
	createAssignment,
	deleteAssignment,
	getUser,
	listUserAssignments
} from '$lib/server/admin';

export const load: PageServerLoad = async ({ cookies, locals, params }) => {
	if (!locals.user) {
		throw redirect(303, '/login');
	}

	const token = cookies.get('schoolrise_session') ?? '';
	const id = Number(params.id);

	if (!Number.isFinite(id) || id <= 0) {
		throw error(400, 'Invalid user id');
	}

	const [user, assignments] = await Promise.all([
		getUser({ token }, id),
		listUserAssignments({ token }, id)
	]);

	if (!user) {
		throw error(404, 'User not found');
	}

	return { user, assignments };
};

export const actions: Actions = {
	assign: async ({ request, cookies, params, url }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const userId = Number(params.id);

		const data = await request.formData();
		const role = String(data.get('role') ?? '').trim();
		const scopeRaw = String(data.get('scope_node_id') ?? '').trim();
		const scopeNodeId = scopeRaw === '' ? null : Number(scopeRaw);

		if (!role) {
			return fail(400, { error: 'Role required.' });
		}

		if (scopeNodeId !== null && (!Number.isFinite(scopeNodeId) || scopeNodeId <= 0)) {
			return fail(400, { error: 'Scope must be a node id or left blank for global.' });
		}

		const result = await createAssignment({ token }, userId, role, scopeNodeId);
		if (!result.ok) {
			return fail(result.status, { error: result.message });
		}

		throw redirect(303, url.pathname);
	},
	revoke: async ({ request, cookies, url }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const assignmentId = Number(data.get('assignment_id'));

		if (!Number.isFinite(assignmentId) || assignmentId <= 0) {
			return fail(400, { error: 'Invalid assignment id.' });
		}

		const result = await deleteAssignment({ token }, assignmentId);
		if (!result.ok) {
			return fail(400, { error: result.message ?? 'Could not revoke.' });
		}

		throw redirect(303, url.pathname);
	}
};
