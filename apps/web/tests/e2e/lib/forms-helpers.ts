import { request, type APIRequestContext } from '@playwright/test';

const GATEWAY = process.env.E2E_GATEWAY_URL ?? 'http://localhost:8080';
const ADMIN_EMAIL = process.env.E2E_ADMIN_EMAIL ?? 'admin@local.test';
const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD ?? 'ChangeMe123!';

export async function adminToken(): Promise<string> {
	const ctx = await request.newContext();
	const res = await ctx.post(`${GATEWAY}/v1/auth/login`, {
		data: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD }
	});
	const body = await res.json();
	await ctx.dispose();
	return body.SessionToken as string;
}

export async function adminCtx(token?: string): Promise<APIRequestContext> {
	const t = token ?? (await adminToken());
	return await request.newContext({
		extraHTTPHeaders: { Authorization: `Bearer ${t}` }
	});
}

export type CreatedForm = { formId: number; questionIds: number[]; clientIds: string[] };

export async function createFormWithQuestions(
	ctx: APIRequestContext,
	title: string,
	questions: Array<{
		type: string;
		title: string;
		extra?: Record<string, unknown>;
		options?: Array<{ label: string; value: string }>;
		grading?: Record<string, unknown>;
		validation?: Record<string, unknown>;
		required?: boolean;
	}>
): Promise<CreatedForm> {
	const formRes = await ctx.post(`${GATEWAY}/v1/forms`, {
		data: { title, description: `Created at ${new Date().toISOString()}` }
	});
	const formBody = await formRes.json();
	const formId: number = formBody.form?.id ?? formBody.id;

	const questionIds: number[] = [];
	const clientIds: string[] = [];
	for (let i = 0; i < questions.length; i++) {
		const q = questions[i];
		const cid = `q_${i}_${Math.random().toString(36).slice(2, 10)}`;
		const qRes = await ctx.post(`${GATEWAY}/v1/forms/items/${formId}/questions`, {
			data: {
				client_id: cid,
				sort_order: (i + 1) * 10,
				title: q.title,
				type: q.type,
				required: q.required ?? false,
				options: q.options ?? undefined,
				extra: q.extra ?? undefined,
				grading: q.grading ?? undefined,
				validation: q.validation ?? undefined
			}
		});
		const qBody = await qRes.json();
		questionIds.push(qBody.id);
		clientIds.push(cid);
	}
	return { formId, questionIds, clientIds };
}

export async function publishForm(ctx: APIRequestContext, formId: number): Promise<{ versionId: number }> {
	const res = await ctx.post(`${GATEWAY}/v1/forms/items/${formId}/publish`);
	const body = await res.json();
	return { versionId: body.version.id };
}

export async function setLogicRules(
	ctx: APIRequestContext,
	formId: number,
	rules: Array<{
		id: string;
		target_question_client_id: string;
		operator: 'show_if' | 'hide_if';
		conditions: Array<{ source_question_client_id: string; op: string; value: string | number | string[] }>;
	}>
): Promise<void> {
	const formRes = await ctx.get(`${GATEWAY}/v1/forms/items/${formId}`);
	const f = (await formRes.json()).form;
	const settings = { ...(f.settings ?? {}), logic_rules: rules };
	await ctx.put(`${GATEWAY}/v1/forms/items/${formId}`, {
		data: { title: f.title, description: f.description, settings }
	});
}

export async function newAssignmentTokenForCampaign(
	ctx: APIRequestContext,
	campaignId: number
): Promise<string> {
	const stamp = Date.now().toString().slice(-6);
	const studentRes = await ctx.post(`${GATEWAY}/v1/people/students`, {
		data: { institutionId: 2, fullName: `E2E Renderer Student ${stamp}` }
	});
	const student = (await studentRes.json()).student;

	const assignRes = await ctx.post(`${GATEWAY}/v1/campaigns/${campaignId}/assign`, {
		data: { student_ids: [student.id], notify_by_email: false }
	});
	const access = (await assignRes.json()).created?.[0]?.access_token as string;
	if (!access) throw new Error('no access token from /v1/campaigns/:id/assign');
	return access;
}

export async function createCampaignForForm(
	ctx: APIRequestContext,
	formId: number,
	versionId: number,
	title: string
): Promise<number> {
	const res = await ctx.post(`${GATEWAY}/v1/campaigns`, {
		data: {
			title,
			scale_code: 'french_5level',
			form_id: formId,
			form_version_id: versionId,
			period_id: 10,
			scope_node_id: 2
		}
	});
	const body = await res.json();
	const id = body.id ?? body.campaign?.id;
	await ctx.post(`${GATEWAY}/v1/campaigns/${id}/open`);
	return id;
}

export async function buildPublishedFormWithToken(
	ctx: APIRequestContext,
	title: string,
	questions: Array<{
		type: string;
		title: string;
		extra?: Record<string, unknown>;
		options?: Array<{ label: string; value: string }>;
		required?: boolean;
	}>,
	logicRules?: Array<{
		id: string;
		target_question_client_id: string;
		operator: 'show_if' | 'hide_if';
		conditions: Array<{ source_question_client_id: string; op: string; value: string | number | string[] }>;
	}>
): Promise<{ formId: number; versionId: number; campaignId: number; token: string; clientIds: string[] }> {
	const created = await createFormWithQuestions(ctx, title, questions);
	if (logicRules && logicRules.length > 0) {
		const idMap = new Map<string, string>();
		for (let i = 0; i < created.clientIds.length; i++) idMap.set(`__Q${i}__`, created.clientIds[i]);
		const resolved = logicRules.map((r) => ({
			...r,
			target_question_client_id: idMap.get(r.target_question_client_id) ?? r.target_question_client_id,
			conditions: r.conditions.map((c) => ({
				...c,
				source_question_client_id:
					idMap.get(c.source_question_client_id) ?? c.source_question_client_id
			}))
		}));
		await setLogicRules(ctx, created.formId, resolved);
	}
	const { versionId } = await publishForm(ctx, created.formId);
	const campaignId = await createCampaignForForm(ctx, created.formId, versionId, title);
	const token = await newAssignmentTokenForCampaign(ctx, campaignId);
	return { formId: created.formId, versionId, campaignId, token, clientIds: created.clientIds };
}
