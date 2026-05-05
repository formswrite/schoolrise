import type { RequestHandler } from './$types';
import { redirect } from '@sveltejs/kit';
import { logoutRequest } from '$lib/server/encore';

export const POST: RequestHandler = async ({ cookies }) => {
	const token = cookies.get('schoolrise_session') ?? '';

	await logoutRequest(token);

	cookies.delete('schoolrise_session', { path: '/' });

	throw redirect(303, '/login');
};
