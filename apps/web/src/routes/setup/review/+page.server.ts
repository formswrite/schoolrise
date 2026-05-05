import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { invalidateSetupStatusCache, setupFinalize } from '$lib/server/setup';

export const load: PageServerLoad = async ({ cookies }) => {
	if (!cookies.get('schoolrise_setup_session')) {
		throw redirect(303, '/setup/unlock');
	}

	return {};
};

export const actions: Actions = {
	default: async ({ cookies }) => {
		const session = cookies.get('schoolrise_setup_session') ?? '';

		const result = await setupFinalize(session);
		if (!result.ok) {
			return fail(result.status, { error: result.message ?? 'Finalize failed.' });
		}

		cookies.delete('schoolrise_setup_session', { path: '/setup' });
		invalidateSetupStatusCache();

		throw redirect(303, '/login?installed=1');
	}
};
