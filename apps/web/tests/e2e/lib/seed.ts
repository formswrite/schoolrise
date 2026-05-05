import { request, type APIRequestContext } from '@playwright/test';

const GATEWAY = process.env.E2E_GATEWAY_URL ?? 'http://localhost:8080';
const ADMIN_EMAIL = process.env.E2E_ADMIN_EMAIL ?? 'admin@local.test';
const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD ?? 'ChangeMe123!';

export type Personas = {
	admin: { email: string; password: string };
	inspector: { email: string; password: string; userID: number; scopeNodeID: number };
	teacher: { email: string; password: string; userID: number; staffID: number; classID: number; scopeNodeID: number };
	student: { email: string };
	context: { regionID: number; schoolID: number; classID: number; campaignID: number; periodID: number; studentIDs: number[] };
};

async function adminCtx(): Promise<APIRequestContext> {
	const ctx = await request.newContext();
	const res = await ctx.post(`${GATEWAY}/v1/auth/login`, {
		data: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD }
	});
	const body = await res.json();
	await ctx.dispose();
	return await request.newContext({
		extraHTTPHeaders: { Authorization: `Bearer ${body.SessionToken}` }
	});
}

async function ok<T>(res: { ok(): boolean; status(): number; json(): Promise<T>; text(): Promise<string> }, label: string): Promise<T> {
	if (!res.ok()) {
		const body = await res.text();
		throw new Error(`${label} failed: ${res.status()} ${body}`);
	}
	return await res.json();
}

async function findOrCreate<T>(
	listFn: () => Promise<T[]>,
	matchFn: (item: T) => boolean,
	createFn: () => Promise<T>
): Promise<T> {
	const list = await listFn();
	const found = list.find(matchFn);
	if (found) return found;
	return await createFn();
}

export async function seedPersonas(): Promise<Personas> {
	const ctx = await adminCtx();

	const regionRes = await ctx.get(`${GATEWAY}/v1/tenancy/nodes`);
	const regionList = (await regionRes.json()).nodes ?? [];
	let region = regionList.find((n: { code: string }) => n.code === 'e2e-region');
	if (!region) {
		region = (await ok<{ node: { id: number } }>(
			await ctx.post(`${GATEWAY}/v1/tenancy/nodes`, { data: { code: 'e2e-region', label: 'E2E Region', level: 'region' } }),
			'create region'
		)).node;
	}
	const regionID = region.id;

	const schoolListRes = await ctx.get(`${GATEWAY}/v1/tenancy/nodes?parentId=${regionID}`);
	const schoolList = (await schoolListRes.json()).nodes ?? [];
	let school = schoolList.find((n: { code: string }) => n.code === 'e2e-school');
	if (!school) {
		school = (await ok<{ node: { id: number } }>(
			await ctx.post(`${GATEWAY}/v1/tenancy/nodes`, { data: { parentId: regionID, code: 'e2e-school', label: 'E2E School', level: 'institution' } }),
			'create school'
		)).node;
	}
	const schoolID = school.id;

	const periodsRes = await ctx.get(`${GATEWAY}/v1/academics/periods`);
	const periods = (await periodsRes.json()).periods ?? [];
	let period = periods.find((p: { code: string }) => p.code === 'e2e-2025-2026');
	if (!period) {
		period = await ok<{ id: number }>(
			await ctx.post(`${GATEWAY}/v1/academics/periods`, {
				data: { code: 'e2e-2025-2026', label: 'E2E 2025-2026', starts_on: '2025-09-01', ends_on: '2026-06-30', is_current: true }
			}),
			'create period'
		);
	}
	const periodID = period.id;

	const niveauxRes = await ctx.get(`${GATEWAY}/v1/academics/niveaux`);
	const niveaux = (await niveauxRes.json()).niveaux ?? [];
	let niveau = niveaux.find((n: { code: string }) => n.code === 'e2e-CE1');
	if (!niveau) {
		niveau = await ok<{ id: number }>(
			await ctx.post(`${GATEWAY}/v1/academics/niveaux`, { data: { code: 'e2e-CE1', label: 'E2E CE1', sort_order: 10 } }),
			'create niveau'
		);
	}

	const classesRes = await ctx.get(`${GATEWAY}/v1/academics/classes?institution_id=${schoolID}`);
	const classes = (await classesRes.json()).classes ?? [];
	let cls = classes.find((c: { code: string }) => c.code === 'E2E-CE1-A');
	if (!cls) {
		cls = await ok<{ id: number }>(
			await ctx.post(`${GATEWAY}/v1/academics/classes`, {
				data: { period_id: periodID, niveau_id: niveau.id, institution_id: schoolID, code: 'E2E-CE1-A', label: 'E2E CE1 morning', capacity: 30 }
			}),
			'create class'
		);
	}
	const classID = cls.id;

	const studentIDs: number[] = [];
	for (const name of ['E2E Alpha Student', 'E2E Beta Student', 'E2E Gamma Student', 'E2E Delta Student']) {
		const code = name.split(' ').slice(1, 2)[0].toLowerCase();
		const existingRes = await ctx.get(`${GATEWAY}/v1/people/students?institutionId=${schoolID}&limit=200`);
		const existing = (await existingRes.json()).students ?? [];
		const match = existing.find((s: { studentCode: string; person: { fullName: string } }) =>
			s.studentCode === code || s.person?.fullName === name
		);
		let id: number;
		if (match) {
			id = match.id;
		} else {
			const created = await ok<{ student: { id: number } }>(
				await ctx.post(`${GATEWAY}/v1/people/students`, {
					data: { institutionId: schoolID, fullName: name, studentCode: code, gender: 'female' }
				}),
				`create student ${name}`
			);
			id = created.student.id;
			await ctx.post(`${GATEWAY}/v1/academics/classes/${classID}/students`, { data: { student_id: id } });
			await ctx.post(`${GATEWAY}/v1/enrollment/enrollments`, {
				data: { student_id: id, institution_id: schoolID, period_id: periodID, enrolled_on: '2025-09-01' }
			});
		}
		studentIDs.push(id);
	}

	const formsRes = await ctx.get(`${GATEWAY}/v1/forms`);
	const forms = (await formsRes.json()).forms ?? [];
	let form = forms.find((f: { title: string }) => f.title === 'E2E Personas Form');
	let formVersionID: number;
	if (!form) {
		form = await ok<{ id: number }>(
			await ctx.post(`${GATEWAY}/v1/forms`, { data: { title: 'E2E Personas Form', description: 'seeded by playwright' } }),
			'create form'
		);
		await ctx.post(`${GATEWAY}/v1/forms/items/${form.id}/questions`, {
			data: { client_id: 'pq1', sort_order: 10, title: 'Pilot Q1', type: 'SHORT_ANSWER' }
		});
		const pub = await ok<{ version: { id: number } }>(
			await ctx.post(`${GATEWAY}/v1/forms/items/${form.id}/publish`),
			'publish form'
		);
		formVersionID = pub.version.id;
	} else {
		const versionsRes = await ctx.get(`${GATEWAY}/v1/forms/items/${form.id}/versions`);
		const versions = (await versionsRes.json()).versions ?? [];
		formVersionID = versions[0]?.id;
	}

	const campaignsRes = await ctx.get(`${GATEWAY}/v1/campaigns?scope_node_id=${schoolID}`);
	const campaigns = (await campaignsRes.json()).campaigns ?? [];
	let campaign = campaigns.find((c: { title: string }) => c.title === 'E2E Personas Campaign');
	if (!campaign) {
		campaign = await ok<{ id: number; status: string }>(
			await ctx.post(`${GATEWAY}/v1/campaigns`, {
				data: {
					title: 'E2E Personas Campaign',
					scale_code: 'french_5level',
					form_id: form.id,
					form_version_id: formVersionID,
					period_id: periodID,
					scope_node_id: schoolID
				}
			}),
			'create campaign'
		);
	}
	if (campaign.status !== 'open') {
		await ctx.post(`${GATEWAY}/v1/campaigns/${campaign.id}/open`);
	}

	const teacherEmail = 'e2e.teacher@local.test';
	const teacherPassword = 'TeachE2E1!';
	const usersRes = await ctx.get(`${GATEWAY}/v1/auth/users`);
	const users = (await usersRes.json()).users ?? [];
	let teacherUser = users.find((u: { email: string }) => u.email === teacherEmail);
	if (!teacherUser) {
		teacherUser = (await ok<{ user: { id: number } }>(
			await ctx.post(`${GATEWAY}/v1/auth/users`, {
				data: { email: teacherEmail, password: teacherPassword, fullName: 'E2E Teacher', role: 'teacher', mustChangePassword: false }
			}),
			'create teacher user'
		)).user;
	}

	const inspectorEmail = 'e2e.inspector@local.test';
	const inspectorPassword = 'Inspect3!';
	let inspectorUser = users.find((u: { email: string }) => u.email === inspectorEmail);
	if (!inspectorUser) {
		inspectorUser = (await ok<{ user: { id: number } }>(
			await ctx.post(`${GATEWAY}/v1/auth/users`, {
				data: { email: inspectorEmail, password: inspectorPassword, fullName: 'E2E Inspector', role: 'inspector', mustChangePassword: false }
			}),
			'create inspector user'
		)).user;
	}

	const teacherStaffRes = await ctx.get(`${GATEWAY}/v1/people/staff?scopeNodeId=${schoolID}&limit=200`);
	const teacherStaffList = (await teacherStaffRes.json()).staff ?? [];
	let teacherStaff = teacherStaffList.find((s: { person: { email: string } }) => s.person?.email === teacherEmail);
	if (!teacherStaff) {
		teacherStaff = (await ok<{ staff: { id: number } }>(
			await ctx.post(`${GATEWAY}/v1/people/staff`, {
				data: { scopeNodeId: schoolID, fullName: 'E2E Teacher', position: 'Teacher', email: teacherEmail }
			}),
			'create teacher staff'
		)).staff;
	}

	const teacherAssignRes = await ctx.get(`${GATEWAY}/v1/auth/users/${teacherUser.id}/assignments`);
	const teacherAssignments = (await teacherAssignRes.json()).assignments ?? [];
	if (!teacherAssignments.some((a: { role: string; scopeNodeId: number | null }) => a.role === 'teacher' && a.scopeNodeId === schoolID)) {
		await ctx.post(`${GATEWAY}/v1/auth/assignments`, {
			data: { userId: teacherUser.id, role: 'teacher', scopeNodeId: schoolID }
		});
	}

	const inspectorAssignRes = await ctx.get(`${GATEWAY}/v1/auth/users/${inspectorUser.id}/assignments`);
	const inspectorAssignments = (await inspectorAssignRes.json()).assignments ?? [];
	if (!inspectorAssignments.some((a: { role: string; scopeNodeId: number | null }) => a.role === 'inspector' && a.scopeNodeId === regionID)) {
		await ctx.post(`${GATEWAY}/v1/auth/assignments`, {
			data: { userId: inspectorUser.id, role: 'inspector', scopeNodeId: regionID }
		});
	}

	await ctx.post(`${GATEWAY}/v1/academics/classes/${classID}/staff`, {
		data: { staff_id: teacherStaff.id, role: 'teacher' }
	});

	await ctx.dispose();

	return {
		admin: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD },
		inspector: { email: inspectorEmail, password: inspectorPassword, userID: inspectorUser.id, scopeNodeID: regionID },
		teacher: { email: teacherEmail, password: teacherPassword, userID: teacherUser.id, staffID: teacherStaff.id, classID, scopeNodeID: schoolID },
		student: { email: 'e2e.student@local.test' },
		context: { regionID, schoolID, classID, campaignID: campaign.id, periodID, studentIDs }
	};
}

export async function newAssignmentToken(personas: Personas): Promise<string> {
	const ctx = await adminCtx();
	const stamp = Date.now().toString().slice(-7);
	const created = (await ok<{ student: { id: number } }>(
		await ctx.post(`${GATEWAY}/v1/people/students`, {
			data: { institutionId: personas.context.schoolID, fullName: `E2E Public ${stamp}`, studentCode: `pub${stamp}`, gender: 'female' }
		}),
		'create one-off student'
	)).student;
	await ctx.post(`${GATEWAY}/v1/enrollment/enrollments`, {
		data: { student_id: created.id, institution_id: personas.context.schoolID, period_id: personas.context.periodID, enrolled_on: '2025-09-01' }
	});
	const res = await ctx.post(`${GATEWAY}/v1/campaigns/${personas.context.campaignID}/assign`, {
		data: { student_ids: [created.id], notify_by_email: false }
	});
	const body = await res.json();
	await ctx.dispose();
	return body.created?.[0]?.access_token as string;
}
