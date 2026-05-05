import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { setupSetLevels, type LevelInput } from '$lib/server/setup';

export const load: PageServerLoad = async ({ cookies }) => {
	if (!cookies.get('schoolrise_setup_session')) {
		throw redirect(303, '/setup/unlock');
	}

	return {};
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const session = cookies.get('schoolrise_setup_session') ?? '';

		const data = await request.formData();
		const codes = data.getAll('code').map((v) => String(v).trim());
		const labels = data.getAll('label').map((v) => String(v).trim());
		const parents = data.getAll('parent').map((v) => String(v).trim());

		const levels: LevelInput[] = [];

		for (let i = 0; i < codes.length; i++) {
			const code = codes[i];
			const label = labels[i];
			const parent = parents[i] ?? '';

			if (!code || !label) continue;

			levels.push({ code, label, parent, depth: i, sort: i });
		}

		if (levels.length === 0) {
			return fail(400, { error: 'Add at least one level.' });
		}

		const result = await setupSetLevels(session, levels);
		if (!result.ok) {
			return fail(result.status, { error: result.message ?? 'Could not save levels.' });
		}

		throw redirect(303, '/setup/schools');
	}
};
