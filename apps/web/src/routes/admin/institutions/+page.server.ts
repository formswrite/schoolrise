import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import {
	createNode,
	deleteNode,
	getNode,
	listLevels,
	listNodes,
	type Node
} from '$lib/server/tenancy';
import { listStudents, listStaff } from '$lib/server/people';
import { listClassesByInstitution } from '$lib/server/academics';

function parseParentId(value: string | null): number | null {
	if (!value) return null;
	const n = Number(value);
	return Number.isFinite(n) && n > 0 ? n : null;
}

export const load: PageServerLoad = async ({ url, cookies, locals }) => {
	if (!locals.user) {
		throw redirect(303, '/login');
	}

	const token = cookies.get('schoolrise_session') ?? '';
	const parentId = parseParentId(url.searchParams.get('parent'));

	const [levels, nodes] = await Promise.all([
		listLevels({ token }),
		listNodes({ token }, parentId)
	]);

	let parent: Node | null = null;
	const breadcrumbs: Node[] = [];

	if (parentId !== null) {
		parent = await getNode({ token }, parentId);
		let current = parent;
		while (current) {
			breadcrumbs.unshift(current);
			if (!current.parentId) break;
			current = await getNode({ token }, current.parentId);
		}
	}

	const childLevelCode = computeChildLevel(levels, parent);

	let leafSummary: { studentCount: number; staffCount: number; classCount: number } | null = null;
	if (parent && parent.level === 'institution' && nodes.length === 0) {
		const [students, staff, classes] = await Promise.all([
			listStudents({ token }, parentId!).catch(() => []),
			listStaff({ token }, parentId!).catch(() => []),
			listClassesByInstitution({ token }, parentId!).catch(() => [])
		]);
		leafSummary = {
			studentCount: students.length,
			staffCount: staff.length,
			classCount: classes.length
		};
	}

	return { levels, nodes, parent, parentId, breadcrumbs, childLevelCode, leafSummary };
};

function computeChildLevel(
	levels: { code: string; parentLevel: string }[],
	parent: Node | null
): string {
	if (parent === null) {
		const root = levels.find((l) => l.parentLevel === '');
		return root?.code ?? '';
	}

	const child = levels.find((l) => l.parentLevel === parent.level);
	return child?.code ?? '';
}

export const actions: Actions = {
	create: async ({ request, cookies, url }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const parentId = parseParentId(String(data.get('parent_id') ?? '') || null);
		const level = String(data.get('level') ?? '').trim();
		const code = String(data.get('code') ?? '').trim();
		const label = String(data.get('label') ?? '').trim();

		if (!level || !code || !label) {
			return fail(400, { error: 'Level, code and label required.' });
		}

		const result = await createNode({ token }, parentId, level, code, label);
		if (!result.ok) {
			return fail(result.status, { error: result.message });
		}

		const target = parentId !== null ? `${url.pathname}?parent=${parentId}` : url.pathname;
		throw redirect(303, target);
	},

	delete: async ({ request, cookies, url }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const id = Number(data.get('id'));

		if (!Number.isFinite(id) || id <= 0) {
			return fail(400, { error: 'Invalid node id.' });
		}

		const result = await deleteNode({ token }, id);
		if (!result.ok) {
			return fail(400, { error: result.message ?? 'Could not delete.' });
		}

		throw redirect(303, url.pathname + url.search);
	}
};
