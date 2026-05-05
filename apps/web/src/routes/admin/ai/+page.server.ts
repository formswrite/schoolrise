import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { generateDistractors, getProviderStatus, listJobs, suggestItems } from '$lib/server/ai';

export const load: PageServerLoad = async ({ cookies, locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	const token = cookies.get('schoolrise_session') ?? '';
	const [provider, jobs] = await Promise.all([
		getProviderStatus({ token }),
		listJobs({ token }, 50)
	]);
	return { provider, jobs };
};

export const actions: Actions = {
	suggest: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const body = {
			topic: String(data.get('topic') ?? '').trim(),
			scale_code: String(data.get('scale_code') ?? 'french_5level'),
			niveau_label: String(data.get('niveau_label') ?? 'CE1'),
			count: Number(data.get('count') ?? 3)
		};
		if (!body.topic) return fail(400, { error: 'Topic required.' });
		const res = await suggestItems({ token }, body);
		if (!res.ok) return fail(500, { error: res.error });
		return {
			success: true,
			suggested: res.items,
			toast: { type: 'success' as const, message: `Generated ${res.items.length} items.` }
		};
	},
	distractors: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const body = {
			question_title: String(data.get('question_title') ?? '').trim(),
			correct_answer: String(data.get('correct_answer') ?? '').trim(),
			count: Number(data.get('count') ?? 3)
		};
		if (!body.question_title || !body.correct_answer) {
			return fail(400, { error: 'Question and correct answer are required.' });
		}
		const res = await generateDistractors({ token }, body);
		if (!res.ok) return fail(500, { error: res.error });
		return {
			success: true,
			distractors: res.distractors,
			toast: { type: 'success' as const, message: `Generated ${res.distractors.length} distractors.` }
		};
	}
};
