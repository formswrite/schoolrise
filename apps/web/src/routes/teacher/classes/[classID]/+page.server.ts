import type { PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';
import { listTeacherCampaigns } from '$lib/server/teacher';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const classID = Number(params.classID);
	if (!Number.isFinite(classID) || classID <= 0) {
		throw error(400, 'invalid class id');
	}
	const token = cookies.get('schoolrise_session') ?? '';
	const campaigns = await listTeacherCampaigns({ token }, classID);
	return { classID, campaigns };
};
