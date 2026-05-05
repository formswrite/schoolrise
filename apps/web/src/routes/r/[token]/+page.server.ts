import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { getPublicVersion, lookupAssignment, submitResponse } from '$lib/server/forms';

export const load: PageServerLoad = async ({ params }) => {
	const lookup = await lookupAssignment(params.token);
	if (!lookup) {
		return { found: false as const, token: params.token };
	}

	if (lookup.assignment.submitted_at) {
		return { found: true as const, alreadySubmitted: true, lookup, version: null };
	}

	const version = await getPublicVersion(lookup.form_version_id);
	return { found: true as const, alreadySubmitted: false, lookup, version };
};

export const actions: Actions = {
	submit: async ({ request, params }) => {
		const data = await request.formData();
		const payload: Record<string, unknown> = {};
		for (const [k, v] of data.entries()) {
			if (k.startsWith('q_')) {
				payload[k.slice(2)] = v;
			}
		}

		const totalQuestions = Number(data.get('_total_questions') ?? 0);
		let answered = 0;
		for (const k of Object.keys(payload)) {
			const v = payload[k];
			if (typeof v === 'string' && v.trim() !== '') answered++;
		}
		const rawScore = totalQuestions > 0
			? Math.min(100, Math.round((answered / totalQuestions) * 100))
			: 0;

		const res = await submitResponse({
			access_token: params.token,
			payload,
			raw_score: rawScore
		});
		if (!res.ok) {
			return fail(res.status, { error: res.data?.message ?? 'Could not submit.' });
		}
		throw redirect(303, `/r/${params.token}/done`);
	}
};
