import type { Actions, PageServerLoad } from './$types';
import { error, fail, redirect } from '@sveltejs/kit';
import {
	addStaffToClass,
	addStudentToClass,
	getClass,
	getClassRoster,
	listNiveaux,
	listPeriods,
	removeStaffFromClass,
	removeStudentFromClass
} from '$lib/server/academics';
import { listStaff, listStudents } from '$lib/server/people';
import { getNode } from '$lib/server/tenancy';

export const load: PageServerLoad = async ({ params, cookies, locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	const id = Number(params.id);
	if (!Number.isFinite(id) || id <= 0) throw error(400, 'invalid id');
	const token = cookies.get('schoolrise_session') ?? '';

	const cls = await getClass({ token }, id);
	if (!cls) throw error(404, 'class not found');

	const [roster, allStudents, allStaff, periods, niveaux, institution] = await Promise.all([
		getClassRoster({ token }, id),
		listStudents({ token }, cls.institution_id),
		listStaff({ token }, cls.institution_id),
		listPeriods({ token }),
		listNiveaux({ token }),
		getNode({ token }, cls.institution_id)
	]);

	const studentIDs = roster?.student_ids ?? [];
	const staffMembers = roster?.staff ?? [];

	const inClass = allStudents.filter((s) => studentIDs.includes(s.id));
	const eligible = allStudents.filter((s) => !studentIDs.includes(s.id));

	const onStaff = allStaff.filter((s) => staffMembers.some((m) => m.staff_id === s.id));
	const eligibleStaff = allStaff.filter((s) => !staffMembers.some((m) => m.staff_id === s.id));

	const period = periods.find((p) => p.id === cls.period_id);
	const niveau = niveaux.find((n) => n.id === cls.niveau_id);

	return {
		cls,
		institution,
		period,
		niveau,
		inClass,
		eligible,
		onStaff,
		eligibleStaff,
		staffRoles: staffMembers
	};
};

export const actions: Actions = {
	addStudent: async ({ params, request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const studentID = Number(data.get('student_id'));
		if (!studentID) return fail(400, { error: 'Pick a student.' });
		const res = await addStudentToClass({ token }, Number(params.id), studentID);
		if (!res.ok) return fail(res.status, { error: 'Could not add student.' });
		return { success: true, message: 'Student added to class.' };
	},

	removeStudent: async ({ params, request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const studentID = Number(data.get('student_id'));
		const res = await removeStudentFromClass({ token }, Number(params.id), studentID);
		if (!res.ok) return fail(res.status, { error: 'Could not remove student.' });
		return { success: true, message: 'Student removed.' };
	},

	addStaff: async ({ params, request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const staffID = Number(data.get('staff_id'));
		const role = String(data.get('role') ?? 'teacher').trim() || 'teacher';
		if (!staffID) return fail(400, { error: 'Pick a staff member.' });
		const res = await addStaffToClass({ token }, Number(params.id), staffID, role);
		if (!res.ok) return fail(res.status, { error: 'Could not add staff.' });
		return { success: true, message: 'Staff added to class.' };
	},

	removeStaff: async ({ params, request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const data = await request.formData();
		const staffID = Number(data.get('staff_id'));
		const role = String(data.get('role') ?? 'teacher');
		const res = await removeStaffFromClass({ token }, Number(params.id), staffID, role);
		if (!res.ok) return fail(res.status, { error: 'Could not remove staff.' });
		return { success: true, message: 'Staff removed.' };
	}
};
