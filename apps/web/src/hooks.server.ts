import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';
import { fetchSession } from '$lib/server/encore';
import { fetchSetupStatus } from '$lib/server/setup';

export const handle: Handle = async ({ event, resolve }) => {
	const status = await fetchSetupStatus();
	event.locals.setupComplete = status.setupComplete;

	const path = event.url.pathname;
	const isSetupRoute = path.startsWith('/setup');

	if (!status.setupComplete && !isSetupRoute) {
		throw redirect(303, '/setup/welcome');
	}

	if (status.setupComplete && isSetupRoute) {
		throw redirect(303, '/login');
	}

	if (status.setupComplete) {
		const token = event.cookies.get('schoolrise_session') ?? '';
		event.locals.user = await fetchSession(token);
	} else {
		event.locals.user = null;
	}

	return resolve(event);
};
