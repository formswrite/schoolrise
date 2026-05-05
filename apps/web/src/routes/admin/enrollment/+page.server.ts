import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { listPeriods } from '$lib/server/academics';
import { resolveDefaultScope } from '$lib/server/default-scope';
import {
	createEnrollment,
	dropEnrollment,
	getCoverage,
	listEnrollments,
	transferEnrollment
} from '$lib/server/enrollment';

function parseId(value: string | null): number | null {
	if (!value) return null;
	const n = Number(value);
	return Number.isFinite(n) && n > 0 ? n : null;
}

export const load: PageServerLoad = async ({ url, cookies, locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	const token = cookies.get('schoolrise_session') ?? '';

	const requestedScope = parseId(url.searchParams.get('scope'));
	const periods = await listPeriods({ token });

	let periodId = parseId(url.searchParams.get('period'));
	if (!periodId) {
		const cur = periods.find((p) => p.is_current);
		periodId = cur?.id ?? null;
	}

	const resolved = await resolveDefaultScope(token, locals.user, requestedScope);
	if (!resolved || !periodId) {
		return {
			scope: null,
			scopeNodeId: null,
			scopeOptions: resolved?.options ?? [],
			periodId,
			periods,
			enrollments: [],
			coverage: null
		};
	}

	const [enrollments, coverage] = await Promise.all([
		listEnrollments({ token }, resolved.scopeNodeId, periodId, true),
		getCoverage({ token }, resolved.scopeNodeId, periodId)
	]);

	return {
		scope: resolved.scope,
		scopeNodeId: resolved.scopeNodeId,
		scopeOptions: resolved.options,
		periodId,
		periods,
		enrollments,
		coverage
	};
};

export const actions: Actions = {
	enroll: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const body = {
			student_id: Number(data.get('student_id')),
			institution_id: Number(data.get('institution_id')),
			period_id: Number(data.get('period_id')),
			enrolled_on: String(data.get('enrolled_on') ?? '').trim(),
			note: String(data.get('note') ?? '')
		};
		if (!body.student_id || !body.institution_id || !body.period_id || !body.enrolled_on) {
			return fail(400, {
				error: 'student_id, institution_id, period_id, enrolled_on are required.'
			});
		}
		const res = await createEnrollment({ token }, body);
		if (!res.ok) return fail(res.status, { error: res.data?.message ?? 'Could not enroll.' });
		return { success: true };
	},

	transfer: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const body = {
			student_id: Number(data.get('student_id')),
			period_id: Number(data.get('period_id')),
			to_institution_id: Number(data.get('to_institution_id')),
			effective_on: String(data.get('effective_on') ?? ''),
			note: String(data.get('note') ?? '')
		};
		const res = await transferEnrollment({ token }, body);
		if (!res.ok) return fail(res.status, { error: res.data?.message ?? 'Could not transfer.' });
		return { success: true };
	},

	drop: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const body = {
			enrollment_id: Number(data.get('enrollment_id')),
			ended_on: String(data.get('ended_on') ?? ''),
			note: String(data.get('note') ?? '')
		};
		if (!body.enrollment_id) return fail(400, { error: 'enrollment_id required.' });
		const res = await dropEnrollment({ token }, body);
		if (!res.ok) return fail(res.status, { error: res.data?.message ?? 'Could not drop.' });
		return { success: true };
	}
};
