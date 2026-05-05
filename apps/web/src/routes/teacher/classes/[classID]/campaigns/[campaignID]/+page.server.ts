import type { Actions, PageServerLoad } from './$types';
import { error, fail } from '@sveltejs/kit';
import { getGradingRoster, submitProctoredScores } from '$lib/server/teacher';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const classID = Number(params.classID);
	const campaignID = Number(params.campaignID);
	if (!Number.isFinite(classID) || !Number.isFinite(campaignID)) {
		throw error(400, 'invalid params');
	}
	const token = cookies.get('schoolrise_session') ?? '';
	const roster = await getGradingRoster({ token }, classID, campaignID);
	if (!roster) {
		throw error(404, 'campaign or class not found');
	}
	return { classID, campaignID, roster };
};

export const actions: Actions = {
	submit: async ({ params, request, cookies }) => {
		const classID = Number(params.classID);
		const campaignID = Number(params.campaignID);
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();

		const entries: { student_id: number; raw_score?: number; mode: 'proctored_score' }[] = [];
		for (const [k, v] of data.entries()) {
			if (!k.startsWith('score_')) continue;
			const sid = Number(k.slice('score_'.length));
			const raw = String(v ?? '').trim();
			if (raw === '') continue;
			const score = Number(raw);
			if (!Number.isFinite(score)) continue;
			entries.push({ student_id: sid, raw_score: Math.round(score), mode: 'proctored_score' });
		}

		if (entries.length === 0) {
			return fail(400, { error: 'No scores entered.' });
		}

		const res = await submitProctoredScores({ token }, classID, campaignID, entries);
		if (!res.ok) {
			return fail(res.status, { error: 'Submission failed.', detail: res.data });
		}
		const created = res.data.created ?? 0;
		const updated = res.data.updated ?? 0;
		const errors = res.data.errors ?? [];
		const total = created + updated;
		const msg = `Saved ${total} score${total === 1 ? '' : 's'} (${created} new, ${updated} updated)${errors.length ? ` · ${errors.length} error(s)` : ''}.`;
		const toastType: 'success' | 'warning' = errors.length ? 'warning' : 'success';
		return {
			success: true,
			created, updated, errors,
			toast: { type: toastType, message: msg }
		};
	}
};
