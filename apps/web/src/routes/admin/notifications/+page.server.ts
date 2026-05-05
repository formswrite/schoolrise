import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import {
	getProviderStatus,
	listEmails,
	processOutbox,
	sendTestEmail
} from '$lib/server/notifications';

export const load: PageServerLoad = async ({ cookies, locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	const token = cookies.get('schoolrise_session') ?? '';
	const [emails, provider] = await Promise.all([
		listEmails({ token }, 50),
		getProviderStatus({ token })
	]);
	return { emails, provider };
};

export const actions: Actions = {
	test: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const to = String(data.get('to') ?? '').trim();
		if (!to) return fail(400, { error: 'recipient email required.' });
		const res = await sendTestEmail({ token }, { to });
		if (!res.ok) return fail(res.status, { error: res.data?.message ?? 'Send failed.' });
		return { success: true, sentTo: to };
	},
	process: async ({ cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const res = await processOutbox({ token });
		if (!res.ok) return fail(res.status, { error: 'Process failed.' });
		return { success: true };
	}
};
