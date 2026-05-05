import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { setupSaveSystem } from '$lib/server/setup';

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
		const instanceName = String(data.get('instance_name') ?? '').trim();
		const defaultLocale = String(data.get('default_locale') ?? 'en').trim();
		const baseURL = String(data.get('base_url') ?? '').trim();
		const timeZone = String(data.get('time_zone') ?? 'UTC').trim();

		if (!instanceName || !baseURL) {
			return fail(400, { instanceName, defaultLocale, baseURL, timeZone, error: 'Instance name and base URL required.' });
		}

		const result = await setupSaveSystem(session, instanceName, defaultLocale, baseURL, timeZone);
		if (!result.ok) {
			return fail(result.status, { instanceName, defaultLocale, baseURL, timeZone, error: result.message ?? 'Could not save settings.' });
		}

		throw redirect(303, '/setup/levels');
	}
};
