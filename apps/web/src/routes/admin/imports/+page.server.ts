import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { importStudents } from '$lib/server/imports';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	return {};
};

export const actions: Actions = {
	upload: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const institutionId = Number(data.get('institution_id'));
		const dryRun = data.get('dry_run') === 'on';
		const file = data.get('csv_file') as File | null;
		const pasted = String(data.get('csv_pasted') ?? '').trim();

		if (!Number.isFinite(institutionId) || institutionId <= 0) {
			return fail(400, { error: 'institution_id is required.' });
		}

		let csvData = '';
		if (file && file.size > 0) {
			csvData = await file.text();
		} else if (pasted) {
			csvData = pasted;
		} else {
			return fail(400, { error: 'Upload a CSV file or paste CSV text.' });
		}

		const res = await importStudents({ token }, {
			institution_id: institutionId,
			csv_data: csvData,
			dry_run: dryRun
		});

		if (!res.ok) {
			return fail(res.status, { error: res.message ?? 'Import failed.' });
		}
		return { job: res.job };
	}
};
