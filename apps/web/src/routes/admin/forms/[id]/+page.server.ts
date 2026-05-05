import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import {
	addQuestion,
	deleteQuestion,
	getForm,
	publishForm,
	updateForm,
	updateQuestion
} from '$lib/server/forms';
import { FieldType, getDefaultQuestion, type FieldTypeValue } from '$lib/forms/field-types';
import {
	newRuleId,
	type Condition,
	type ConditionOperator,
	type LogicRule,
	type RuleOperator
} from '$lib/forms/logic';

const VALID_OPERATORS: ReadonlySet<ConditionOperator> = new Set([
	'equals',
	'not_equals',
	'gt',
	'lt',
	'gte',
	'lte',
	'contains',
	'either'
]);

function rulesFromSettings(settings: Record<string, unknown>): LogicRule[] {
	const raw = (settings?.logic_rules ?? []) as unknown;
	if (!Array.isArray(raw)) return [];
	return raw.filter((r): r is LogicRule => {
		return (
			!!r &&
			typeof r === 'object' &&
			typeof (r as LogicRule).id === 'string' &&
			typeof (r as LogicRule).target_question_client_id === 'string' &&
			((r as LogicRule).operator === 'show_if' || (r as LogicRule).operator === 'hide_if') &&
			Array.isArray((r as LogicRule).conditions)
		);
	});
}

export const load: PageServerLoad = async ({ params, cookies, locals }) => {
	if (!locals.user) throw redirect(303, '/login');
	const token = cookies.get('schoolrise_session') ?? '';
	const id = Number(params.id);
	const data = await getForm({ token }, id);
	if (!data) throw redirect(303, '/admin/forms');
	return {
		form: data.form,
		questions: data.questions,
		logicRules: rulesFromSettings(data.form.settings)
	};
};

const VALID_TYPES = new Set(Object.values(FieldType));

function isValidType(t: string): t is FieldTypeValue {
	return VALID_TYPES.has(t as FieldTypeValue);
}

function generateClientId(): string {
	const chars = 'abcdefghijkmnpqrstuvwxyz23456789';
	let s = '';
	for (let i = 0; i < 16; i++) s += chars[Math.floor(Math.random() * chars.length)];
	return s;
}

function parseOptions(raw: string): Array<{ label: string; value: string }> {
	if (!raw.trim()) return [];
	return raw
		.split('\n')
		.map((line) => line.trim())
		.filter((v) => v.length > 0)
		.map((v) => ({ label: v, value: v }));
}

export const actions: Actions = {
	addQuestion: async ({ request, params, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const formData = await request.formData();
		const id = Number(params.id);
		const type = String(formData.get('type') ?? '').trim();
		if (!isValidType(type)) {
			return fail(400, { error: 'Invalid field type.' });
		}

		const existing = (await getForm({ token }, id))?.questions ?? [];
		const sortOrder = (existing.length + 1) * 10;
		const defaults = getDefaultQuestion(type, sortOrder);

		const res = await addQuestion(
			{ token },
			id,
			{
				client_id: generateClientId(),
				sort_order: sortOrder,
				title: defaults.title ?? '',
				type,
				required: defaults.required ?? false,
				options: (defaults.options ?? []) as never,
				scale_min: defaults.scale_min,
				scale_max: defaults.scale_max
			}
		);
		if (!res.ok) return fail(res.status, { error: res.data?.message ?? 'Could not add question.' });
		return { success: true };
	},

	updateQuestion: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const formData = await request.formData();
		const qid = Number(formData.get('question_id'));
		if (!qid) return fail(400, { error: 'Invalid question id.' });

		const type = String(formData.get('type') ?? '').trim();
		if (type && !isValidType(type)) {
			return fail(400, { error: 'Invalid field type.' });
		}

		const body: Record<string, unknown> = {
			title: String(formData.get('title') ?? '').trim(),
			description: String(formData.get('description') ?? '').trim(),
			required: formData.get('required') === 'on',
			sort_order: Number(formData.get('sort_order') ?? 0)
		};
		if (type) body.type = type;

		const optionsRaw = formData.get('options');
		if (optionsRaw !== null) {
			body.options = parseOptions(String(optionsRaw));
		}
		const scaleMin = formData.get('scale_min');
		const scaleMax = formData.get('scale_max');
		if (scaleMin !== null && String(scaleMin).length > 0) body.scale_min = Number(scaleMin);
		if (scaleMax !== null && String(scaleMax).length > 0) body.scale_max = Number(scaleMax);

		const extra: Record<string, unknown> = {};
		const extraLatex = formData.get('extra_latex');
		if (extraLatex !== null) extra.latex = String(extraLatex);
		const extraLanguage = formData.get('extra_language');
		if (extraLanguage !== null) extra.language = String(extraLanguage);
		const extraTemplate = formData.get('extra_template');
		if (extraTemplate !== null) extra.template = String(extraTemplate);
		const extraFields = formData.get('extra_fields');
		if (extraFields !== null) {
			extra.fields = String(extraFields)
				.split('\n')
				.map((s) => s.trim())
				.filter(Boolean);
		}
		const extraRows = formData.get('extra_rows');
		if (extraRows !== null) {
			extra.rows = String(extraRows)
				.split('\n')
				.map((s) => s.trim())
				.filter(Boolean);
		}
		const extraColumns = formData.get('extra_columns');
		if (extraColumns !== null) {
			extra.columns = String(extraColumns)
				.split('\n')
				.map((s) => s.trim())
				.filter(Boolean);
		}
		const extraPairs = formData.get('extra_pairs');
		if (extraPairs !== null) {
			extra.pairs = String(extraPairs)
				.split('\n')
				.map((line) => {
					const m = line.split(/\s*→\s*|\s*->\s*/);
					return { left: (m[0] ?? '').trim(), right: (m[1] ?? '').trim() };
				})
				.filter((p) => p.left || p.right);
		}
		if (Object.keys(extra).length > 0) body.extra = extra;

		const validation: Record<string, unknown> = {};
		for (const k of ['min_length', 'max_length', 'min', 'max']) {
			const v = formData.get(`val_${k}`);
			if (v !== null && String(v).length > 0) validation[k] = Number(v);
		}
		const valPattern = formData.get('val_pattern');
		if (valPattern !== null && String(valPattern).length > 0) validation.pattern = String(valPattern);
		if (Object.keys(validation).length > 0) body.validation = validation;

		const grading: Record<string, unknown> = {};
		const gPoints = formData.get('grading_points_max');
		if (gPoints !== null && String(gPoints).length > 0) grading.points_max = Number(gPoints);
		const gCorrect = formData.get('grading_correct_value');
		if (gCorrect !== null && String(gCorrect).length > 0) grading.correct_value = String(gCorrect);
		const gAnswers = formData.get('grading_answers');
		if (gAnswers !== null) {
			grading.answers = String(gAnswers)
				.split('\n')
				.map((s) => s.trim())
				.filter(Boolean);
		}
		const gRubric = formData.get('grading_rubric_url');
		if (gRubric !== null && String(gRubric).length > 0) grading.rubric_url = String(gRubric);
		if (Object.keys(grading).length > 0) body.grading = grading;

		const res = await updateQuestion({ token }, qid, body);
		if (!res.ok) return fail(res.status, { error: res.data?.message ?? 'Could not save question.' });
		return { success: true };
	},

	reorderQuestion: async ({ request, params, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const formData = await request.formData();
		const id = Number(params.id);
		const orderedRaw = String(formData.get('ordered_ids') ?? '');
		const orderedIds = orderedRaw
			.split(',')
			.map((s) => Number(s.trim()))
			.filter((n) => Number.isFinite(n) && n > 0);
		if (orderedIds.length === 0) return fail(400, { error: 'No order provided.' });

		const current = (await getForm({ token }, id))?.questions ?? [];
		const byId = new Map(current.map((q) => [q.id, q]));

		for (let i = 0; i < orderedIds.length; i++) {
			const q = byId.get(orderedIds[i]);
			if (!q) continue;
			const res = await updateQuestion({ token }, q.id, {
				title: q.title,
				description: q.description,
				type: q.type,
				required: q.required,
				sort_order: (i + 1) * 10,
				options: q.options as never,
				scale_min: q.scale_min,
				scale_max: q.scale_max,
				scale_labels: q.scale_labels,
				validation: q.validation,
				grading: q.grading,
				extra: q.extra
			});
			if (!res.ok) return fail(res.status, { error: 'Could not reorder.' });
		}
		return { success: true };
	},

	deleteQuestion: async ({ request, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const formData = await request.formData();
		const qid = Number(formData.get('question_id'));
		if (!qid) return fail(400, { error: 'Invalid question id.' });
		const res = await deleteQuestion({ token }, qid);
		if (!res.ok) return fail(res.status, { error: 'Could not delete.' });
		return { success: true };
	},

	createLogicRule: async ({ request, params, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const formData = await request.formData();
		const id = Number(params.id);

		const target = String(formData.get('target_client_id') ?? '').trim();
		const ruleOperator = String(formData.get('rule_operator') ?? 'show_if') as RuleOperator;
		const sourceCID = String(formData.get('source_client_id') ?? '').trim();
		const condOp = String(formData.get('cond_op') ?? 'equals') as ConditionOperator;
		const condValue = String(formData.get('cond_value') ?? '').trim();

		if (!target || !sourceCID) return fail(400, { error: 'Need target and source.' });
		if (ruleOperator !== 'show_if' && ruleOperator !== 'hide_if') {
			return fail(400, { error: 'Invalid rule operator.' });
		}
		if (!VALID_OPERATORS.has(condOp)) return fail(400, { error: 'Invalid condition operator.' });
		if (target === sourceCID) {
			return fail(400, { error: 'A question cannot reference itself.' });
		}

		const current = await getForm({ token }, id);
		if (!current) return fail(404, { error: 'Form not found.' });
		const rules = rulesFromSettings(current.form.settings);

		const condition: Condition = { source_question_client_id: sourceCID, op: condOp, value: condValue };
		const newRule: LogicRule = {
			id: newRuleId(),
			target_question_client_id: target,
			operator: ruleOperator,
			conditions: [condition]
		};
		rules.push(newRule);

		const settings = { ...current.form.settings, logic_rules: rules };
		const res = await updateForm({ token }, id, {
			title: current.form.title,
			description: current.form.description,
			settings
		});
		if (!res.ok) return fail(res.status, { error: res.data?.message ?? 'Could not save rule.' });
		return { success: true };
	},

	deleteLogicRule: async ({ request, params, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const formData = await request.formData();
		const id = Number(params.id);
		const ruleId = String(formData.get('rule_id') ?? '');
		if (!ruleId) return fail(400, { error: 'Missing rule id.' });

		const current = await getForm({ token }, id);
		if (!current) return fail(404, { error: 'Form not found.' });
		const rules = rulesFromSettings(current.form.settings).filter((r) => r.id !== ruleId);

		const settings = { ...current.form.settings, logic_rules: rules };
		const res = await updateForm({ token }, id, {
			title: current.form.title,
			description: current.form.description,
			settings
		});
		if (!res.ok) return fail(res.status, { error: 'Could not delete rule.' });
		return { success: true };
	},

	publish: async ({ params, cookies }) => {
		const token = cookies.get('schoolrise_session') ?? '';
		const id = Number(params.id);
		const res = await publishForm({ token }, id);
		if (!res.ok) return fail(res.status, { error: res.data?.message ?? 'Could not publish.' });
		return { success: true, version: res.data.version };
	}
};
