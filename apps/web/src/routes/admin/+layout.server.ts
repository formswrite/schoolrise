import type { LayoutServerLoad } from './$types';
import { redirect, error } from '@sveltejs/kit';
import { summarizeRoles } from '$lib/server/encore';
import { canAccessAdminRoute, visibleNav } from '$lib/server/nav';

export const load: LayoutServerLoad = async ({ locals, url }) => {
	if (!locals.user) {
		throw redirect(303, '/login');
	}

	const summary = summarizeRoles(locals.user);

	if (!summary.isGlobalAdmin && !summary.isInspector) {
		if (summary.isTeacher) {
			throw redirect(303, '/teacher');
		}
		throw redirect(303, '/login');
	}

	if (!canAccessAdminRoute(url.pathname, summary)) {
		throw error(403, 'You do not have access to this section.');
	}

	return {
		user: locals.user,
		roleSummary: summary,
		navItems: visibleNav(summary)
	};
};
