import type { Actions, PageServerLoad } from './$types';
import { error, fail, redirect } from '@sveltejs/kit';
import {
	assignStudents,
	getCampaign,
	listAssignments,
	listScores,
	setCampaignStatus
} from '$lib/server/campaigns';
import { listStudents } from '$lib/server/people';

export const load: PageServerLoad = async ({ params, cookies, locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	const id = Number(params.id);
	if (!Number.isFinite(id) || id <= 0) throw error(400, 'invalid id');
	const token = cookies.get('schoolrise_session') ?? '';

	const campaign = await getCampaign({ token }, id);
	if (!campaign) throw error(404, 'campaign not found');

	const [assignments, scores, eligibleStudents] = await Promise.all([
		listAssignments({ token }, id),
		listScores({ token }, id),
		listStudents({ token }, campaign.scope_node_id)
	]);

	return { campaign, assignments, scores, eligibleStudents };
};

export const actions: Actions = {
	open: async ({ params, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const res = await setCampaignStatus({ token }, Number(params.id), 'open');
		if (!res.ok) return fail(res.status, { error: 'Could not open campaign.' });
		return { success: true, message: 'Campaign opened.' };
	},
	close: async ({ params, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const res = await setCampaignStatus({ token }, Number(params.id), 'close');
		if (!res.ok) return fail(res.status, { error: 'Could not close campaign.' });
		return { success: true, message: 'Campaign closed.' };
	},
	assign: async ({ params, request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const studentIds = data.getAll('student_id').map((v) => Number(v)).filter((n) => Number.isFinite(n) && n > 0);
		const notify = data.get('notify_by_email') === 'on';
		if (studentIds.length === 0) {
			return fail(400, { error: 'Select at least one student.' });
		}
		const res = await assignStudents({ token }, Number(params.id), {
			student_ids: studentIds,
			notify_by_email: notify
		});
		if (!res.ok) return fail(res.status, { error: 'Assignment failed.' });
		const created = res.data.created?.length ?? 0;
		const existing = res.data.existing ?? 0;
		const sent = res.data.emails_sent ?? 0;
		return {
			success: true,
			toast: {
				type: 'success' as const,
				message: `Assigned ${created} new${existing ? ` (${existing} already assigned)` : ''}${notify ? ` · ${sent} emails sent` : ''}.`
			}
		};
	}
};
