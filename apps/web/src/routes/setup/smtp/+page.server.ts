import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { setupSaveSMTP, setupSkipSMTP } from '$lib/server/setup';

export const load: PageServerLoad = async ({ cookies }) => {
	if (!cookies.get('schoolrise_setup_session')) {
		throw redirect(303, '/setup/unlock');
	}

	return {};
};

export const actions: Actions = {
	save: async ({ request, cookies }) => {
		const session = cookies.get('schoolrise_setup_session') ?? '';

		const data = await request.formData();
		const host = String(data.get('host') ?? '').trim();
		const port = Number(data.get('port') ?? 587);
		const username = String(data.get('username') ?? '');
		const password = String(data.get('password') ?? '');
		const useTLS = data.get('use_tls') === 'on';
		const fromAddress = String(data.get('from_address') ?? '').trim();

		if (!host || !fromAddress || port <= 0) {
			return fail(400, { error: 'Host, port and from address required.' });
		}

		const result = await setupSaveSMTP(
			session,
			host,
			port,
			username,
			password,
			useTLS,
			fromAddress
		);
		if (!result.ok) {
			return fail(result.status, { error: result.message ?? 'Could not save SMTP config.' });
		}

		throw redirect(303, '/setup/review');
	},
	skip: async ({ cookies }) => {
		const session = cookies.get('schoolrise_setup_session') ?? '';

		await setupSkipSMTP(session);

		throw redirect(303, '/setup/review');
	}
};
