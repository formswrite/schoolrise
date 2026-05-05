import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { setupImportSchools, setupSkipSchools } from '$lib/server/setup';

export const load: PageServerLoad = async ({ cookies }) => {
	if (!cookies.get('schoolrise_setup_session')) {
		throw redirect(303, '/setup/unlock');
	}

	return {};
};

export const actions: Actions = {
	import: async ({ request, cookies }) => {
		const session = cookies.get('schoolrise_setup_session') ?? '';

		const data = await request.formData();
		const csv = String(data.get('csv') ?? '').trim();

		if (!csv) {
			return fail(400, { error: 'CSV is empty.' });
		}

		const result = await setupImportSchools(session, csv);
		if (!result.ok) {
			return fail(result.status, { error: result.message ?? 'Import failed.' });
		}

		const data2 = result.data as { Imported: number; Errors: string[] | null } | undefined;
		const imported = data2?.Imported ?? 0;
		const errors = data2?.Errors ?? [];

		if (errors.length > 0) {
			return { imported, errors };
		}

		throw redirect(303, '/setup/integrations');
	},
	skip: async ({ cookies }) => {
		const session = cookies.get('schoolrise_setup_session') ?? '';

		await setupSkipSchools(session);

		throw redirect(303, '/setup/integrations');
	}
};
