import type { PageServerLoad } from './$types';
import { listTeacherClasses } from '$lib/server/teacher';

export const load: PageServerLoad = async ({ cookies }) => {
	const token = cookies.get('schoolrise_session') ?? '';
	const { classes } = await listTeacherClasses({ token });
	return { classes };
};
